package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var region = "us-west-2"

var _ = Describe("TF Test", func() {
	Describe("Create cluster test", func() {
		It("TestExampleNegative", func() {

			clusterParam := &EXE.ClusterCreationArgs{
				Token:              CI.GetEnvWithDefault(CON.TokenENVName, ""),
				OCMENV:             "staging",
				ClusterName:        "xuelitf",
				OperatorRolePrefix: "xueli",
				AccountRolePrefix:  "xueli",
				Replicas:           3,
				OpenshiftVersion:   "invalid",
				OIDCConfig:         "managed",
			}

			_, err := EXE.CreateMyTFCluster(clusterParam, "-auto-approve")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("version %s is not in the list", clusterParam.OpenshiftVersion))
		})
		It("TestExampleCritical", func() {
			accRolePrefix := "xueli-2"
			By("Create VPCs")
			args := &EXE.VPCArgs{
				Name:      "xueli",
				AWSRegion: region,
				MultiAZ:   true,
				VPCCIDR:   "11.0.0.0/16",
				AZIDs:     []string{"us-west-2a", "us-west-2b", "us-west-2c"},
			}
			priSubnets, pubSubnets, zones, err := EXE.CreateAWSVPC(args)
			Expect(err).ToNot(HaveOccurred())
			defer EXE.DestroyAWSVPC(args)

			By("Create account-roles")
			accRoleParam := &EXE.AccountRolesArgs{
				Token:             CI.GetEnvWithDefault(CON.TokenENVName, ""),
				AccountRolePrefix: accRolePrefix,
			}
			_, err = EXE.CreateMyTFAccountRoles(accRoleParam)
			Expect(err).ToNot(HaveOccurred())
			defer EXE.DestroyMyTFAccountRoles(accRoleParam)

			By("Create Cluster")
			clusterParam := &EXE.ClusterCreationArgs{
				Token:                CI.GetEnvWithDefault(CON.TokenENVName, ""),
				OCMENV:               "staging",
				ClusterName:          "xuelitf",
				OperatorRolePrefix:   "xuelitf",
				AccountRolePrefix:    accRolePrefix,
				Replicas:             3,
				AWSRegion:            region,
				AWSAvailabilityZones: zones,
				AWSSubnetIDs:         append(priSubnets, pubSubnets...),
				MultiAZ:              true,
				MachineCIDR:          args.VPCCIDR,
				OIDCConfig:           "managed",
			}

			clusterID, err := EXE.CreateMyTFCluster(clusterParam, CON.ROSAClassic)
			defer EXE.DestroyMyTFCluster(clusterParam, CI.GetEnvWithDefault(CON.TokenENVName, ""))
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).ToNot(BeEmpty())

		})

		It("TestClusterE2EFlowByProfile", func() {

			// Generate/build cluster by profile selected
			profile, creationArgs, manifests_dir := CI.PrepareRHCSClusterByProfileENV()

			// Create rhcs cluster
			clusterID, err := CI.CreateRHCSClusterByProfile(profile, creationArgs, manifests_dir)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).ToNot(BeEmpty())

			// Destroy cluster's resources
			CI.DestroyRHCSClusterByCreationArgs(creationArgs, manifests_dir)
		})
	})
})
