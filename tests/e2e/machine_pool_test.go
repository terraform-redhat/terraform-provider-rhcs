package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-online/ocm-sdk-go/logging"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var logger logging.Logger

var _ = Describe("TF Test", func() {
	Describe("Create MachinePool test cases", func() {
		var mpService *exe.MachinePoolService
		BeforeEach(func() {
			mpService = exe.NewMachinePoolService(con.MachinePoolDir)
		})
		AfterEach(func() {
			err := mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
			return
		})
		It("MachinePoolExampleNegative", func() {

			MachinePoolArgs := &exe.MachinePoolArgs{
				Token:       token,
				Cluster:     clusterID,
				Replicas:    3,
				MachineType: "invalid",
				Name:        "testmp",
			}

			err := mpService.Create(MachinePoolArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("is not supported for cloud provider 'aws'"))
		})
		Context("Author:amalykhi-High-OCP-64757 @OCP-64757 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-64757 Create a second machine pool", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("Create a second machine pool")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64757"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
				}

				err := mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())
				By("Verify the parameters of the created machinepool")
				mpOut, err := mpService.Output()
				Expect(err).ToNot(HaveOccurred())
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Replicas()).To(Equal(mpOut.Replicas))
				Expect(mpResponseBody.InstanceType()).To(Equal(mpOut.MachineType))
				Expect(mpResponseBody.ID()).To(Equal(mpOut.Name))
			})
		})
		Context("Author:amalykhi-High-OCP-64905 @OCP-64905 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-64905 Edit/delete second machinepool labels", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("Create additional machinepool with labels")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64905"
				creationLabels := map[string]string{"fo1": "bar1", "fo2": "baz2"}
				updatingLabels := map[string]string{"fo1": "bar3", "fo3": "baz3"}
				emptyLabels := map[string]string{}
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
					Labels:      creationLabels,
				}

				err := mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())
				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Labels()).To(Equal(creationLabels))
				By("Edit the labels of the machinepool")
				MachinePoolArgs.Labels = updatingLabels
				err = mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Labels()).To(Equal(updatingLabels))

				By("Delete the labels of the machinepool")
				MachinePoolArgs.Labels = emptyLabels
				err = mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Labels()).To(BeNil())

			})
		})
	})
})
