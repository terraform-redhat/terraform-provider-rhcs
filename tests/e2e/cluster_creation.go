***REMOVED***

***REMOVED***

	// nolint

***REMOVED***

***REMOVED***
***REMOVED***
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
***REMOVED***

var region = "us-west-2"

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Create cluster test", func(***REMOVED*** {
		It("TestExampleNegative", func(***REMOVED*** {

			clusterParam := &EXE.ClusterCreationArgs{
				Token:              token,
				OCMENV:             "staging",
				ClusterName:        "xuelitf",
				OperatorRolePrefix: "xueli",
				AccountRolePrefix:  "xueli",
				Replicas:           3,
				OpenshiftVersion:   "invalid",
				OIDCConfig:         "managed",
	***REMOVED***

			_, err := EXE.CreateMyTFCluster(clusterParam, "-auto-approve"***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("version %s is not in the list", clusterParam.OpenshiftVersion***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("TestExampleCritical", func(***REMOVED*** {
			accRolePrefix := "xueli-2"
			By("Create VPCs"***REMOVED***
			args := &EXE.VPCVariables{
				Name:      "xueli",
				AWSRegion: region,
				MultiAZ:   true,
				VPCCIDR:   "11.0.0.0/16",
				AZIDs:     []string{"us-west-2a", "us-west-2b", "us-west-2c"},
	***REMOVED***
			priSubnets, pubSubnets, zones, err := EXE.CreateAWSVPC(args***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			defer EXE.DestroyAWSVPC(args***REMOVED***

			By("Create account-roles"***REMOVED***
			accRoleParam := &EXE.AccountRolesArgs{
				Token:             token,
				AccountRolePrefix: accRolePrefix,
	***REMOVED***
			_, err = EXE.CreateMyTFAccountRoles(accRoleParam***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			defer EXE.DestroyMyTFAccountRoles(accRoleParam***REMOVED***

			By("Create Cluster"***REMOVED***
			clusterParam := &EXE.ClusterCreationArgs{
				Token:                token,
				OCMENV:               "staging",
				ClusterName:          "xuelitf",
				OperatorRolePrefix:   "xuelitf",
				AccountRolePrefix:    accRolePrefix,
				Replicas:             3,
				AWSRegion:            region,
				AWSAvailabilityZones: zones,
				AWSSubnetIDs:         append(priSubnets, pubSubnets...***REMOVED***,
				MultiAZ:              true,
				MachineCIDR:          args.VPCCIDR,
				OIDCConfig:           "managed",
	***REMOVED***

			clusterID, err := EXE.CreateMyTFCluster(clusterParam, CON.ROSAClassic***REMOVED***
			defer EXE.DestroyMyTFCluster(clusterParam, CON.ROSAClassic***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			Expect(clusterID***REMOVED***.ToNot(BeEmpty(***REMOVED******REMOVED***

***REMOVED******REMOVED***

		It("TestCreateClusterByProfile", func(***REMOVED*** {
			profile := &CI.Profile{
				ClusterName:   "xueli-tf",
				MultiAZ:       false,
				OIDCConfig:    "managed",
				BYOVPC:        true,
				Region:        "us-west-2",
				Version:       "4.13.4",
				InstanceType:  "r5.xlarge",
				STS:           true,
				NetWorkingSet: true,
				ManifestsDIR:  CON.ROSAClassic,
				// OIDCConfig:    "managed",
	***REMOVED***
			clusterID, err := CI.CreateRHCSClusterByProfile(profile***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			fmt.Println(clusterID***REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
