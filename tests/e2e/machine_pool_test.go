package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("TF Test", func() {
	Describe("Create MachinePool test cases", func() {
		var mpService *exe.MachinePoolService
		BeforeEach(func() {
			mpService = exe.NewMachinePoolService(con.MachinePoolDir)
		})
		AfterEach(func() {
			err := mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
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
				_, err = mpService.Output()
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

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Labels()).To(BeNil())

			})
		})
		Context("Author:amalykhi-Critical-OCP-68296 @OCP-68296 @amalykhi", func() {
			It("Author:amalykhi-Critical-OCP-68296 Enable/disable/update autoscaling for additional machinepool", ci.Day2, ci.Critical, ci.FeatureMachinepool, func() {
				By("Create additional machinepool with autoscaling")
				replicas := 9
				minReplicas := 3
				maxReplicas := 6
				machineType := "r5.xlarge"
				name := "ocp-68296"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:              token,
					Cluster:            clusterID,
					MinReplicas:        minReplicas,
					MaxReplicas:        maxReplicas,
					MachineType:        machineType,
					Name:               name,
					AutoscalingEnabled: true,
				}

				err := mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas))
				Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas))

				By("Change the number of replicas of the machinepool")
				MachinePoolArgs.MinReplicas = minReplicas * 2
				MachinePoolArgs.MaxReplicas = maxReplicas * 2
				err = mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas * 2))
				Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas * 2))

				By("Disable autoscaling of the machinepool")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
				}

				err = mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Autoscaling()).To(BeNil())

			})
		})
		Context("Author:amalykhi-High-OCP-64904 @ocp-64904 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-64905 Edit second machinepool taints", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("Create additional machinepool with labels")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64904"
				taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": con.NoExecute}
				taint1 := map[string]string{"key": "k2", "value": "val2", "schedule_type": con.NoSchedule}
				taint2 := map[string]string{"key": "k3", "value": "val3", "schedule_type": con.PreferNoSchedule}
				taints := []map[string]string{taint0, taint1}
				emptyTaints := []map[string]string{}
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
					Taints:      taints,
				}

				err := mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				respTaints := mpResponseBody.Taints()
				for index, taint := range respTaints {
					Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
					Expect(taint.Key()).To(Equal(taints[index]["key"]))
					Expect(taint.Value()).To(Equal(taints[index]["value"]))
				}
				By("Edit the existing taint of the machinepool")
				taint1["key"] = "k2updated"
				taint1["value"] = "val2updated"

				By("Append new one to the machinepool")
				taints = append(taints, taint2)

				By("Apply the changes to the machinepool")
				err = mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				respTaints = mpResponseBody.Taints()
				for index, taint := range respTaints {
					Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
					Expect(taint.Key()).To(Equal(taints[index]["key"]))
					Expect(taint.Value()).To(Equal(taints[index]["value"]))
				}

				By("Delete the taints of the machinepool")
				MachinePoolArgs.Taints = emptyTaints
				err = mpService.Create(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Taints()).To(BeNil())

			})
		})
	})
})
