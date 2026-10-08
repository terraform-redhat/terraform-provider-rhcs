// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfplatform "github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	hfrest "github.com/openshift-online/rosa-hyperfleet-api/clientset/rest"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var hyperFleetAWSRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

type hyperFleetBackend struct {
	service exec.HyperfleetClusterService
	client  *hyperfleet.Clientset
	args    *exec.HyperfleetClusterArgs
	oidcID  string
}

type hyperFleetFetcher struct {
	client  *hyperfleet.Clientset
	service exec.HyperfleetClusterService
}

type hyperFleetDestroyer struct {
	cluster   exec.HyperfleetClusterService
	iam       exec.HyperfleetIAMService
	oidc      exec.HyperfleetOidcConfigService
	vpc       exec.HyperfleetVPCService
	client    *hyperfleet.Clientset
	clusterID string
	ownsVPC   bool
	region    string
}

func newHyperFleetDestroyer(profile profilehandler.ProfileHandler, workspace string) (*hyperFleetDestroyer, error) {
	clusterName := config.GetRHCSClusterName()
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	if workspace == "" {
		workspace = clusterName + "-hyperfleet"
	}
	hyperfleetURL, err := requiredEnv("HYPERFLEET_URL")
	if err != nil {
		return nil, err
	}
	region := config.GetRegion()
	if region == "" {
		region = hyperFleetAWSRegionRE.FindString(hyperfleetURL)
	}
	if region == "" {
		return nil, fmt.Errorf(
			"AWS region is not set in REGION and cannot be derived from HYPERFLEET_URL %q",
			hyperfleetURL,
		)
	}
	clusterService, err := exec.NewHyperfleetClusterService(workspace)
	if err != nil {
		return nil, err
	}
	iamService, err := exec.NewHyperfleetIAMService(workspace + "-iam")
	if err != nil {
		return nil, err
	}
	oidcService, err := exec.NewHyperfleetOidcConfigService(workspace + "-oidc")
	if err != nil {
		return nil, err
	}
	vpcService, err := exec.NewHyperfleetVPCService(workspace + "-vpc")
	if err != nil {
		return nil, err
	}
	client, err := buildHyperFleetClient(hyperfleetURL, region)
	if err != nil {
		return nil, err
	}
	clusterID := os.Getenv(config.EnvClusterID)
	if clusterID == "" {
		output, outputErr := clusterService.Output()
		if outputErr != nil {
			return nil, outputErr
		}
		clusterID = output.ClusterID
	}
	if clusterID == "" {
		return nil, fmt.Errorf("%s is not set and Terraform state has no HyperFleet cluster ID", config.EnvClusterID)
	}
	return &hyperFleetDestroyer{
		cluster:   clusterService,
		iam:       iamService,
		oidc:      oidcService,
		vpc:       vpcService,
		client:    client,
		clusterID: clusterID,
		ownsVPC:   len(config.GetSubnetIDList()) == 0 || len(config.GetAvailabilityZoneList()) == 0,
		region:    region,
	}, nil
}

func (d *hyperFleetDestroyer) Destroy(ctx context.Context) error {
	if config.IsNoClusterDestroy() {
		return nil
	}
	if _, err := d.cluster.Destroy(); err != nil {
		return err
	}
	if err := waitForHyperFleetClusterDeletion(ctx, d.client, d.clusterID); err != nil {
		return err
	}
	if _, err := d.iam.Destroy(); err != nil {
		return err
	}
	if _, err := d.oidc.Destroy(); err != nil {
		return err
	}
	if d.ownsVPC {
		vpcOutput, err := d.vpc.Output()
		if err != nil {
			return fmt.Errorf("reading HyperFleet VPC outputs before destroy: %w", err)
		}
		return DestroyHyperFleetVPC(ctx, d.vpc, vpcOutput, d.region)
	}
	return nil
}

