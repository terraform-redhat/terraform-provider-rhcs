package resource

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/openshift-online/ocm-common/pkg/cluster/validations"
	kmsArnRegexpValidator "github.com/openshift-online/ocm-common/pkg/resource/validations"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/rosa"
)

var privateHostedZoneRoleArnRE = regexp.MustCompile(
	`^arn:aws:iam::\d{12}:role\/[A-Za-z0-9]+(?:-[A-Za-z0-9]+)+$`,
)

type Cluster struct {
	clusterBuilder *cmv1.ClusterBuilder
}

func NewCluster() *Cluster {
	return &Cluster{
		clusterBuilder: cmv1.NewCluster(),
	}
}

func (c *Cluster) GetClusterBuilder() *cmv1.ClusterBuilder {
	return c.clusterBuilder
}

func (c *Cluster) Build() (object *cmv1.Cluster, err error) {
	return c.clusterBuilder.Build()
}

func (c *Cluster) CreateNodes(autoScalingEnabled bool, replicas *int64, minReplicas *int64,
	maxReplicas *int64, computeMachineType *string, labels map[string]string,
	availabilityZones []string, multiAZ bool, workerDiskSize *int64) error {
	nodes := cmv1.NewClusterNodes()
	if computeMachineType != nil {
		nodes.ComputeMachineType(
			cmv1.NewMachineType().ID(*computeMachineType),
		)
	}

	if workerDiskSize != nil {
		nodes.ComputeRootVolume(
			cmv1.NewRootVolume().AWS(
				cmv1.NewAWSVolume().Size(int(*workerDiskSize)),
			),
		)
	}

	if labels != nil {
		nodes.ComputeLabels(labels)
	}

	if availabilityZones != nil {
		if err := validations.ValidateAvailabilityZonesCount(multiAZ, len(availabilityZones)); err != nil {
			return err
		}
		nodes.AvailabilityZones(availabilityZones...)
	}

	if autoScalingEnabled {
		if replicas != nil {
			return errors.New("When autoscaling is enabled, replicas should not be configured")
		}

		autoscaling := cmv1.NewMachinePoolAutoscaling()
		minReplicasVal := 2
		if minReplicas != nil {
			minReplicasVal = int(*minReplicas)
		}
		if err := validations.MinReplicasValidator(minReplicasVal, multiAZ, false, 0); err != nil {
			return err
		}
		autoscaling.MinReplicas(minReplicasVal)
		maxReplicasVal := 2
		if maxReplicas != nil {
			maxReplicasVal = int(*maxReplicas)
		}
		if err := validations.MaxReplicasValidator(minReplicasVal, maxReplicasVal, multiAZ, false, 0); err != nil {
			return err
		}
		autoscaling.MaxReplicas(maxReplicasVal)
		if !autoscaling.Empty() {
			nodes.AutoscaleCompute(autoscaling)
		}
	} else {
		if minReplicas != nil || maxReplicas != nil {
			return errors.New("Autoscaling must be enabled in order to set min and max replicas")
		}

		replicasVal := 2
		if replicas != nil {
			replicasVal = int(*replicas)
		}
		if err := validations.MinReplicasValidator(replicasVal, multiAZ, false, 0); err != nil {
			return err
		}
		nodes.Compute(replicasVal)
	}

	if !nodes.Empty() {
		c.clusterBuilder.Nodes(nodes)
	}

	return nil
}

