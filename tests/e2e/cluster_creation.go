package e2e

import (

	// nolint

	"fmt"

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
				Token:              token,
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
				Token:             token,
				AccountRolePrefix: accRolePrefix,
			}
			_, err = EXE.CreateMyTFAccountRoles(accRoleParam)
			Expect(err).ToNot(HaveOccurred())
			defer EXE.DestroyMyTFAccountRoles(accRoleParam)

			By("Create Cluster")
			clusterParam := &EXE.ClusterCreationArgs{
				Token:                token,
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
			defer EXE.DestroyMyTFCluster(clusterParam, CON.ROSAClassic)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).ToNot(BeEmpty())

		})

		It("TestCreateClusterByProfile", func() {
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
			}
			clusterID, err := CI.CreateRHCSClusterByProfile(profile)
			Expect(err).ToNot(HaveOccurred())
			fmt.Println(clusterID)
		})
	})
})