// DestroyHyperFleetVPC removes operator-created dependencies before destroying
// the Terraform-managed VPC resources, retrying while asynchronous AWS cleanup completes.
func DestroyHyperFleetVPC(
	ctx context.Context,
	service exec.HyperfleetVPCService,
	output *exec.HyperfleetVPCOutput,
	region string,
) error {
	var lastErr error
	err := wait.PollUntilContextTimeout(ctx, 2*time.Minute, 30*time.Minute, true, func(context.Context) (bool, error) {
		// The cluster operator releases VPC resources asynchronously after the
		// cluster object is deleted, and it leaks resources Terraform does not
		// track (classic ELBs, VPC endpoints, security groups, hosted-zone
		// records) that block terraform destroy. Re-prune them on every retry —
		// not just once — so transient states (e.g. a vpce-private-router
		// security group still pinned by an endpoint ENI that has not finished
		// releasing) resolve within the retry window instead of failing the
		// whole teardown on the first attempt.
		if output != nil {
			if cleanupErr := helper.DeleteClassicLoadBalancers(region, output.VPCID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] classic ELB cleanup failed (will retry): %v", cleanupErr)
				return false, nil
			}
			// Delete VPC endpoints before security groups: their managed ENIs
			// pin the <infra-id>-vpce-private-router security group, so the SG
			// cannot be removed until the endpoints are gone and their ENIs
			// have drained.
			if cleanupErr := helper.DeleteVPCEndpoints(region, output.VPCID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] VPC endpoint cleanup failed (will retry): %v", cleanupErr)
				return false, nil
			}
			if cleanupErr := helper.DeleteNonDefaultSecurityGroups(region, output.VPCID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] security group cleanup failed (will retry): %v", cleanupErr)
				return false, nil
			}
			// The cluster operator writes CNAME records (api.*, *.apps.*) into
			// the private <name>.hypershift.local hosted zone. terraform destroy
			// of aws_route53_zone.hyperfleet fails with HostedZoneNotEmpty while
			// those records remain, so purge them before destroying the VPC.
			if cleanupErr := helper.PurgeHostedZoneRecords(region, output.HostedZoneID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] hosted zone purge failed (will retry): %v", cleanupErr)
				return false, nil
			}
		}
		_, lastErr = service.Destroy()
		if lastErr != nil {
			log.Logger.Infof("[teardown] VPC destroy attempt failed (will retry): %v", lastErr)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if lastErr != nil {
			return lastErr
		}
		return err
	}
	return nil
}

func newHyperFleetFetcher(profile profilehandler.ProfileHandler) (*hyperFleetFetcher, error) {
	hyperfleetURL, err := requiredEnv("HYPERFLEET_URL")
	if err != nil {
		return nil, err
	}
	region := hyperFleetAWSRegionRE.FindString(hyperfleetURL)
	if region == "" {
		return nil, fmt.Errorf("cannot derive AWS region from HYPERFLEET_URL %q", hyperfleetURL)
	}
	client, err := buildHyperFleetClient(hyperfleetURL, region)
	if err != nil {
		return nil, err
	}
	clusterName := config.GetRHCSClusterName()
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	service, err := exec.NewHyperfleetClusterService(clusterName)
	if err != nil {
		return nil, err
	}
	return &hyperFleetFetcher{client: client, service: service}, nil
}

func (f *hyperFleetFetcher) FetchClusterID(ctx context.Context) (string, error) {
	clusterID := os.Getenv(config.EnvClusterID)
	if clusterID == "" {
		output, err := f.service.Output()
		if err != nil {
			// A create-focused run may not have Terraform state yet.
			return "", nil
		}
		clusterID = output.ClusterID
	}
	if clusterID == "" {
		return "", nil
	}
	if _, err := f.client.HyperfleetV1alpha1().Clusters().Get(ctx, clusterID, hfplatform.GetOptions{}); err != nil {
		return "", err
	}
	return clusterID, nil
}

