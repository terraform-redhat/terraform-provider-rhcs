// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/smithy-go"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

const (
	testManagedVPCTagKey   = "kubernetes.io/cluster/unmanaged"
	testManagedVPCTagValue = "true"

	redHatManagedTagKey   = "red-hat-managed"
	redHatManagedTagValue = "true"

	clusterIDTagKey         = "api.openshift.com/id"
	clusterOwnedTagTemplate = "kubernetes.io/cluster/%s"

	leakedSecurityGroupMaxAttempts   = 8
	leakedSecurityGroupRetryInterval = 15 * time.Second
)

var leakedSecurityGroupNamePattern = regexp.MustCompile(`^([a-z0-9]{32})-(vpce-private-router|default-sg)$`)

type ec2SecurityGroupAPI interface {
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DeleteSecurityGroup(ctx context.Context, params *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
}

type securityGroupCleanupOptions struct {
	maxAttempts   int
	retryInterval time.Duration
	sleep         func(time.Duration)
}

func defaultSecurityGroupCleanupOptions() securityGroupCleanupOptions {
	return securityGroupCleanupOptions{
		maxAttempts:   leakedSecurityGroupMaxAttempts,
		retryInterval: leakedSecurityGroupRetryInterval,
		sleep:         time.Sleep,
	}
}

// TODO: Remove this once this bug is fixed (https://issues.redhat.com/browse/OCPBUGS-74960)
func DeleteExtraSecurityGroups(region string, vpcId string) error {
	if strings.TrimSpace(vpcId) == "" {
		Logger.Warn("Skipping extra security groups cleanup because VPC ID is empty")
		return nil
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}
	client := ec2.NewFromConfig(cfg)
	return deleteExtraSecurityGroupsWithClient(ctx, client, vpcId, defaultSecurityGroupCleanupOptions())
}

func deleteExtraSecurityGroupsWithClient(ctx context.Context, client ec2SecurityGroupAPI, vpcId string, opts securityGroupCleanupOptions) error {
	opts = normalizeSecurityGroupCleanupOptions(opts)

	vpc, err := retrieveVPC(ctx, client, vpcId)
	if err != nil {
		return err
	}
	if vpc == nil {
		Logger.Warnf("Skipping extra security groups cleanup because VPC '%s' was not found", vpcId)
		return nil
	}
	if !isTestManagedVPC(vpc) {
		Logger.Warnf("Skipping extra security groups cleanup for VPC '%s' because it is not tagged as a test-managed VPC", vpcId)
		return nil
	}

	var securityGroups []types.SecurityGroup
	paginator := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		securityGroups = append(securityGroups, page.SecurityGroups...)
	}

	var errs []error
	for _, securityGroup := range securityGroups {
		if !isKnownLeakedSecurityGroup(securityGroup) {
			continue
		}
		if err := deleteSecurityGroupWithRetry(ctx, client, vpcId, securityGroup, opts); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func normalizeSecurityGroupCleanupOptions(opts securityGroupCleanupOptions) securityGroupCleanupOptions {
	if opts.maxAttempts <= 0 {
		opts.maxAttempts = leakedSecurityGroupMaxAttempts
	}
	if opts.retryInterval <= 0 {
		opts.retryInterval = leakedSecurityGroupRetryInterval
	}
	if opts.sleep == nil {
		opts.sleep = time.Sleep
	}
	return opts
}

func retrieveVPC(ctx context.Context, client ec2SecurityGroupAPI, vpcId string) (*types.Vpc, error) {
	result, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcId},
	})
	if err != nil {
		if hasAWSErrorCode(err, "InvalidVpcID.NotFound") {
			return nil, nil
		}
		return nil, err
	}
	if result == nil || len(result.Vpcs) == 0 {
		return nil, nil
	}
	return &result.Vpcs[0], nil
}

func isTestManagedVPC(vpc *types.Vpc) bool {
	if vpc == nil {
		return false
	}
	tags := tagsToMap(vpc.Tags)
	return strings.EqualFold(tags[testManagedVPCTagKey], testManagedVPCTagValue)
}

func isKnownLeakedSecurityGroup(securityGroup types.SecurityGroup) bool {
	if securityGroup.GroupName == nil || securityGroup.GroupId == nil {
		return false
	}
	if aws.ToString(securityGroup.GroupName) == "default" {
		return false
	}

	matches := leakedSecurityGroupNamePattern.FindStringSubmatch(aws.ToString(securityGroup.GroupName))
	if len(matches) != 3 {
		return false
	}
	infraID := matches[1]
	tags := tagsToMap(securityGroup.Tags)

	return hasExpectedClusterOwnership(tags, infraID)
}

