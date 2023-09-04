package resource

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/openshift-online/ocm-common/pkg/cluster/validations"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var kmsArnRE = regexp.MustCompile(
	`^arn:aws[\w-]*:kms:[\w-]+:\d{12}:key\/mrk-[0-9a-f]{32}$|[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
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
	availabilityZones []string, multiAZ bool) error {
	nodes := cmv1.NewClusterNodes()
	if computeMachineType != nil {
		nodes.ComputeMachineType(
			cmv1.NewMachineType().ID(*computeMachineType),
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

func (c *Cluster) CreateAWSBuilder(awsTags map[string]string, ec2MetadataHttpTokens *string, kmsKeyARN *string,
	isPrivateLink bool, awsAccountID *string, stsBuilder *cmv1.STSBuilder, awsSubnetIDs []string,
	privateHostedZoneID *string, privateHostedZoneRoleARN *string) error {

	if isPrivateLink && awsSubnetIDs == nil {
		return errors.New("Clusters with PrivateLink must have a pre-configured VPC. Make sure to specify the subnet ids.")
	}

	awsBuilder := cmv1.NewAWS()

	if awsTags != nil {
		awsBuilder.Tags(awsTags)
	}

	ec2MetadataHttpTokensVal := cmv1.Ec2MetadataHttpTokensOptional
	if ec2MetadataHttpTokens != nil {
		ec2MetadataHttpTokensVal = cmv1.Ec2MetadataHttpTokens(*ec2MetadataHttpTokens)
	}
	awsBuilder.Ec2MetadataHttpTokens(ec2MetadataHttpTokensVal)

	if kmsKeyARN != nil {
		if !kmsArnRE.MatchString(*kmsKeyARN) {
			return errors.New(fmt.Sprintf("Expected a valid value for kms-key-arn matching %s", kmsArnRE))
		}
		awsBuilder.KMSKeyArn(*kmsKeyARN)
	}

	if awsAccountID != nil {
		awsBuilder.AccountID(*awsAccountID)
	}

	awsBuilder.PrivateLink(isPrivateLink)

	if awsSubnetIDs != nil {
		awsBuilder.SubnetIDs(awsSubnetIDs...)
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
	instanceIamRoles.MasterRoleARN(masterRoleARN)
	instanceIamRoles.WorkerRoleARN(workerRoleARN)
	sts.InstanceIAMRoles(instanceIamRoles)

	// set OIDC config ID
	if oidcConfigID != nil {
		sts.OidcConfig(cmv1.NewOidcConfig().ID(*oidcConfigID))
	}

	sts.OperatorRolePrefix(operatorRolePrefix)
	return sts
}