func (c *Cluster) CreateAWSBuilder(clusterTopology rosa.ClusterTopology,
	awsTags map[string]string, ec2MetadataHttpTokens *string, kmsKeyARN *string,
	isPrivateLink bool, awsAccountID *string, awsBillingAccountId *string,
	stsBuilder *cmv1.STSBuilder, awsSubnetIDs []string,
	privateHostedZoneID *string, privateHostedZoneRoleARN *string,
	additionalComputeSecurityGroupIds []string,
	additionalInfraSecurityGroupIds []string,
	additionalControlPlaneSecurityGroupIds []string) error {

	if isPrivateLink && awsSubnetIDs == nil {
		return errors.New("Clusters with PrivateLink must have a pre-configured VPC. Make sure to specify the subnet ids.")
	}

	awsBuilder := cmv1.NewAWS()

	if awsTags != nil {
		awsBuilder.Tags(awsTags)
	}

	if clusterTopology == rosa.Classic {
		ec2MetadataHttpTokensVal := cmv1.Ec2MetadataHttpTokensOptional
		if ec2MetadataHttpTokens != nil {
			ec2MetadataHttpTokensVal = cmv1.Ec2MetadataHttpTokens(*ec2MetadataHttpTokens)
		}
		awsBuilder.Ec2MetadataHttpTokens(ec2MetadataHttpTokensVal)
	}

	err := c.ProcessKMSKeyARN(kmsKeyARN, awsBuilder)
	if err != nil {
		return err
	}

	if awsAccountID != nil {
		awsBuilder.AccountID(*awsAccountID)
	}

	if clusterTopology == rosa.Hcp {
		awsBuilder.BillingAccountID(*awsBillingAccountId)
	}

	awsBuilder.PrivateLink(isPrivateLink)

	if awsSubnetIDs != nil {
		awsBuilder.SubnetIDs(awsSubnetIDs...)
	}

	if additionalComputeSecurityGroupIds != nil {
		awsBuilder.AdditionalComputeSecurityGroupIds(additionalComputeSecurityGroupIds...)
	}

	if additionalInfraSecurityGroupIds != nil {
		awsBuilder.AdditionalInfraSecurityGroupIds(additionalInfraSecurityGroupIds...)
	}

	if additionalControlPlaneSecurityGroupIds != nil {
		awsBuilder.AdditionalControlPlaneSecurityGroupIds(additionalControlPlaneSecurityGroupIds...)
	}

	if stsBuilder != nil {
		awsBuilder.STS(stsBuilder)
	}

	if privateHostedZoneID != nil && privateHostedZoneRoleARN != nil {
		if !privateHostedZoneRoleArnRE.MatchString(*privateHostedZoneRoleARN) {
			return errors.New(fmt.Sprintf("Expected a valid value for PrivateHostedZoneRoleARN matching %s. Got %s", privateHostedZoneRoleArnRE, *privateHostedZoneRoleARN))
		}
		if awsSubnetIDs == nil || stsBuilder == nil {
			return errors.New("PrivateHostedZone parameters require STS and SubnetIDs configurations.")
		}
		awsBuilder.PrivateHostedZoneID(*privateHostedZoneID)
		awsBuilder.PrivateHostedZoneRoleARN(*privateHostedZoneRoleARN)
	}

	c.clusterBuilder.AWS(awsBuilder)

	return nil
}

func (c *Cluster) ProcessKMSKeyARN(kmsKeyARN *string, awsBuilder *cmv1.AWSBuilder) error {
	err := kmsArnRegexpValidator.ValidateKMSKeyARN(kmsKeyARN)
	if err != nil || kmsKeyARN == nil {
		return err
	}
	awsBuilder.KMSKeyArn(*kmsKeyARN)
	return nil
}

func (c *Cluster) SetAPIPrivacy(isPrivate bool, isPrivateLink bool, isSTS bool) error {
	if isSTS && !isPrivate && isPrivateLink {
		return errors.New("PrivateLink is only supported on private clusters")
	}
	api := cmv1.NewClusterAPI()
	if isPrivate {
		api.Listening(cmv1.ListeningMethodInternal)
	} else {
		api.Listening(cmv1.ListeningMethodExternal)
	}
	c.clusterBuilder.API(api)
	return nil
}

func CreateSTS(installerRoleARN, supportRoleARN, masterRoleARN, workerRoleARN,
	operatorRolePrefix string, oidcConfigID *string) *cmv1.STSBuilder {
	sts := cmv1.NewSTS()
	sts.RoleARN(installerRoleARN)
	sts.SupportRoleARN(supportRoleARN)
	instanceIamRoles := cmv1.NewInstanceIAMRoles()
	if masterRoleARN != "" {
		instanceIamRoles.MasterRoleARN(masterRoleARN)
	}
	instanceIamRoles.WorkerRoleARN(workerRoleARN)
	sts.InstanceIAMRoles(instanceIamRoles)

	// set OIDC config ID
	if oidcConfigID != nil {
		sts.OidcConfig(cmv1.NewOidcConfig().ID(*oidcConfigID))
	}

	sts.OperatorRolePrefix(operatorRolePrefix)
	return sts
}