func hasExpectedClusterOwnership(tags map[string]string, infraID string) bool {
	if tags[clusterIDTagKey] == infraID && strings.EqualFold(tags[redHatManagedTagKey], redHatManagedTagValue) {
		return true
	}
	ownedTagKey := fmt.Sprintf(clusterOwnedTagTemplate, infraID)
	return strings.EqualFold(tags[ownedTagKey], "owned")
}

func tagsToMap(tags []types.Tag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, tag := range tags {
		if tag.Key == nil || tag.Value == nil {
			continue
		}
		m[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return m
}

func deleteSecurityGroupWithRetry(ctx context.Context, client ec2SecurityGroupAPI, vpcId string, securityGroup types.SecurityGroup, opts securityGroupCleanupOptions) error {
	groupId := aws.ToString(securityGroup.GroupId)
	groupName := aws.ToString(securityGroup.GroupName)

	var lastErr error
	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		attached, err := isSecurityGroupAttached(ctx, client, vpcId, groupId)
		if err != nil {
			return fmt.Errorf("failed to inspect security group '%s' attachments: %w", groupId, err)
		}
		if attached {
			lastErr = fmt.Errorf("security group '%s' is still attached to network interfaces", groupId)
		} else {
			_, err = client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
				GroupId: securityGroup.GroupId,
			})
			if err == nil || hasAWSErrorCode(err, "InvalidGroup.NotFound") {
				return nil
			}
			if !hasAWSErrorCode(err, "DependencyViolation") {
				return fmt.Errorf("failed to delete leaked security group '%s' (%s): %w", groupName, groupId, err)
			}
			lastErr = err
		}

		if attempt < opts.maxAttempts {
			opts.sleep(opts.retryInterval)
		}
	}

	return fmt.Errorf(
		"failed to delete leaked security group '%s' (%s) after %d attempts: %w",
		groupName,
		groupId,
		opts.maxAttempts,
		lastErr,
	)
}

func isSecurityGroupAttached(ctx context.Context, client ec2SecurityGroupAPI, vpcId string, securityGroupId string) (bool, error) {
	output, err := client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
			{
				Name:   aws.String("group-id"),
				Values: []string{securityGroupId},
			},
		},
	})
	if err != nil {
		return false, err
	}
	return output != nil && len(output.NetworkInterfaces) > 0, nil
}

func hasAWSErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == code
	}
	return false
}

// DeleteNonDefaultSecurityGroups deletes every security group in vpcID except
// the VPC's built-in "default" group, which AWS does not allow deleting.
// This is a broader cleanup than DeleteExtraSecurityGroups: it removes any SG
// that would block VPC teardown regardless of its name or tags.
func DeleteNonDefaultSecurityGroups(region, vpcID string) error {
	if strings.TrimSpace(vpcID) == "" {
		Logger.Warn("Skipping security groups cleanup because VPC ID is empty")
		return nil
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}
	return deleteNonDefaultSecurityGroupsWithClient(
		ctx, ec2.NewFromConfig(cfg), vpcID, defaultSecurityGroupCleanupOptions(),
	)
}