func newHyperFleetBackend(profile profilehandler.ProfileHandler, workspace string) (*Backend, error) {
	if !profile.Profile().IsBYOVPC() {
		return nil, fmt.Errorf("HyperFleet cluster profiles must set byovpc: true")
	}
	hyperfleetURL, err := requiredEnv("HYPERFLEET_URL")
	if err != nil {
		return nil, err
	}
	clusterName := config.GetRHCSClusterName()
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	operatorRolesPrefix := clusterName
	if workspace == "" {
		workspace = clusterName + "-hyperfleet"
	}
	region := config.GetRegion()
	if region == "" {
		region = hyperFleetAWSRegionRE.FindString(hyperfleetURL)
	}
	if region == "" {
		return nil, fmt.Errorf(
			"AWS region is not set in REGION and cannot be derived from HYPERFLEET_URL %q",
			hyperfleetURL,
		)
	}

	var vpcID, subnetID, availabilityZone string
	var vpcService exec.HyperfleetVPCService
	subnetIDs := config.GetSubnetIDList()
	availabilityZones := config.GetAvailabilityZoneList()
	if len(availabilityZones) == 0 {
		availabilityZones = []string{region + "a"}
	}
	availabilityZone = availabilityZones[0]
	if len(subnetIDs) > 0 && len(availabilityZones) > 0 {
		subnetID = subnetIDs[0]
		vpcID, err = vpcIDFromSubnet(subnetID, region)
		if err != nil {
			return nil, err
		}
	} else {
		vpcService, err = exec.NewHyperfleetVPCService(workspace + "-vpc")
		if err != nil {
			return nil, err
		}
		vpcCIDR := envOrDefault("HYPERFLEET_VPC_CIDR", "10.0.0.0/16")
		if _, err := vpcService.Apply(&exec.HyperfleetVPCArgs{
			AWSRegion:        &region,
			NamePrefix:       &clusterName,
			VPCCIDR:          &vpcCIDR,
			AvailabilityZone: &availabilityZone,
		}); err != nil {
			return nil, err
		}
		output, err := vpcService.Output()
		if err != nil {
			return nil, err
		}
		vpcID = output.VPCID
		subnetID = output.PrivateSubnetID
		availabilityZone = output.AvailabilityZone
	}

	service, err := exec.NewHyperfleetClusterService(workspace)
	if err != nil {
		return nil, err
	}
	client, err := buildHyperFleetClient(hyperfleetURL, region)
	if err != nil {
		return nil, err
	}
	oidcService, err := exec.NewHyperfleetOidcConfigService(workspace + "-oidc")
	if err != nil {
		return nil, err
	}
	oidcType := profile.Profile().GetOIDCConfig()
	if oidcType == "" {
		oidcType = "managed"
	}
	if _, err := oidcService.Apply(&exec.HyperfleetOidcConfigArgs{
		HyperfleetURL: &hyperfleetURL,
		Type:          &oidcType,
	}); err != nil {
		return nil, err
	}
	oidcOutput, err := oidcService.Output()
	if err != nil {
		return nil, err
	}
	iamService, err := exec.NewHyperfleetIAMService(workspace + "-iam")
	if err != nil {
		return nil, err
	}
	if _, err := iamService.Apply(&exec.HyperfleetIAMArgs{
		AWSRegion:           &region,
		OperatorRolesPrefix: &operatorRolesPrefix,
		OIDCIssuerURL:       &oidcOutput.IssuerURL,
		OIDCThumbprint:      &oidcOutput.Thumbprint,
	}); err != nil {
		return nil, err
	}
	schedulerProfile := profile.Profile().GetSchedulerProfile()
	var schedulerProfileValue *string
	if schedulerProfile != "" {
		schedulerProfileValue = &schedulerProfile
	}
	backend := &hyperFleetBackend{
		service: service,
		client:  client,
		oidcID:  oidcOutput.OidcConfigID,
		args: &exec.HyperfleetClusterArgs{
			HyperfleetURL:       &hyperfleetURL,
			AWSRegion:           &region,
			ClusterName:         &clusterName,
			OperatorRolesPrefix: &operatorRolesPrefix,
			SubnetID:            &subnetID,
			VPCID:               &vpcID,
			AvailabilityZone:    &availabilityZone,
			OIDCConfigID:        &oidcOutput.OidcConfigID,
			SchedulerProfile:    schedulerProfileValue,
		},
	}
	return &Backend{Lifecycle: backend}, nil
}

