// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/smithy-go"
)

type fakeEC2SecurityGroupClient struct {
	describeVpcsOutput *ec2.DescribeVpcsOutput
	describeVpcsErr    error

	securityGroups            []types.SecurityGroup
	describeSecurityGroupsErr error

	networkInterfacesByGroup   map[string][]types.NetworkInterface
	describeNetworkIfacesError error

	deleteErrorsByGroup map[string][]error
	deleteCallsByGroup  map[string]int
}

func (f *fakeEC2SecurityGroupClient) DescribeVpcs(_ context.Context, _ *ec2.DescribeVpcsInput, _ ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	if f.describeVpcsOutput == nil {
		return &ec2.DescribeVpcsOutput{}, f.describeVpcsErr
	}
	return f.describeVpcsOutput, f.describeVpcsErr
}

func (f *fakeEC2SecurityGroupClient) DescribeSecurityGroups(_ context.Context, _ *ec2.DescribeSecurityGroupsInput, _ ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	if f.describeSecurityGroupsErr != nil {
		return nil, f.describeSecurityGroupsErr
	}
	return &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: f.securityGroups,
	}, nil
}

func (f *fakeEC2SecurityGroupClient) DescribeNetworkInterfaces(_ context.Context, input *ec2.DescribeNetworkInterfacesInput, _ ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	if f.describeNetworkIfacesError != nil {
		return nil, f.describeNetworkIfacesError
	}
	groupID := filterValue(input.Filters, "group-id")
	return &ec2.DescribeNetworkInterfacesOutput{
		NetworkInterfaces: f.networkInterfacesByGroup[groupID],
	}, nil
}

func (f *fakeEC2SecurityGroupClient) DeleteSecurityGroup(_ context.Context, input *ec2.DeleteSecurityGroupInput, _ ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
	if f.deleteCallsByGroup == nil {
		f.deleteCallsByGroup = make(map[string]int)
	}

	groupID := aws.ToString(input.GroupId)
	call := f.deleteCallsByGroup[groupID]
	f.deleteCallsByGroup[groupID] = call + 1

	errorsByCall := f.deleteErrorsByGroup[groupID]
	if call < len(errorsByCall) && errorsByCall[call] != nil {
		return nil, errorsByCall[call]
	}

	return &ec2.DeleteSecurityGroupOutput{}, nil
}

