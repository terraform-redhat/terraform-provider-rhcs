package resource

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/openshift-online/ocm-common/pkg/cluster/validations"
	diskValidator "github.com/openshift-online/ocm-common/pkg/machinepool/validations"
	kmsArnRegexpValidator "github.com/openshift-online/ocm-common/pkg/resource/validations"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	rosaTypes "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common/types"
)

var privateHostedZoneRoleArnRE = regexp.MustCompile(
	`^arn:aws:iam::\d{12}:role(?:(?:\/?.+\/?)?)(?:\/[0-9A-Za-z\\+\\.@_,-]{1,64})$`,
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

func (c *Cluster) CreateNodes(clusterTopology rosaTypes.ClusterTopology, autoScalingEnabled bool, replicas *int64, minReplicas *int64,
	maxReplicas *int64, computeMachineType *string, labels map[string]string,
	availabilityZones []string, multiAZ bool, workerDiskSize *int64, version *string) error {
	nodes := cmv1.NewClusterNodes()
	if computeMachineType != nil {
		nodes.ComputeMachineType(
			cmv1.NewMachineType().ID(*computeMachineType),
		)
	}

	if workerDiskSize != nil {
		if clusterTopology == rosaTypes.Hcp {
			err := diskValidator.ValidateNodePoolRootDiskSize(int(*workerDiskSize))
			if err != nil {
				return err
			}
		} else {
			if version == nil {
				return errors.New("Version must be set if a custom root disk size is configured")
			}
			err := diskValidator.ValidateMachinePoolRootDiskSize(*version, int(*workerDiskSize))
			if err != nil {
				return err
			}
		}
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
		// TODO Add availability zones count validation for HCP
		if clusterTopology == rosaTypes.Classic {
			if err := validations.ValidateAvailabilityZonesCount(multiAZ, len(availabilityZones)); err != nil {
				return err
			}
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
		// TODO need to identify the private subnet or remove this validation from TF
		if clusterTopology == rosaTypes.Classic {
			if err := validations.MinReplicasValidator(minReplicasVal, multiAZ, clusterTopology == rosaTypes.Hcp, 0); err != nil {
				return err
			}
		}
		autoscaling.MinReplicas(minReplicasVal)
		maxReplicasVal := 2
		if maxReplicas != nil {
			maxReplicasVal = int(*maxReplicas)
		}
		// TODO need to identify the private subnet or remove this validation from TF
		if clusterTopology == rosaTypes.Classic {
			if err := validations.MaxReplicasValidator(minReplicasVal, maxReplicasVal, multiAZ, clusterTopology == rosaTypes.Hcp, 0); err != nil {
				return err
			}
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
		// TODO need to identify the private subnet or remove this validation from TF
		if clusterTopology == rosaTypes.Classic {
			if err := validations.MinReplicasValidator(replicasVal, multiAZ, clusterTopology == rosaTypes.Hcp, 1); err != nil {
				return err
			}
		}
		nodes.Compute(replicasVal)
	}

	if !nodes.Empty() {
		c.clusterBuilder.Nodes(nodes)
	}

	return nil
}

func (c *Cluster) CreateAWSBuilder(clusterTopology rosaTypes.ClusterTopology,
	awsTags map[string]string, ec2MetadataHttpTokens *string,
	rootVolumeKmsKeyArn *string, etcdKmsKeyArn *string,
	isPrivateLink bool, awsAccountID *string, awsBillingAccountId *string,
	stsBuilder *cmv1.STSBuilder, awsSubnetIDs []string,
	privateHostedZoneID *string, privateHostedZoneRoleARN *string,
	hcpInternalCommunicationPrivateHostedZoneId *string, vpceRoleArn *string,
	additionalComputeSecurityGroupIds []string,
	additionalInfraSecurityGroupIds []string,
	additionalControlPlaneSecurityGroupIds []string,
	additionalAllowedPrincipals []string) error {

	if clusterTopology == rosaTypes.Hcp && awsSubnetIDs == nil {
		return errors.New("Hosted Control Plane clusters must have a pre-configure VPC. Make sure to specify the subnet ids.")
	}

	if clusterTopology == rosaTypes.Classic && etcdKmsKeyArn != nil {
		return errors.New("Etcd encryption using custom KMS is not supported on Classic clusters")
	}

	if isPrivateLink && awsSubnetIDs == nil {
		return errors.New("Clusters with PrivateLink must have a pre-configured VPC. Make sure to specify the subnet ids.")
	}

	awsBuilder := cmv1.NewAWS()

	if awsTags != nil {
		awsBuilder.Tags(awsTags)
	}

	awsBuilder.Ec2MetadataHttpTokens(cmv1.Ec2MetadataHttpTokensOptional)
	if ec2MetadataHttpTokens != nil {
		awsBuilder.Ec2MetadataHttpTokens(cmv1.Ec2MetadataHttpTokens(*ec2MetadataHttpTokens))
	}

	err := c.ProcessKMSKeyARN(rootVolumeKmsKeyArn, awsBuilder)
	if err != nil {
		return err
	}

	if clusterTopology == rosaTypes.Hcp {
		err := c.ProcessEtcKMSKeyARN(etcdKmsKeyArn, awsBuilder)
		if err != nil {
			return err
		}
		if awsBillingAccountId != nil {
			awsBuilder.BillingAccountID(*awsBillingAccountId)
		}
	}

	if awsAccountID != nil {
		awsBuilder.AccountID(*awsAccountID)
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
		if awsSubnetIDs == nil || stsBuilder == nil {
			return errors.New("Shared VPC parameters require STS and SubnetIDs configurations.")
		}
		privateRoleArnField := "PrivateHostedZoneRoleARN"
		if clusterTopology == rosaTypes.Hcp {
			privateRoleArnField = "Route53RoleArn"
		}
		if !privateHostedZoneRoleArnRE.MatchString(*privateHostedZoneRoleARN) {
			return errors.New(fmt.Sprintf("Expected a valid value for %s matching %s. Got %s",
				privateRoleArnField, privateHostedZoneRoleArnRE, *privateHostedZoneRoleARN))
		}
		awsBuilder.PrivateHostedZoneID(*privateHostedZoneID)
		awsBuilder.PrivateHostedZoneRoleARN(*privateHostedZoneRoleARN)
		if clusterTopology == rosaTypes.Hcp && hcpInternalCommunicationPrivateHostedZoneId != nil && vpceRoleArn != nil {
			if !privateHostedZoneRoleArnRE.MatchString(*vpceRoleArn) {
				return errors.New(fmt.Sprintf("Expected a valid value for VpcEndpointRoleArn matching %s. Got %s", privateHostedZoneRoleArnRE, *vpceRoleArn))
			}
			awsBuilder.HcpInternalCommunicationHostedZoneId(*hcpInternalCommunicationPrivateHostedZoneId)
			awsBuilder.VpcEndpointRoleArn(*vpceRoleArn)
		}
	}

	if additionalAllowedPrincipals != nil {
		awsBuilder.AdditionalAllowedPrincipals(additionalAllowedPrincipals...)
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

func (c *Cluster) ProcessEtcKMSKeyARN(kmsKeyARN *string, awsBuilder *cmv1.AWSBuilder) error {
	err := kmsArnRegexpValidator.ValidateKMSKeyARN(kmsKeyARN)
	if err != nil || kmsKeyARN == nil {
		return err
	}
	awsBuilder.EtcdEncryption(cmv1.NewAwsEtcdEncryption().KMSKeyARN(*kmsKeyARN))
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

func CreateSTS(installerRoleARN, supportRoleARN string, masterRoleARN *string, workerRoleARN,
	operatorRolePrefix string, oidcConfigID *string) *cmv1.STSBuilder {
	sts := cmv1.NewSTS()
	sts.RoleARN(installerRoleARN)
	sts.SupportRoleARN(supportRoleARN)
	instanceIamRoles := cmv1.NewInstanceIAMRoles()
	if masterRoleARN != nil {
		instanceIamRoles.MasterRoleARN(*masterRoleARN)
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
