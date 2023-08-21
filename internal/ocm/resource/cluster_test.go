package resource

***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

var _ = Describe("Cluster", func(***REMOVED*** {
	var cluster *Cluster
	BeforeEach(func(***REMOVED*** {
		cluster = NewCluster(***REMOVED***
	}***REMOVED***
	Context("CreateNodes validation", func(***REMOVED*** {
		It("Autoscaling disabled minReplicas set - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Autoscaling must be enabled in order to set min and max replicas"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled maxReplicas set - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, nil, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Autoscaling must be enabled in order to set min and max replicas"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled replicas smaller than 2 - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(false, pointer(int64(1***REMOVED******REMOVED***, nil, nil, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Cluster requires at least 2 compute nodes"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled default replicas - success", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, nil, nil, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled 3 replicas - success", func(***REMOVED*** {
			err := cluster.CreateNodes(false, pointer(int64(3***REMOVED******REMOVED***, nil, nil, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(3***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling enabled replicas set - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(true, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("When autoscaling is enabled, replicas should not be configured"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling enabled default minReplicas & maxReplicas - success", func(***REMOVED*** {
			err := cluster.CreateNodes(true, nil, nil, nil, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(0***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(autoscaleCompute.MinReplicas(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
			Expect(autoscaleCompute.MaxReplicas(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling enabled default maxReplicas smaller than minReplicas - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(true, nil, pointer(int64(4***REMOVED******REMOVED***, pointer(int64(3***REMOVED******REMOVED***, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("max-replicas must be greater or equal to min-replicas"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling enabled set minReplicas & maxReplicas - success", func(***REMOVED*** {
			err := cluster.CreateNodes(true, nil, pointer(int64(2***REMOVED******REMOVED***, pointer(int64(4***REMOVED******REMOVED***, nil, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(0***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(autoscaleCompute.MinReplicas(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
			Expect(autoscaleCompute.MaxReplicas(***REMOVED******REMOVED***.To(Equal(4***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled set ComputeMachineType - success", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, nil, nil, pointer("asdf"***REMOVED***, nil, nil, false***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			machineType := ocmClusterNode.ComputeMachineType(***REMOVED***
			Expect(machineType***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(machineType.ID(***REMOVED******REMOVED***.To(Equal("asdf"***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled set compute labels - success", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, nil, nil, nil, map[string]string{"key1": "val1"}, nil, false***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			computeLabels := ocmClusterNode.ComputeLabels(***REMOVED***
			Expect(computeLabels***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
			Expect(computeLabels["key1"]***REMOVED***.To(Equal("val1"***REMOVED******REMOVED***
			Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled multiAZ false set one availability zone - success", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, nil, nil, nil, nil, []string{"us-east-1a"}, false***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			azs := ocmClusterNode.AvailabilityZones(***REMOVED***
			Expect(azs***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled multiAZ false set three availability zones - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, false***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("The number of availability zones for a single AZ cluster should be 1, instead received: 3"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled multiAZ true set three availability zones and two replicas - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(false, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Multi AZ cluster requires at least 3 compute nodes"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled multiAZ true set three availability zones and three replicas - success", func(***REMOVED*** {
			err := cluster.CreateNodes(false, pointer(int64(3***REMOVED******REMOVED***, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmClusterNode := ocmCluster.Nodes(***REMOVED***
			Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
			azs := ocmClusterNode.AvailabilityZones(***REMOVED***
			Expect(azs***REMOVED***.To(HaveLen(3***REMOVED******REMOVED***
			Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(3***REMOVED******REMOVED***
			autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
			Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Autoscaling disabled multiAZ true set one zone - failure", func(***REMOVED*** {
			err := cluster.CreateNodes(false, nil, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Multi AZ cluster requires at least 3 compute nodes"***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
	Context("CreateAWSBuilder validation", func(***REMOVED*** {
		It("PrivateLink true subnets IDs empty - failure", func(***REMOVED*** {
			err := cluster.CreateAWSBuilder(nil, nil, nil, true, nil, nil, nil, nil, nil***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Clusters with PrivateLink must have a pre-configured VPC. Make sure to specify the subnet ids."***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateLink false invalid kmsKeyARN - failure", func(***REMOVED*** {
			err := cluster.CreateAWSBuilder(nil, nil, pointer("test"***REMOVED***, false, nil, nil, nil, nil, nil***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal(fmt.Sprintf("Expected a valid value for kms-key-arn matching %s", kmsArnRE***REMOVED******REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateLink false empty kmsKeyARN - success", func(***REMOVED*** {
			err := cluster.CreateAWSBuilder(nil, nil, nil, false, nil, nil, nil, nil, nil***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			aws := ocmCluster.AWS(***REMOVED***
			Expect(aws.Tags(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(aws.Ec2MetadataHttpTokens(***REMOVED******REMOVED***.To(Equal(cmv1.Ec2MetadataHttpTokensOptional***REMOVED******REMOVED***
			Expect(aws.KMSKeyArn(***REMOVED******REMOVED***.To(Equal(""***REMOVED******REMOVED***
			Expect(aws.AccountID(***REMOVED******REMOVED***.To(Equal(""***REMOVED******REMOVED***
			Expect(aws.PrivateLink(***REMOVED******REMOVED***.To(Equal(false***REMOVED******REMOVED***
			Expect(aws.SubnetIDs(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(aws.STS(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateLink false invalid Ec2MetadataHttpTokens - success", func(***REMOVED*** {
			// TODO Need to add validation for Ec2MetadataHttpTokens
			err := cluster.CreateAWSBuilder(nil, pointer("test"***REMOVED***, nil, false, nil, nil, nil, nil, nil***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			aws := ocmCluster.AWS(***REMOVED***
			Expect(aws.Tags(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			ec2MetadataHttpTokens := aws.Ec2MetadataHttpTokens(***REMOVED***
			Expect(string(ec2MetadataHttpTokens***REMOVED******REMOVED***.To(Equal("test"***REMOVED******REMOVED***
			Expect(aws.KMSKeyArn(***REMOVED******REMOVED***.To(Equal(""***REMOVED******REMOVED***
			Expect(aws.AccountID(***REMOVED******REMOVED***.To(Equal(""***REMOVED******REMOVED***
			Expect(aws.PrivateLink(***REMOVED******REMOVED***.To(Equal(false***REMOVED******REMOVED***
			Expect(aws.SubnetIDs(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(aws.STS(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateLink true set all parameters - success", func(***REMOVED*** {
			validKmsKey := "arn:aws:kms:us-east-1:111111111111:key/mrk-0123456789abcdef0123456789abcdef"
			accountID := "111111111111"
			subnets := []string{"subnet-1a1a1a1a1a1a1a1a1", "subnet-2b2b2b2b2b2b2b2b2", "subnet-3c3c3c3c3c3c3c3c3"}
			installerRole := "arn:aws:iam::111111111111:role/aaa-Installer-Role"
			supportRole := "arn:aws:iam::111111111111:role/aaa-Support-Role"
			masterRole := "arn:aws:iam::111111111111:role/aaa-ControlPlane-Role"
			workerRole := "arn:aws:iam::111111111111:role/aaa-Worker-Role"
			operatorRolePrefix := "bbb"
			oidcConfigID := "1234567dgsdfgh"
			sts := CreateSTS(installerRole, supportRole, masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID***REMOVED******REMOVED***
			err := cluster.CreateAWSBuilder(map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired***REMOVED******REMOVED***,
				pointer(validKmsKey***REMOVED***, true, pointer(accountID***REMOVED***,
				sts, subnets, nil, nil***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			aws := ocmCluster.AWS(***REMOVED***
			tags := aws.Tags(***REMOVED***
			Expect(tags***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(len(tags***REMOVED******REMOVED***.To(Equal(1***REMOVED******REMOVED***
			Expect(tags["key1"]***REMOVED***.To(Equal("val1"***REMOVED******REMOVED***
			ec2MetadataHttpTokens := aws.Ec2MetadataHttpTokens(***REMOVED***
			Expect(ec2MetadataHttpTokens***REMOVED***.To(Equal(cmv1.Ec2MetadataHttpTokensRequired***REMOVED******REMOVED***
			Expect(aws.KMSKeyArn(***REMOVED******REMOVED***.To(Equal(validKmsKey***REMOVED******REMOVED***
			Expect(aws.AccountID(***REMOVED******REMOVED***.To(Equal(accountID***REMOVED******REMOVED***
			Expect(aws.PrivateLink(***REMOVED******REMOVED***.To(Equal(true***REMOVED******REMOVED***
			subnetsIDs := aws.SubnetIDs(***REMOVED***
			Expect(subnetsIDs***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(subnetsIDs***REMOVED***.To(Equal(subnets***REMOVED******REMOVED***
			stsResult := aws.STS(***REMOVED***
			Expect(stsResult***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
			Expect(stsResult.RoleARN(***REMOVED******REMOVED***.To(Equal(installerRole***REMOVED******REMOVED***
			Expect(stsResult.SupportRoleARN(***REMOVED******REMOVED***.To(Equal(supportRole***REMOVED******REMOVED***
			Expect(stsResult.InstanceIAMRoles(***REMOVED***.MasterRoleARN(***REMOVED******REMOVED***.To(Equal(masterRole***REMOVED******REMOVED***
			Expect(stsResult.InstanceIAMRoles(***REMOVED***.WorkerRoleARN(***REMOVED******REMOVED***.To(Equal(workerRole***REMOVED******REMOVED***
			Expect(stsResult.OidcConfig(***REMOVED***.ID(***REMOVED******REMOVED***.To(Equal(oidcConfigID***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateHostedZone set with all needed parameters - success", func(***REMOVED*** {
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
			sts := CreateSTS(installerRole, supportRole, masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID***REMOVED******REMOVED***
			err := cluster.CreateAWSBuilder(map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired***REMOVED******REMOVED***,
				pointer(validKmsKey***REMOVED***, true, pointer(accountID***REMOVED***,
				sts, subnets, &privateHZId, &privateHZRoleArn***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			aws := ocmCluster.AWS(***REMOVED***
			Expect(aws.PrivateHostedZoneID(***REMOVED******REMOVED***.To(Equal(privateHZId***REMOVED******REMOVED***
			Expect(aws.PrivateHostedZoneRoleARN(***REMOVED******REMOVED***.To(Equal(privateHZRoleArn***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateHostedZone set with invalid role ARN - fail", func(***REMOVED*** {
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
			sts := CreateSTS(installerRole, supportRole, masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID***REMOVED******REMOVED***
			err := cluster.CreateAWSBuilder(map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired***REMOVED******REMOVED***,
				pointer(validKmsKey***REMOVED***, true, pointer(accountID***REMOVED***,
				sts, subnets, &privateHZId, &privateHZRoleArn***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateHostedZone set missing STS - fail", func(***REMOVED*** {
			validKmsKey := "arn:aws:kms:us-east-1:111111111111:key/mrk-0123456789abcdef0123456789abcdef"
			accountID := "111111111111"
			subnets := []string{"subnet-1a1a1a1a1a1a1a1a1", "subnet-2b2b2b2b2b2b2b2b2", "subnet-3c3c3c3c3c3c3c3c3"}
			privateHZRoleArn := "arn:aws:iam::111111111111:role/aaa-hosted-zone-Role"
			privateHZId := "123123"
			err := cluster.CreateAWSBuilder(map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired***REMOVED******REMOVED***,
				pointer(validKmsKey***REMOVED***, true, pointer(accountID***REMOVED***,
				nil, subnets, &privateHZId, &privateHZRoleArn***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("PrivateHostedZone set missing subnet ids - fail", func(***REMOVED*** {
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
			sts := CreateSTS(installerRole, supportRole, masterRole, workerRole,
				operatorRolePrefix, pointer(oidcConfigID***REMOVED******REMOVED***
			err := cluster.CreateAWSBuilder(map[string]string{"key1": "val1"},
				pointer(string(cmv1.Ec2MetadataHttpTokensRequired***REMOVED******REMOVED***,
				pointer(validKmsKey***REMOVED***, true, pointer(accountID***REMOVED***,
				sts, nil, &privateHZId, &privateHZRoleArn***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
	Context("SetAPIPrivacy validation", func(***REMOVED*** {
		It("Private STS cluster without private link - failure", func(***REMOVED*** {
			err := cluster.SetAPIPrivacy(true, false, true***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Private STS clusters are only supported through AWS PrivateLink"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Private cluster - success", func(***REMOVED*** {
			err := cluster.SetAPIPrivacy(true, true, true***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			api := ocmCluster.API(***REMOVED***
			Expect(api.Listening(***REMOVED******REMOVED***.To(Equal(cmv1.ListeningMethodInternal***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Non private cluster - success", func(***REMOVED*** {
			err := cluster.SetAPIPrivacy(false, true, true***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			ocmCluster, err := cluster.Build(***REMOVED***
			Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
			api := ocmCluster.API(***REMOVED***
			Expect(api.Listening(***REMOVED******REMOVED***.To(Equal(cmv1.ListeningMethodExternal***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

func pointer[T any](src T***REMOVED*** *T {
	return &src
}