func TestIsKnownLeakedSecurityGroup(t *testing.T) {
	const infraID = "2l26dls6dsbg0im4n971v4nlgorn2mku"
	ownedTag := fmt.Sprintf(clusterOwnedTagTemplate, infraID)

	tests := []struct {
		name  string
		group types.SecurityGroup
		want  bool
	}{
		{
			name: "vpce-private-router with managed tags",
			group: securityGroup(
				"sg-1",
				infraID+"-vpce-private-router",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
			want: true,
		},
		{
			name: "default-sg with owned cluster tag",
			group: securityGroup(
				"sg-2",
				infraID+"-default-sg",
				tag(ownedTag, "owned"),
			),
			want: true,
		},
		{
			name: "matching name but no ownership tags",
			group: securityGroup(
				"sg-3",
				infraID+"-vpce-private-router",
				tag("Name", infraID+"-vpce-private-router"),
			),
			want: false,
		},
		{
			name:  "default security group must be skipped",
			group: securityGroup("sg-4", "default"),
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isKnownLeakedSecurityGroup(tc.group)
			if got != tc.want {
				t.Fatalf("isKnownLeakedSecurityGroup() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDeleteExtraSecurityGroupsSkipsNonTestManagedVPC(t *testing.T) {
	const (
		vpcID   = "vpc-1"
		infraID = "2l26dls6dsbg0im4n971v4nlgorn2mku"
		groupID = "sg-1"
	)

	client := &fakeEC2SecurityGroupClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{{
				VpcId: aws.String(vpcID),
				Tags:  []types.Tag{tag("Name", "non-test-vpc")},
			}},
		},
		securityGroups: []types.SecurityGroup{
			securityGroup(
				groupID,
				infraID+"-vpce-private-router",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
		},
	}

	err := deleteExtraSecurityGroupsWithClient(context.Background(), client, vpcID, securityGroupCleanupOptions{
		maxAttempts:   1,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("deleteExtraSecurityGroupsWithClient() unexpected error: %v", err)
	}
	if client.deleteCallsByGroup[groupID] != 0 {
		t.Fatalf("expected no delete calls, got %d", client.deleteCallsByGroup[groupID])
	}
}

func TestDeleteExtraSecurityGroupsRetriesDependencyViolationAndSucceeds(t *testing.T) {
	const (
		vpcID   = "vpc-1"
		infraID = "2l26dls6dsbg0im4n971v4nlgorn2mku"
		groupID = "sg-1"
	)

	client := &fakeEC2SecurityGroupClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{{
				VpcId: aws.String(vpcID),
				Tags:  []types.Tag{tag(testManagedVPCTagKey, testManagedVPCTagValue)},
			}},
		},
		securityGroups: []types.SecurityGroup{
			securityGroup(
				groupID,
				infraID+"-vpce-private-router",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
		},
		deleteErrorsByGroup: map[string][]error{
			groupID: {
				&smithy.GenericAPIError{Code: "DependencyViolation", Message: "still referenced"},
				nil,
			},
		},
	}

	err := deleteExtraSecurityGroupsWithClient(context.Background(), client, vpcID, securityGroupCleanupOptions{
		maxAttempts:   3,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("deleteExtraSecurityGroupsWithClient() unexpected error: %v", err)
	}
	if client.deleteCallsByGroup[groupID] != 2 {
		t.Fatalf("expected 2 delete calls, got %d", client.deleteCallsByGroup[groupID])
	}
}

func TestDeleteExtraSecurityGroupsReturnsErrorWhenSecurityGroupStaysAttached(t *testing.T) {
	const (
		vpcID   = "vpc-1"
		infraID = "2l26dls6dsbg0im4n971v4nlgorn2mku"
		groupID = "sg-1"
	)

	client := &fakeEC2SecurityGroupClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{{
				VpcId: aws.String(vpcID),
				Tags:  []types.Tag{tag(testManagedVPCTagKey, testManagedVPCTagValue)},
			}},
		},
		securityGroups: []types.SecurityGroup{
			securityGroup(
				groupID,
				infraID+"-default-sg",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
		},
		networkInterfacesByGroup: map[string][]types.NetworkInterface{
			groupID: {
				{
					NetworkInterfaceId: aws.String("eni-1"),
				},
			},
		},
	}

	err := deleteExtraSecurityGroupsWithClient(context.Background(), client, vpcID, securityGroupCleanupOptions{
		maxAttempts:   2,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err == nil {
		t.Fatal("expected error when security group stays attached, got nil")
	}
	if !strings.Contains(err.Error(), "still attached") {
		t.Fatalf("expected attachment error, got: %v", err)
	}
	if client.deleteCallsByGroup[groupID] != 0 {
		t.Fatalf("expected 0 delete calls, got %d", client.deleteCallsByGroup[groupID])
	}
}

func securityGroup(groupID string, groupName string, tags ...types.Tag) types.SecurityGroup {
	return types.SecurityGroup{
		GroupId:   aws.String(groupID),
		GroupName: aws.String(groupName),
		Tags:      tags,
	}
}

func tag(key string, value string) types.Tag {
	return types.Tag{
		Key:   aws.String(key),
		Value: aws.String(value),
	}
}

// ── DeleteNonDefaultSecurityGroups ────────────────────────────────────────────

func TestDeleteNonDefaultSecurityGroups_SkipsDefault(t *testing.T) {
	client := &fakeEC2SecurityGroupClient{
		securityGroups: []types.SecurityGroup{
			securityGroup("sg-default", "default"),
			securityGroup("sg-other", "my-sg"),
		},
		networkInterfacesByGroup: map[string][]types.NetworkInterface{},
	}
	err := deleteNonDefaultSecurityGroupsWithClient(context.Background(), client, "vpc-123", securityGroupCleanupOptions{
		maxAttempts:   1,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.deleteCallsByGroup["sg-default"] != 0 {
		t.Error("default SG should not have been deleted")
	}
	if client.deleteCallsByGroup["sg-other"] != 1 {
		t.Errorf("non-default SG should have been deleted once, got %d calls", client.deleteCallsByGroup["sg-other"])
	}
}

func TestDeleteNonDefaultSecurityGroups_EmptyVPCID(t *testing.T) {
	client := &fakeEC2SecurityGroupClient{}
	err := deleteNonDefaultSecurityGroupsWithClient(context.Background(), client, "", securityGroupCleanupOptions{
		maxAttempts: 1, retryInterval: time.Millisecond, sleep: func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("unexpected error for empty VPC: %v", err)
	}
}

func TestDeleteNonDefaultSecurityGroups_RetriesDependencyViolation(t *testing.T) {
	depErr := &smithy.GenericAPIError{Code: "DependencyViolation"}
	client := &fakeEC2SecurityGroupClient{
		securityGroups: []types.SecurityGroup{
			securityGroup("sg-leak", "leaked-sg"),
		},
		networkInterfacesByGroup: map[string][]types.NetworkInterface{},
		deleteErrorsByGroup: map[string][]error{
			"sg-leak": {depErr, nil},
		},
	}
	err := deleteNonDefaultSecurityGroupsWithClient(context.Background(), client, "vpc-123", securityGroupCleanupOptions{
		maxAttempts:   3,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.deleteCallsByGroup["sg-leak"] != 2 {
		t.Errorf("expected 2 delete attempts, got %d", client.deleteCallsByGroup["sg-leak"])
	}
}

func filterValue(filters []types.Filter, name string) string {
	for _, filter := range filters {
		if aws.ToString(filter.Name) == name && len(filter.Values) > 0 {
			return filter.Values[0]
		}
	}
	return ""
}

// ── retrieveVPC ───────────────────────────────────────────────────────────────

func TestRetrieveVPC_ReturnsNilOnNotFound(t *testing.T) {
	client := &fakeEC2SecurityGroupClient{
		describeVpcsErr: &smithy.GenericAPIError{Code: "InvalidVpcID.NotFound", Message: "vpc not found"},
	}
	vpc, err := retrieveVPC(context.Background(), client, "vpc-nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for NotFound, got %v", err)
	}
	if vpc != nil {
		t.Fatal("expected nil vpc for NotFound")
	}
}

func TestRetrieveVPC_PropagatesOtherErrors(t *testing.T) {
	client := &fakeEC2SecurityGroupClient{
		describeVpcsErr: errors.New("network failure"),
	}
	_, err := retrieveVPC(context.Background(), client, "vpc-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── deleteExtraSecurityGroupsWithClient – describe error ──────────────────────

func TestDeleteExtraSecurityGroups_DescribeSecurityGroupsError(t *testing.T) {
	client := &fakeEC2SecurityGroupClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{{
				VpcId: aws.String("vpc-1"),
				Tags:  []types.Tag{tag(testManagedVPCTagKey, testManagedVPCTagValue)},
			}},
		},
		describeSecurityGroupsErr: errors.New("describe security groups failed"),
	}
	err := deleteExtraSecurityGroupsWithClient(context.Background(), client, "vpc-1", securityGroupCleanupOptions{
		maxAttempts: 1, retryInterval: time.Millisecond, sleep: func(time.Duration) {},
	})
	if err == nil {
		t.Fatal("expected error from DescribeSecurityGroups, got nil")
	}
}

// ── deleteClassicLoadBalancersWithClient ──────────────────────────────────────

type fakeClassicELBClient struct {
	loadBalancers []elbtypes.LoadBalancerDescription
	describeErr   error
	deleteErrs    map[string]error
	deleteCalls   map[string]int
}

func (f *fakeClassicELBClient) DescribeLoadBalancers(_ context.Context, _ *elasticloadbalancing.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	return &elasticloadbalancing.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: f.loadBalancers,
	}, nil
}

func (f *fakeClassicELBClient) DeleteLoadBalancer(_ context.Context, input *elasticloadbalancing.DeleteLoadBalancerInput, _ ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DeleteLoadBalancerOutput, error) {
	if f.deleteCalls == nil {
		f.deleteCalls = make(map[string]int)
	}
	name := aws.ToString(input.LoadBalancerName)
	f.deleteCalls[name]++
	if err, ok := f.deleteErrs[name]; ok {
		return nil, err
	}
	return &elasticloadbalancing.DeleteLoadBalancerOutput{}, nil
}

func TestDeleteClassicLoadBalancers_DeletesMatchingVPC(t *testing.T) {
	client := &fakeClassicELBClient{
		loadBalancers: []elbtypes.LoadBalancerDescription{
			{LoadBalancerName: aws.String("elb-match"), VPCId: aws.String("vpc-target")},
			{LoadBalancerName: aws.String("elb-other"), VPCId: aws.String("vpc-other")},
		},
	}
	err := deleteClassicLoadBalancersWithClient(context.Background(), client, "vpc-target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.deleteCalls["elb-match"] != 1 {
		t.Errorf("expected elb-match deleted once, got %d", client.deleteCalls["elb-match"])
	}
	if client.deleteCalls["elb-other"] != 0 {
		t.Error("elb-other should not have been deleted")
	}
}

func TestDeleteClassicLoadBalancers_ToleratesLoadBalancerNotFound(t *testing.T) {
	client := &fakeClassicELBClient{
		loadBalancers: []elbtypes.LoadBalancerDescription{
			{LoadBalancerName: aws.String("elb-gone"), VPCId: aws.String("vpc-1")},
		},
		deleteErrs: map[string]error{
			"elb-gone": &smithy.GenericAPIError{Code: "LoadBalancerNotFound"},
		},
	}
	err := deleteClassicLoadBalancersWithClient(context.Background(), client, "vpc-1")
	if err != nil {
		t.Fatalf("LoadBalancerNotFound should be tolerated, got: %v", err)
	}
}

func TestDeleteClassicLoadBalancers_ReturnsErrorOnOtherDeleteFailure(t *testing.T) {
	client := &fakeClassicELBClient{
		loadBalancers: []elbtypes.LoadBalancerDescription{
			{LoadBalancerName: aws.String("elb-fail"), VPCId: aws.String("vpc-1")},
		},
		deleteErrs: map[string]error{
			"elb-fail": errors.New("delete failed"),
		},
	}
	err := deleteClassicLoadBalancersWithClient(context.Background(), client, "vpc-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── purgeHostedZoneRecordsWithClient ──────────────────────────────────────────

type fakeRoute53Client struct {
	recordSets []route53types.ResourceRecordSet
	listErr    error
	changeErr  error
	lastChange *route53types.ChangeBatch
	changeCall int
}

func (f *fakeRoute53Client) ListResourceRecordSets(_ context.Context, _ *route53.ListResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &route53.ListResourceRecordSetsOutput{ResourceRecordSets: f.recordSets}, nil
}

func (f *fakeRoute53Client) ChangeResourceRecordSets(_ context.Context, input *route53.ChangeResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
	f.changeCall++
	f.lastChange = input.ChangeBatch
	if f.changeErr != nil {
		return nil, f.changeErr
	}
	return &route53.ChangeResourceRecordSetsOutput{}, nil
}

func TestPurgeHostedZoneRecords_DeletesNonNSSOARecords(t *testing.T) {
	client := &fakeRoute53Client{
		recordSets: []route53types.ResourceRecordSet{
			{Name: aws.String("zone."), Type: route53types.RRTypeNs},
			{Name: aws.String("zone."), Type: route53types.RRTypeSoa},
			{Name: aws.String("api.zone."), Type: route53types.RRTypeCname},
			{Name: aws.String("*.apps.zone."), Type: route53types.RRTypeCname},
		},
	}
	err := purgeHostedZoneRecordsWithClient(context.Background(), client, "Z123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.changeCall != 1 {
		t.Fatalf("expected 1 change call, got %d", client.changeCall)
	}
	if got := len(client.lastChange.Changes); got != 2 {
		t.Fatalf("expected 2 records deleted, got %d", got)
	}
	for _, c := range client.lastChange.Changes {
		if c.Action != route53types.ChangeActionDelete {
			t.Errorf("expected DELETE action, got %s", c.Action)
		}
		if c.ResourceRecordSet.Type == route53types.RRTypeNs || c.ResourceRecordSet.Type == route53types.RRTypeSoa {
			t.Errorf("NS/SOA record should not be deleted: %s", c.ResourceRecordSet.Type)
		}
	}
}

func TestPurgeHostedZoneRecords_NoChangeWhenOnlyNSSOA(t *testing.T) {
	client := &fakeRoute53Client{
		recordSets: []route53types.ResourceRecordSet{
			{Name: aws.String("zone."), Type: route53types.RRTypeNs},
			{Name: aws.String("zone."), Type: route53types.RRTypeSoa},
		},
	}
	err := purgeHostedZoneRecordsWithClient(context.Background(), client, "Z123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.changeCall != 0 {
		t.Errorf("expected no change call, got %d", client.changeCall)
	}
}

func TestPurgeHostedZoneRecords_ReturnsErrorOnListFailure(t *testing.T) {
	client := &fakeRoute53Client{listErr: errors.New("list failed")}
	err := purgeHostedZoneRecordsWithClient(context.Background(), client, "Z123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPurgeHostedZoneRecords_ReturnsErrorOnChangeFailure(t *testing.T) {
	client := &fakeRoute53Client{
		recordSets: []route53types.ResourceRecordSet{
			{Name: aws.String("api.zone."), Type: route53types.RRTypeCname},
		},
		changeErr: errors.New("change failed"),
	}
	err := purgeHostedZoneRecordsWithClient(context.Background(), client, "Z123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPurgeHostedZoneRecords_EmptyZoneID(t *testing.T) {
	if err := PurgeHostedZoneRecords("us-east-1", ""); err != nil {
		t.Fatalf("empty zone ID should be a no-op, got: %v", err)
	}
}

// ── deleteVPCEndpointsWithClient ──────────────────────────────────────────────

type fakeVPCEndpointClient struct {
	endpoints    []types.VpcEndpoint
	describeErr  error
	deleteErr    error
	unsuccessful []types.UnsuccessfulItem
	deletedIDs   []string
	deleteCalls  int
}

func (f *fakeVPCEndpointClient) DescribeVpcEndpoints(_ context.Context, _ *ec2.DescribeVpcEndpointsInput, _ ...func(*ec2.Options)) (*ec2.DescribeVpcEndpointsOutput, error) {
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	return &ec2.DescribeVpcEndpointsOutput{VpcEndpoints: f.endpoints}, nil
}

func (f *fakeVPCEndpointClient) DeleteVpcEndpoints(_ context.Context, input *ec2.DeleteVpcEndpointsInput, _ ...func(*ec2.Options)) (*ec2.DeleteVpcEndpointsOutput, error) {
	f.deleteCalls++
	f.deletedIDs = append(f.deletedIDs, input.VpcEndpointIds...)
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	return &ec2.DeleteVpcEndpointsOutput{Unsuccessful: f.unsuccessful}, nil
}

func TestDeleteVPCEndpoints_DeletesAllInVPC(t *testing.T) {
	client := &fakeVPCEndpointClient{
		endpoints: []types.VpcEndpoint{
			{VpcEndpointId: aws.String("vpce-1")},
			{VpcEndpointId: aws.String("vpce-2")},
		},
	}
	err := deleteVPCEndpointsWithClient(context.Background(), client, "vpc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.deleteCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", client.deleteCalls)
	}
	if got := strings.Join(client.deletedIDs, ","); got != "vpce-1,vpce-2" {
		t.Fatalf("expected both endpoints deleted, got %q", got)
	}
}

func TestDeleteVPCEndpoints_NoEndpointsIsNoOp(t *testing.T) {
	client := &fakeVPCEndpointClient{}
	err := deleteVPCEndpointsWithClient(context.Background(), client, "vpc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.deleteCalls != 0 {
		t.Fatalf("expected no delete call when there are no endpoints, got %d", client.deleteCalls)
	}
}

func TestDeleteVPCEndpoints_DescribeError(t *testing.T) {
	client := &fakeVPCEndpointClient{describeErr: errors.New("describe failed")}
	if err := deleteVPCEndpointsWithClient(context.Background(), client, "vpc-1"); err == nil {
		t.Fatal("expected error from DescribeVpcEndpoints, got nil")
	}
}

func TestDeleteVPCEndpoints_DeleteError(t *testing.T) {
	client := &fakeVPCEndpointClient{
		endpoints: []types.VpcEndpoint{{VpcEndpointId: aws.String("vpce-1")}},
		deleteErr: errors.New("delete failed"),
	}
	if err := deleteVPCEndpointsWithClient(context.Background(), client, "vpc-1"); err == nil {
		t.Fatal("expected error from DeleteVpcEndpoints, got nil")
	}
}

func TestDeleteVPCEndpoints_UnsuccessfulItemsReturnError(t *testing.T) {
	client := &fakeVPCEndpointClient{
		endpoints: []types.VpcEndpoint{{VpcEndpointId: aws.String("vpce-1")}},
		unsuccessful: []types.UnsuccessfulItem{
			{
				ResourceId: aws.String("vpce-1"),
				Error:      &types.UnsuccessfulItemError{Message: aws.String("still in use")},
			},
		},
	}
	err := deleteVPCEndpointsWithClient(context.Background(), client, "vpc-1")
	if err == nil {
		t.Fatal("expected error when an endpoint deletion is unsuccessful, got nil")
	}
	if !strings.Contains(err.Error(), "vpce-1") || !strings.Contains(err.Error(), "still in use") {
		t.Fatalf("expected error to name the endpoint and reason, got: %v", err)
	}
}

func TestDeleteVPCEndpoints_EmptyVPCID(t *testing.T) {
	if err := DeleteVPCEndpoints("us-east-1", ""); err != nil {
		t.Fatalf("empty VPC ID should be a no-op, got: %v", err)
	}
}