func (b *hyperFleetBackend) FetchClusterID(ctx context.Context) (string, error) {
	clusterID := os.Getenv(config.EnvClusterID)
	if clusterID == "" {
		return "", nil
	}
	if _, err := b.client.HyperfleetV1alpha1().Clusters().Get(ctx, clusterID, hfplatform.GetOptions{}); err != nil {
		return "", err
	}
	return clusterID, nil
}

func (b *hyperFleetBackend) Create() (string, error) {
	if _, err := b.service.Apply(b.args); err != nil {
		return "", err
	}
	output, err := b.service.Output()
	if err != nil {
		return "", err
	}
	return output.ClusterID, nil
}

func (b *hyperFleetBackend) WaitReady(ctx context.Context, clusterID string) error {
	if err := b.client.HyperfleetV1alpha1().OidcConfigs().WaitUntil(
		ctx,
		b.oidcID,
		func(config *v1alpha1.OidcConfig) bool {
			if config == nil {
				log.Logger.Infof("[hyperfleet] OIDC config %s not found while waiting for Ready", b.oidcID)
				return false
			}
			log.Logger.Infof("[hyperfleet] OIDC config %s phase: %s", b.oidcID, config.Status.Phase)
			return config.Status.Phase == v1alpha1.OidcConfigPhaseReady
		},
		30*time.Second,
		15*time.Minute,
	); err != nil {
		return err
	}
	return b.client.HyperfleetV1alpha1().Clusters().WaitUntil(
		ctx,
		clusterID,
		func(cluster *v1alpha1.Cluster) bool {
			if cluster == nil {
				log.Logger.Infof("[hyperfleet] cluster %s not found while waiting for Ready", clusterID)
				return false
			}
			log.Logger.Infof("[hyperfleet] cluster %s phase: %s", clusterID, cluster.Status.Phase)
			return cluster.Status.Phase == v1alpha1.ClusterPhaseReady
		},
		30*time.Second,
		90*time.Minute,
	)
}

func waitForHyperFleetClusterDeletion(ctx context.Context, client *hyperfleet.Clientset, clusterID string) error {
	return client.HyperfleetV1alpha1().Clusters().WaitUntil(
		ctx,
		clusterID,
		func(cluster *v1alpha1.Cluster) bool {
			if cluster == nil {
				log.Logger.Infof("[hyperfleet] cluster %s deleted", clusterID)
				return true
			}
			log.Logger.Infof("[hyperfleet] cluster %s phase while deleting: %s", clusterID, cluster.Status.Phase)
			return false
		},
		30*time.Second,
		60*time.Minute,
	)
}

func vpcIDFromSubnet(subnetID, region string) (string, error) {
	ctx := context.Background()
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return "", err
	}
	output, err := ec2.NewFromConfig(awsConfig).
		DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{SubnetIds: []string{subnetID}})
	if err != nil {
		return "", err
	}
	if len(output.Subnets) == 0 || output.Subnets[0].VpcId == nil || *output.Subnets[0].VpcId == "" {
		return "", fmt.Errorf("subnet %q did not return a VPC ID", subnetID)
	}
	return *output.Subnets[0].VpcId, nil
}

func buildHyperFleetClient(hyperfleetURL, region string) (*hyperfleet.Clientset, error) {
	ctx := context.Background()
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	identity, err := sts.NewFromConfig(awsConfig).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return hyperfleet.NewForConfig(&hfrest.Config{
		Host:      hyperfleetURL,
		Region:    region,
		AccountID: *identity.Account,
		CallerARN: *identity.Arn,
		AWSConfig: awsConfig,
	})
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}
	return value, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

var _ exec.ClusterReadinessLifecycle = (*hyperFleetBackend)(nil)
