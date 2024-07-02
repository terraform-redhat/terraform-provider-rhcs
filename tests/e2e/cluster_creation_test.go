package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("Create cluster", func() {
	It("CreateClusterByProfile", ci.Day1Prepare,
		func() {

			// Generate/build cluster by profile selected
			profile := ci.LoadProfileYamlFileByENV()
			clusterID, err := ci.CreateRHCSClusterByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).ToNot(BeEmpty())
			//TODO: implement waiter for  the private cluster once bastion is implemented
			if constants.GetEnvWithDefault(constants.WaitOperators, "false") == "true" && !profile.Private {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(ci.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("Cluster can be recreated if it was not deleted from tf - [id:66071]", ci.Day3, ci.Medium,
		func() {

			profile := ci.LoadProfileYamlFileByENV()
			originalClusterID := clusterID

			By("Delete cluster via OCM")
			resp, err := cms.DeleteCluster(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Status()).To(Equal(constants.HTTPNoContent))

			By("Wait for the cluster deleted from OCM")
			err = cms.WaitClusterDeleted(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			By("Create cluster again")
			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())
			clusterArgs, err := clusterService.ReadTFVars()
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
			clusterOutput, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterOutput.ClusterID).ToNot(BeEmpty())
			Expect(clusterOutput.ClusterID).ToNot(Equal(originalClusterID))
		})

	It("cluster waiter should fail on error/uninstalling cluster - [id:74472]", ci.Day1Supplemental, ci.Medium,
		func() {
			By("Retrieve cluster with error/uninstalling status")
			params := map[string]interface{}{
				"search": "status.state='error' or status.state='uninstalling'",
				"size":   -1,
			}
			resp, err := cms.ListClusters(ci.RHCSConnection, params)
			Expect(err).ToNot(HaveOccurred())
			clusters := resp.Items().Slice()

			if len(clusters) <= 0 {
				Skip("No cluster in 'error' or 'uninstalling' state to launch this case")
			}
			cluster := clusters[0]

			By("Wait for cluster")
			cwService, err := exec.NewClusterWaiterService()
			Expect(err).ToNot(HaveOccurred())
			clusterWaiterArgs := exec.ClusterWaiterArgs{
				Cluster:      helper.StringPointer(cluster.ID()),
				TimeoutInMin: helper.IntPointer(60),
			}
			_, err = cwService.Apply(&clusterWaiterArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(`Waiting for cluster creation finished with error`))
		})

	It("Cluster can be created within internal cluster waiter - [id:74473]", ci.Day1Supplemental, ci.Medium,
		func() {
			By("Retrieve random profile")
			profile, err := ci.GetRandomProfile()
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve creation args")
			defer func() {
				By("Clean resources")
				ci.DestroyRHCSClusterByProfile(token, profile)
			}()
			clusterArgs, err := ci.GenerateClusterCreationArgsByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
			clusterArgs.DisableClusterWaiter = helper.BoolPointer(true)

			By("Create cluster")
			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
		})

	It("Cluster can be created with host prefix and cidrs - [id:72466]", ci.Day1Supplemental, ci.High,
		func() {
			hostPrefix := 25
			machineCIDR := "10.0.0.0/17"
			serviceCIDR := "172.50.0.0/20"
			podCIDR := "10.128.0.0/16"

			By("Retrieve random profile")
			profile, err := ci.GetRandomProfile(constants.GetHCPClusterTypes()...)
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve creation args")
			defer func() {
				By("Clean resources")
				ci.DestroyRHCSClusterByProfile(token, profile)
			}()
			clusterArgs, err := ci.GenerateClusterCreationArgsByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
			clusterArgs.HostPrefix = helper.IntPointer(hostPrefix)
			clusterArgs.MachineCIDR = helper.StringPointer(machineCIDR)
			clusterArgs.ServiceCIDR = helper.StringPointer(serviceCIDR)
			clusterArgs.PodCIDR = helper.StringPointer(podCIDR)

			By("Create cluster")
			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify cluster configuration")
			clusterOut, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterID := clusterOut.ClusterID
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			clusterNetwork := clusterResp.Body().Network()
			Expect(clusterNetwork.HostPrefix()).To(Equal(hostPrefix))
			Expect(clusterNetwork.MachineCIDR()).To(Equal(machineCIDR))
			Expect(clusterNetwork.ServiceCIDR()).To(Equal(serviceCIDR))
			Expect(clusterNetwork.PodCIDR()).To(Equal(podCIDR))

		})

	It("Cluster can be created with different encryption keys for etcd and data plane - [id:72485]", ci.Day1Supplemental, ci.High,
		func() {
			By("Retrieve random profile")
			profile, err := ci.GetRandomProfile(constants.GetHCPClusterTypes()...)
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve creation args")
			defer func() {
				By("Clean resources")
				ci.DestroyRHCSClusterByProfile(token, profile)
			}()
			if !profile.KMSKey {
				profile.KMSKey = true
			}
			if !profile.Etcd {
				profile.Etcd = true
			}
			clusterArgs, err := ci.GenerateClusterCreationArgsByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
			kmsKeyArn := *clusterArgs.KmsKeyARN

			By("Prepare new KMS Key")
			etcdKMSKeyArn, err := ci.PrepareKMSKey(constants.KMSSecondDir, profile, fmt.Sprintf("%s-2", *clusterArgs.ClusterName), *clusterArgs.AccountRolePrefix, profile.UnifiedAccRolesPath, profile.GetClusterType())
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				kmsService, err := exec.NewKMSService(constants.KMSSecondDir)
				if err != nil {
					Expect(err).ToNot(HaveOccurred())
				}
				kmsService.Destroy()
			}()
			clusterArgs.EtcdKmsKeyARN = helper.StringPointer(etcdKMSKeyArn)

			By("Create cluster")
			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify cluster configuration")
			clusterOut, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterID := clusterOut.ClusterID
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			clusterAWS := clusterResp.Body().AWS()
			Expect(clusterAWS.KMSKeyArn()).To(Equal(kmsKeyArn))
			Expect(clusterAWS.EtcdEncryption().KMSKeyARN()).To(Equal(etcdKMSKeyArn))
		})

	It("Cluster can be created with default Ingress - [id:72516]", ci.Day1Supplemental, ci.High,
		func() {
			By("Retrieve random profile")
			profile, err := ci.GetRandomProfile(constants.GetHCPClusterTypes()...)
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve creation args")
			defer func() {
				By("Clean resources")
				ci.DestroyRHCSClusterByProfile(token, profile)
			}()
			clusterArgs, err := ci.GenerateClusterCreationArgsByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
			clusterArgs.EnableClusterIngress = helper.BoolPointer(true)

			By("Prepare new KMS Key")
			etcdKMSKeyArn, err := ci.PrepareKMSKey(constants.KMSSecondDir, profile, fmt.Sprintf("%s-2", *clusterArgs.ClusterName), *clusterArgs.AccountRolePrefix, profile.UnifiedAccRolesPath, profile.GetClusterType())
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				kmsService, err := exec.NewKMSService(constants.KMSSecondDir)
				if err != nil {
					Expect(err).ToNot(HaveOccurred())
				}
				kmsService.Destroy()
			}()
			clusterArgs.EtcdKmsKeyARN = helper.StringPointer(etcdKMSKeyArn)

			By("Create cluster")
			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify cluster configuration")
			clusterOut, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterID := clusterOut.ClusterID
			ingResp, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(ingResp.Listening()).To(Equal("internal"))
		})
})