func deleteNonDefaultSecurityGroupsWithClient(
	ctx context.Context, client ec2SecurityGroupAPI, vpcID string, opts securityGroupCleanupOptions,
) error {
	opts = normalizeSecurityGroupCleanupOptions(opts)

	var sgs []types.SecurityGroup
	paginator := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("describing security groups in VPC %s: %w", vpcID, err)
		}
		sgs = append(sgs, page.SecurityGroups...)
	}

	var errs []error
	for _, sg := range sgs {
		if aws.ToString(sg.GroupName) == "default" {
			continue
		}
		Logger.Infof("Deleting security group %s (%s) in VPC %s", aws.ToString(sg.GroupName), aws.ToString(sg.GroupId), vpcID)
		if err := deleteSecurityGroupWithRetry(ctx, client, vpcID, sg, opts); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type classicELBAPI interface {
	DescribeLoadBalancers(
		ctx context.Context,
		params *elasticloadbalancing.DescribeLoadBalancersInput,
		optFns ...func(*elasticloadbalancing.Options),
	) (*elasticloadbalancing.DescribeLoadBalancersOutput, error)
	DeleteLoadBalancer(
		ctx context.Context,
		params *elasticloadbalancing.DeleteLoadBalancerInput,
		optFns ...func(*elasticloadbalancing.Options),
	) (*elasticloadbalancing.DeleteLoadBalancerOutput, error)
}

// DeleteClassicLoadBalancers deletes all classic (v1) ELBs whose VPC matches
// vpcID. This is needed during VPC teardown because the cluster may have
// created ELBs for LoadBalancer-type services that Terraform doesn't track.
func DeleteClassicLoadBalancers(region, vpcID string) error {
	if strings.TrimSpace(vpcID) == "" {
		Logger.Warn("Skipping classic ELB cleanup because VPC ID is empty")
		return nil
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}
	return deleteClassicLoadBalancersWithClient(ctx, elasticloadbalancing.NewFromConfig(cfg), vpcID)
}

// route53RecordsAPI is the subset of the Route53 client used to purge records
// from a hosted zone.
type route53RecordsAPI interface {
	ListResourceRecordSets(
		ctx context.Context,
		params *route53.ListResourceRecordSetsInput,
		optFns ...func(*route53.Options),
	) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(
		ctx context.Context,
		params *route53.ChangeResourceRecordSetsInput,
		optFns ...func(*route53.Options),
	) (*route53.ChangeResourceRecordSetsOutput, error)
}

// PurgeHostedZoneRecords deletes every non-NS/non-SOA record from the given
// Route53 hosted zone. The Hyperfleet cluster operator writes CNAME records
// (api.*, *.apps.*) into the cluster's private `<name>.hypershift.local` zone;
// terraform destroy of the zone fails with HostedZoneNotEmpty while those
// records remain, so the e2e VPC teardown purges them first. NS and SOA records
// are left in place — they are created and removed by Route53 with the zone
// itself and cannot be deleted independently.
func PurgeHostedZoneRecords(region, hostedZoneID string) error {
	if strings.TrimSpace(hostedZoneID) == "" {
		Logger.Warn("Skipping hosted zone record purge because hosted zone ID is empty")
		return nil
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}
	return purgeHostedZoneRecordsWithClient(ctx, route53.NewFromConfig(cfg), hostedZoneID)
}

func purgeHostedZoneRecordsWithClient(ctx context.Context, client route53RecordsAPI, hostedZoneID string) error {
	out, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
	})
	if err != nil {
		return fmt.Errorf("listing record sets for zone %s: %w", hostedZoneID, err)
	}

	var changes []route53types.Change
	for i := range out.ResourceRecordSets {
		rrs := out.ResourceRecordSets[i]
		if rrs.Type == route53types.RRTypeNs || rrs.Type == route53types.RRTypeSoa {
			continue
		}
		changes = append(changes, route53types.Change{
			Action:            route53types.ChangeActionDelete,
			ResourceRecordSet: &rrs,
		})
	}
	if len(changes) == 0 {
		return nil
	}

	Logger.Infof("Purging %d record(s) from hosted zone %s", len(changes), hostedZoneID)
	_, err = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		ChangeBatch:  &route53types.ChangeBatch{Changes: changes},
	})
	if err != nil {
		return fmt.Errorf("deleting records from zone %s: %w", hostedZoneID, err)
	}
	return nil
}

func deleteClassicLoadBalancersWithClient(ctx context.Context, client classicELBAPI, vpcID string) error {
	var names []string
	paginator := elasticloadbalancing.NewDescribeLoadBalancersPaginator(
		client, &elasticloadbalancing.DescribeLoadBalancersInput{},
	)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("describing classic ELBs: %w", err)
		}
		for _, lb := range page.LoadBalancerDescriptions {
			if aws.ToString(lb.VPCId) == vpcID {
				names = append(names, aws.ToString(lb.LoadBalancerName))
			}
		}
	}

	var errs []error
	for _, name := range names {
		Logger.Infof("Deleting classic ELB %s in VPC %s", name, vpcID)
		_, err := client.DeleteLoadBalancer(ctx, &elasticloadbalancing.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(name),
		})
		if err != nil && !hasAWSErrorCode(err, "LoadBalancerNotFound") {
			errs = append(errs, fmt.Errorf("deleting classic ELB %s: %w", name, err))
		}
	}
	return errors.Join(errs...)
}
