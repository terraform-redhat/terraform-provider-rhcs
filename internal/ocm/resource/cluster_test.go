package resource

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kmsArnRegexpValidator "github.com/openshift-online/ocm-common/pkg/resource/validations"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/rosa"
)

var _ = Describe("Cluster", func() {
	var cluster *Cluster
	BeforeEach(func() {
		cluster = NewCluster()
	})
	Context("CreateNodes validation", func() {
		It("Autoscaling disabled minReplicas set - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, pointer(int64(2)), nil, nil, nil, nil, false, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Autoscaling must be enabled in order to set min and max replicas"))
		})
		It("Autoscaling disabled maxReplicas set - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, nil, pointer(int64(2)), nil, nil, nil, false, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Autoscaling must be enabled in order to set min and max replicas"))
		})
		It("Autoscaling disabled replicas smaller than 2 - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, false, pointer(int64(1)), nil, nil, nil, nil, nil, false, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Cluster requires at least 2 compute nodes"))
		})
		It("Autoscaling disabled default replicas - success", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, nil, nil, nil, nil, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			Expect(ocmClusterNode.ComputeMachineType()).To(BeNil())
			Expect(ocmClusterNode.ComputeLabels()).To(BeEmpty())
			Expect(ocmClusterNode.AvailabilityZones()).To(BeEmpty())
			Expect(ocmClusterNode.Compute()).To(Equal(2))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).To(BeNil())
		})
		It("Autoscaling disabled 3 replicas - success", func() {
			err := cluster.CreateNodes(rosa.Classic, false, pointer(int64(3)), nil, nil, nil, nil, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			Expect(ocmClusterNode.ComputeMachineType()).To(BeNil())
			Expect(ocmClusterNode.ComputeLabels()).To(BeEmpty())
			Expect(ocmClusterNode.AvailabilityZones()).To(BeEmpty())
			Expect(ocmClusterNode.Compute()).To(Equal(3))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).To(BeNil())
		})
		It("Autoscaling enabled replicas set - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, true, pointer(int64(2)), nil, nil, nil, nil, nil, false, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("When autoscaling is enabled, replicas should not be configured"))
		})
		It("Autoscaling enabled default minReplicas & maxReplicas - success", func() {
			err := cluster.CreateNodes(rosa.Classic, true, nil, nil, nil, nil, nil, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			Expect(ocmClusterNode.ComputeMachineType()).To(BeNil())
			Expect(ocmClusterNode.ComputeLabels()).To(BeEmpty())
			Expect(ocmClusterNode.AvailabilityZones()).To(BeEmpty())
			Expect(ocmClusterNode.Compute()).To(Equal(0))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).NotTo(BeNil())
			Expect(autoscaleCompute.MinReplicas()).To(Equal(2))
			Expect(autoscaleCompute.MaxReplicas()).To(Equal(2))
		})
		It("Autoscaling enabled default maxReplicas smaller than minReplicas - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, true, nil, pointer(int64(4)), pointer(int64(3)), nil, nil, nil, false, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("max-replicas must be greater or equal to min-replicas"))
		})
		It("Autoscaling enabled set minReplicas & maxReplicas - success", func() {
			err := cluster.CreateNodes(rosa.Classic, true, nil, pointer(int64(2)), pointer(int64(4)), nil, nil, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			Expect(ocmClusterNode.ComputeMachineType()).To(BeNil())
			Expect(ocmClusterNode.ComputeLabels()).To(BeEmpty())
			Expect(ocmClusterNode.AvailabilityZones()).To(BeEmpty())
			Expect(ocmClusterNode.Compute()).To(Equal(0))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).NotTo(BeNil())
			Expect(autoscaleCompute.MinReplicas()).To(Equal(2))
			Expect(autoscaleCompute.MaxReplicas()).To(Equal(4))
		})
		It("Autoscaling disabled set ComputeMachineType - success", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, nil, nil, pointer("asdf"), nil, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			machineType := ocmClusterNode.ComputeMachineType()
			Expect(machineType).NotTo(BeNil())
			Expect(machineType.ID()).To(Equal("asdf"))
			Expect(ocmClusterNode.ComputeLabels()).To(BeEmpty())
			Expect(ocmClusterNode.AvailabilityZones()).To(BeEmpty())
			Expect(ocmClusterNode.Compute()).To(Equal(2))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).To(BeNil())
		})
		It("Autoscaling disabled set compute labels - success", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, nil, nil, nil, map[string]string{"key1": "val1"}, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			Expect(ocmClusterNode.ComputeMachineType()).To(BeNil())
			computeLabels := ocmClusterNode.ComputeLabels()
			Expect(computeLabels).To(HaveLen(1))
			Expect(computeLabels["key1"]).To(Equal("val1"))
			Expect(ocmClusterNode.AvailabilityZones()).To(BeEmpty())
			Expect(ocmClusterNode.Compute()).To(Equal(2))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).To(BeNil())
		})
		It("Autoscaling disabled multiAZ false set one availability zone - success", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, nil, nil, nil, nil, []string{"us-east-1a"}, false, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			Expect(ocmClusterNode.ComputeMachineType()).To(BeNil())
			Expect(ocmClusterNode.ComputeLabels()).To(BeEmpty())
			azs := ocmClusterNode.AvailabilityZones()
			Expect(azs).To(HaveLen(1))
			Expect(ocmClusterNode.Compute()).To(Equal(2))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).To(BeNil())
		})
		It("Autoscaling disabled multiAZ false set three availability zones - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, false, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("The number of availability zones for a single AZ cluster should be 1, instead received: 3"))
		})
		It("Autoscaling disabled multiAZ true set three availability zones and two replicas - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, false, pointer(int64(2)), nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Multi AZ cluster requires at least 3 compute nodes"))
		})
		It("Autoscaling disabled multiAZ true set three availability zones and three replicas - success", func() {
			err := cluster.CreateNodes(rosa.Classic, false, pointer(int64(3)), nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			Expect(ocmClusterNode.ComputeMachineType()).To(BeNil())
			Expect(ocmClusterNode.ComputeLabels()).To(BeEmpty())
			azs := ocmClusterNode.AvailabilityZones()
			Expect(azs).To(HaveLen(3))
			Expect(ocmClusterNode.Compute()).To(Equal(3))
			autoscaleCompute := ocmClusterNode.AutoscaleCompute()
			Expect(autoscaleCompute).To(BeNil())
		})
		It("Autoscaling disabled multiAZ true set one zone - failure", func() {
			err := cluster.CreateNodes(rosa.Classic, false, nil, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Multi AZ cluster requires at least 3 compute nodes"))
		})
		It("Custom disk size", func() {
			diskSize := int64(543)
			err := cluster.CreateNodes(rosa.Classic, false, pointer(int64(3)), nil, nil, nil, nil, nil, false, &diskSize)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			ocmClusterNode := ocmCluster.Nodes()
			Expect(ocmClusterNode).NotTo(BeNil())
			rootVolume := ocmClusterNode.ComputeRootVolume()
			Expect(rootVolume).NotTo(BeNil())
			awsVolume := rootVolume.AWS()
			Expect(awsVolume).NotTo(BeNil())
			Expect(awsVolume.Size()).To(Equal(int(diskSize)))
		})
	})
	Context("CreateAWSBuilder validation", func() {
		It("PrivateLink true subnets IDs empty - failure", func() {
			err := cluster.CreateAWSBuilder(rosa.Classic, nil, nil, nil, true, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Clusters with PrivateLink must have a pre-configured VPC. Make sure to specify the subnet ids."))
		})
		It("PrivateLink false invalid kmsKeyARN - failure", func() {
			err := cluster.CreateAWSBuilder(rosa.Classic, nil, nil, pointer("test"), false, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("expected the kms-key-arn: %s to match %s", "test", kmsArnRegexpValidator.KmsArnRE)))
		})
		It("PrivateLink false empty kmsKeyARN - success", func() {
			err := cluster.CreateAWSBuilder(rosa.Classic, nil, nil, nil, false, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			aws := ocmCluster.AWS()
			Expect(aws.Tags()).To(BeNil())
			Expect(aws.Ec2MetadataHttpTokens()).To(Equal(cmv1.Ec2MetadataHttpTokensOptional))
			Expect(aws.KMSKeyArn()).To(Equal(""))
			Expect(aws.AccountID()).To(Equal(""))
			Expect(aws.PrivateLink()).To(Equal(false))
			Expect(aws.SubnetIDs()).To(BeNil())
			Expect(aws.STS()).To(BeNil())
		})
		It("PrivateLink false invalid Ec2MetadataHttpTokens - success", func() {
			// TODO Need to add validation for Ec2MetadataHttpTokens
			err := cluster.CreateAWSBuilder(rosa.Classic, nil, pointer("test"), nil, false, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			aws := ocmCluster.AWS()
			Expect(aws.Tags()).To(BeNil())
			ec2MetadataHttpTokens := aws.Ec2MetadataHttpTokens()
			Expect(string(ec2MetadataHttpTokens)).To(Equal("test"))
			Expect(aws.KMSKeyArn()).To(Equal(""))
			Expect(aws.AccountID()).To(Equal(""))
			Expect(aws.PrivateLink()).To(Equal(false))
			Expect(aws.SubnetIDs()).To(BeNil())
			Expect(aws.STS()).To(BeNil())
		})
		It("PrivateLink true set all parameters - success", func() {
			validKmsKey := "arn:aws:kms:us-east-1:111111111111:key/mrk-0123456789abcdef0123456789abcdef"
			accountID := "111111111111"
			subnets := []string{"subnet-1a1a1a1a1a1a1a1a1", "subnet-2b2b2b2b2b2b2b2b2", "subnet-3c3c3c3c3c3c3c3c3"}
			installerRole := "arn:aws:iam::111111111111:role/aaa-Installer-Role"
			supportRole := "arn:aws:iam::111111111111:role/aaa-Support-Role"
			masterRole := "arn:aws:iam::111111111111:role/aaa-ControlPlane-Role"
			workerRole := "arn:aws:iam::111111111111:role/aaa-Worker-Role"
			operatorRolePrefix := "bbb"
			oidcConfigID := "1234567dgsdfgh"
			sts := CreateSTS(installerRole, supportRole, &masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID))
			err := cluster.CreateAWSBuilder(rosa.Classic, map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired)),
				pointer(validKmsKey), true, pointer(accountID), nil,
				sts, subnets, nil, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			aws := ocmCluster.AWS()
			tags := aws.Tags()
			Expect(tags).NotTo(BeNil())
			Expect(len(tags)).To(Equal(1))
			Expect(tags["key1"]).To(Equal("val1"))
			ec2MetadataHttpTokens := aws.Ec2MetadataHttpTokens()
			Expect(ec2MetadataHttpTokens).To(Equal(cmv1.Ec2MetadataHttpTokensRequired))
			Expect(aws.KMSKeyArn()).To(Equal(validKmsKey))
			Expect(aws.AccountID()).To(Equal(accountID))
			Expect(aws.PrivateLink()).To(Equal(true))
			subnetsIDs := aws.SubnetIDs()
			Expect(subnetsIDs).NotTo(BeNil())
			Expect(subnetsIDs).To(Equal(subnets))
			stsResult := aws.STS()
			Expect(stsResult).NotTo(BeNil())
			Expect(stsResult.RoleARN()).To(Equal(installerRole))
			Expect(stsResult.SupportRoleARN()).To(Equal(supportRole))
			Expect(stsResult.InstanceIAMRoles().MasterRoleARN()).To(Equal(masterRole))
			Expect(stsResult.InstanceIAMRoles().WorkerRoleARN()).To(Equal(workerRole))
			Expect(stsResult.OidcConfig().ID()).To(Equal(oidcConfigID))
		})
		It("PrivateHostedZone set with all needed parameters - success", func() {
			validKmsKey := "arn:aws:kms:us-east-1:111111111111:key/mrk-0123456789abcdef0123456789abcdef"
			accountID := "111111111111"
			subnets := []string{"subnet-1a1a1a1a1a1a1a1a1", "subnet-2b2b2b2b2b2b2b2b2", "subnet-3c3c3c3c3c3c3c3c3"}
			installerRole := "arn:aws:iam::111111111111:role/aaa-Installer-Role"
			supportRole := "arn:aws:iam::111111111111:role/aaa-Support-Role"
			masterRole := "arn:aws:iam::111111111111:role/aaa-ControlPlane-Role"
			workerRole := "arn:aws:iam::111111111111:role/aaa-Worker-Role"
			privateHZRoleArn := "arn:aws:iam::111111111111:role/aaa-hosted-zone-Role"
			privateHZId := "123123"
			operatorRolePrefix := "bbb"
			oidcConfigID := "1234567dgsdfgh"
			sts := CreateSTS(installerRole, supportRole, &masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID))
			err := cluster.CreateAWSBuilder(rosa.Classic, map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired)),
				pointer(validKmsKey), true, pointer(accountID), nil,
				sts, subnets, &privateHZId, &privateHZRoleArn, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			aws := ocmCluster.AWS()
			Expect(aws.PrivateHostedZoneID()).To(Equal(privateHZId))
			Expect(aws.PrivateHostedZoneRoleARN()).To(Equal(privateHZRoleArn))
		})
		It("PrivateHostedZone set with invalid role ARN - fail", func() {
			validKmsKey := "arn:aws:kms:us-east-1:111111111111:key/mrk-0123456789abcdef0123456789abcdef"
			accountID := "111111111111"
			subnets := []string{"subnet-1a1a1a1a1a1a1a1a1", "subnet-2b2b2b2b2b2b2b2b2", "subnet-3c3c3c3c3c3c3c3c3"}
			installerRole := "arn:aws:iam::111111111111:role/aaa-Installer-Role"
			supportRole := "arn:aws:iam::111111111111:role/aaa-Support-Role"
			masterRole := "arn:aws:iam::111111111111:role/aaa-ControlPlane-Role"
			workerRole := "arn:aws:iam::111111111111:role/aaa-Worker-Role"
			privateHZRoleArn := "arn:aws:iam::234:role/invalidARN"
			privateHZId := "123123"
			operatorRolePrefix := "bbb"
			oidcConfigID := "1234567dgsdfgh"
			sts := CreateSTS(installerRole, supportRole, &masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID))
			err := cluster.CreateAWSBuilder(rosa.Classic, map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired)),
				pointer(validKmsKey), true, pointer(accountID), nil,
				sts, subnets, &privateHZId, &privateHZRoleArn, nil, nil, nil)
			Expect(err).To(HaveOccurred())
		})
		It("PrivateHostedZone set missing STS - fail", func() {
			validKmsKey := "arn:aws:kms:us-east-1:111111111111:key/mrk-0123456789abcdef0123456789abcdef"
			accountID := "111111111111"
			subnets := []string{"subnet-1a1a1a1a1a1a1a1a1", "subnet-2b2b2b2b2b2b2b2b2", "subnet-3c3c3c3c3c3c3c3c3"}
			privateHZRoleArn := "arn:aws:iam::111111111111:role/aaa-hosted-zone-Role"
			privateHZId := "123123"
			err := cluster.CreateAWSBuilder(rosa.Classic, map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired)),
				pointer(validKmsKey), true, pointer(accountID), nil,
				nil, subnets, &privateHZId, &privateHZRoleArn, nil, nil, nil)
			Expect(err).To(HaveOccurred())
		})
		It("PrivateHostedZone set missing subnet ids - fail", func() {
			validKmsKey := "arn:aws:kms:us-east-1:111111111111:key/mrk-0123456789abcdef0123456789abcdef"
			accountID := "111111111111"
			installerRole := "arn:aws:iam::111111111111:role/aaa-Installer-Role"
			supportRole := "arn:aws:iam::111111111111:role/aaa-Support-Role"
			masterRole := "arn:aws:iam::111111111111:role/aaa-ControlPlane-Role"
			workerRole := "arn:aws:iam::111111111111:role/aaa-Worker-Role"
			privateHZRoleArn := "arn:aws:iam::111111111111:role/aaa-hosted-zone-Role"
			privateHZId := "123123"
			operatorRolePrefix := "bbb"
			oidcConfigID := "1234567dgsdfgh"
			sts := CreateSTS(installerRole, supportRole, &masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID))
			err := cluster.CreateAWSBuilder(rosa.Classic, map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired)),
				pointer(validKmsKey), true, pointer(accountID), nil,
				sts, nil, &privateHZId, &privateHZRoleArn, nil, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})
	Context("SetAPIPrivacy validation", func() {
		It("Private STS cluster without private link - success", func() {
			err := cluster.SetAPIPrivacy(true, false, true)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			api := ocmCluster.API()
			Expect(api.Listening()).To(Equal(cmv1.ListeningMethodInternal))
		})
		It("Public STS cluster with private link - failure", func() {
			err := cluster.SetAPIPrivacy(false, true, true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("PrivateLink is only supported on private clusters"))
		})
		It("Private cluster - success", func() {
			err := cluster.SetAPIPrivacy(true, true, true)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			api := ocmCluster.API()
			Expect(api.Listening()).To(Equal(cmv1.ListeningMethodInternal))
		})
		It("Non private cluster - success", func() {
			err := cluster.SetAPIPrivacy(false, false, true)
			Expect(err).NotTo(HaveOccurred())
			ocmCluster, err := cluster.Build()
			Expect(err).NotTo(HaveOccurred())
			api := ocmCluster.API()
			Expect(api.Listening()).To(Equal(cmv1.ListeningMethodExternal))
		})
	})
})

func pointer[T any](src T) *T {
	return &src
}
