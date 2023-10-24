***REMOVED***

***REMOVED***

	// nolint
***REMOVED***
***REMOVED***
***REMOVED***
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
***REMOVED***
***REMOVED***
***REMOVED***

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Create MachinePool test cases", func(***REMOVED*** {
		var mpService *exe.MachinePoolService
		profile := ci.LoadProfileYamlFileByENV(***REMOVED***
		BeforeEach(func(***REMOVED*** {
			mpService = exe.NewMachinePoolService(con.MachinePoolDir***REMOVED***
***REMOVED******REMOVED***
		AfterEach(func(***REMOVED*** {
			err := mpService.Destroy(***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("MachinePoolExampleNegative", func(***REMOVED*** {

			MachinePoolArgs := &exe.MachinePoolArgs{
				Token:       token,
				Cluster:     clusterID,
				Replicas:    3,
				MachineType: "invalid",
				Name:        "testmp",
	***REMOVED***

			err := mpService.Create(MachinePoolArgs***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("is not supported for cloud provider 'aws'"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-High-OCP-64757 @OCP-64757 @amalykhi", func(***REMOVED*** {
			It("Author:amalykhi-High-OCP-64757 Create a second machine pool", ci.Day2, ci.High, ci.FeatureMachinepool, func(***REMOVED*** {
				By("Create a second machine pool"***REMOVED***
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64757"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
		***REMOVED***

				err := mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				_, err = mpService.Output(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the created machinepool"***REMOVED***
				mpOut, err := mpService.Output(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Replicas(***REMOVED******REMOVED***.To(Equal(mpOut.Replicas***REMOVED******REMOVED***
				Expect(mpResponseBody.InstanceType(***REMOVED******REMOVED***.To(Equal(mpOut.MachineType***REMOVED******REMOVED***
				Expect(mpResponseBody.ID(***REMOVED******REMOVED***.To(Equal(mpOut.Name***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-High-OCP-64905 @OCP-64905 @amalykhi", func(***REMOVED*** {
			It("Author:amalykhi-High-OCP-64905 Edit/delete second machinepool labels", ci.Day2, ci.High, ci.FeatureMachinepool, func(***REMOVED*** {
				By("Create additional machinepool with labels"***REMOVED***
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
		***REMOVED***

				err := mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the created machinepool"***REMOVED***
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Labels(***REMOVED******REMOVED***.To(Equal(creationLabels***REMOVED******REMOVED***

				By("Edit the labels of the machinepool"***REMOVED***
				MachinePoolArgs.Labels = updatingLabels
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Labels(***REMOVED******REMOVED***.To(Equal(updatingLabels***REMOVED******REMOVED***

				By("Delete the labels of the machinepool"***REMOVED***
				MachinePoolArgs.Labels = emptyLabels
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the updated machinepool"***REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Labels(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***

	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-Critical-OCP-68296 @OCP-68296 @amalykhi", func(***REMOVED*** {
			It("Author:amalykhi-Critical-OCP-68296 Enable/disable/update autoscaling for additional machinepool", ci.Day2, ci.Critical, ci.FeatureMachinepool, func(***REMOVED*** {
				By("Create additional machinepool with autoscaling"***REMOVED***
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
		***REMOVED***

				err := mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the created machinepool"***REMOVED***
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Autoscaling(***REMOVED***.MinReplicas(***REMOVED******REMOVED***.To(Equal(minReplicas***REMOVED******REMOVED***
				Expect(mpResponseBody.Autoscaling(***REMOVED***.MaxReplicas(***REMOVED******REMOVED***.To(Equal(maxReplicas***REMOVED******REMOVED***

				By("Change the number of replicas of the machinepool"***REMOVED***
				MachinePoolArgs.MinReplicas = minReplicas * 2
				MachinePoolArgs.MaxReplicas = maxReplicas * 2
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the updated machinepool"***REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Autoscaling(***REMOVED***.MinReplicas(***REMOVED******REMOVED***.To(Equal(minReplicas * 2***REMOVED******REMOVED***
				Expect(mpResponseBody.Autoscaling(***REMOVED***.MaxReplicas(***REMOVED******REMOVED***.To(Equal(maxReplicas * 2***REMOVED******REMOVED***

				By("Disable autoscaling of the machinepool"***REMOVED***
				MachinePoolArgs = &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
		***REMOVED***

				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the updated machinepool"***REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Autoscaling(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***

	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-High-OCP-64904 @ocp-64904 @amalykhi", func(***REMOVED*** {
			It("Author:amalykhi-High-OCP-64905 Edit second machinepool taints", ci.Day2, ci.High, ci.FeatureMachinepool, func(***REMOVED*** {
				By("Create additional machinepool with labels"***REMOVED***
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
		***REMOVED***

				err := mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the created machinepool"***REMOVED***
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				respTaints := mpResponseBody.Taints(***REMOVED***
				for index, taint := range respTaints {
					Expect(taint.Effect(***REMOVED******REMOVED***.To(Equal(taints[index]["schedule_type"]***REMOVED******REMOVED***
					Expect(taint.Key(***REMOVED******REMOVED***.To(Equal(taints[index]["key"]***REMOVED******REMOVED***
					Expect(taint.Value(***REMOVED******REMOVED***.To(Equal(taints[index]["value"]***REMOVED******REMOVED***
		***REMOVED***
				By("Edit the existing taint of the machinepool"***REMOVED***
				taint1["key"] = "k2updated"
				taint1["value"] = "val2updated"

				By("Append new one to the machinepool"***REMOVED***
				taints = append(taints, taint2***REMOVED***

				By("Apply the changes to the machinepool"***REMOVED***
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the updated machinepool"***REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				respTaints = mpResponseBody.Taints(***REMOVED***
				for index, taint := range respTaints {
					Expect(taint.Effect(***REMOVED******REMOVED***.To(Equal(taints[index]["schedule_type"]***REMOVED******REMOVED***
					Expect(taint.Key(***REMOVED******REMOVED***.To(Equal(taints[index]["key"]***REMOVED******REMOVED***
					Expect(taint.Value(***REMOVED******REMOVED***.To(Equal(taints[index]["value"]***REMOVED******REMOVED***
		***REMOVED***

				By("Delete the taints of the machinepool"***REMOVED***
				MachinePoolArgs.Taints = emptyTaints
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

				By("Verify the parameters of the updated machinepool"***REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Taints(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***

	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-High-OCP-68283 @OCP-68283 @amalykhi", func(***REMOVED*** {
			It("Author:amalykhi-High-OCP-68283 Check the validations for the machinepool creation rosa clusters", ci.Day2, ci.High, ci.FeatureMachinepool, func(***REMOVED*** {
				By("Check the validations for the machinepool creation rosa cluster"***REMOVED***
				var (
					machinepoolName                                                                                                           = "ocp-68283"
					invalidMachinepoolName                                                                                                    = "%^#@"
					machineType, InvalidInstanceType                                                                                          = "r5.xlarge", "custom-4-16384"
					mpReplicas, minReplicas, maxReplicas, invalidMpReplicas, invalidMinReplicas4Mutilcluster, invalidMaxReplicas4Mutilcluster = 3, 3, 6, -3, 4, 7
				***REMOVED***
				By("Create machinepool with invalid name"***REMOVED***
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    invalidMpReplicas,
					Name:        invalidMachinepoolName,
					MachineType: machineType,
		***REMOVED***
				err := mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
				Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("Expected a valid value for 'name'"***REMOVED******REMOVED***

				By("Create machinepool with invalid replica value"***REMOVED***
				MachinePoolArgs = &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    invalidMpReplicas,
					Name:        machinepoolName,
					MachineType: machineType,
		***REMOVED***
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
				Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("Attribute 'replicas'\nmust be a non-negative integer"***REMOVED******REMOVED***

				By("Create machinepool with invalid instance type"***REMOVED***
				MachinePoolArgs = &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    mpReplicas,
					Name:        machinepoolName,
					MachineType: InvalidInstanceType,
		***REMOVED***
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
				Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("Machine type\n'%s' is not supported for cloud provider", InvalidInstanceType***REMOVED******REMOVED***

				By("Create machinepool with setting replicas and enable-autoscaling at the same time"***REMOVED***
				MachinePoolArgs = &exe.MachinePoolArgs{
					Token:              token,
					Cluster:            clusterID,
					Replicas:           mpReplicas,
					Name:               machinepoolName,
					AutoscalingEnabled: true,
					MachineType:        machineType,
		***REMOVED***
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
				Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("when\nenabling autoscaling, should set value for maxReplicas"***REMOVED******REMOVED***

				By("Create machinepool with setting min-replicas large than max-replicas"***REMOVED***
				MachinePoolArgs = &exe.MachinePoolArgs{
					Token:              token,
					Cluster:            clusterID,
					MinReplicas:        maxReplicas,
					MaxReplicas:        minReplicas,
					Name:               machinepoolName,
					AutoscalingEnabled: true,
					MachineType:        machineType,
		***REMOVED***
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
				Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("'min_replicas' must be less than or equal to 'max_replicas'"***REMOVED******REMOVED***

				By("Create machinepool with setting min-replicas and max-replicas but without setting --enable-autoscaling"***REMOVED***
				MachinePoolArgs = &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					MinReplicas: minReplicas,
					MaxReplicas: maxReplicas,
					Name:        machinepoolName,
					MachineType: machineType,
		***REMOVED***
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
				Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("when\ndisabling autoscaling, cannot set min_replicas and/or max_replicas"***REMOVED******REMOVED***

				By("Create machinepool with setting min-replicas large than max-replicas"***REMOVED***

				if profile.MultiAZ {
					By("Create machinepool with setting min-replicas and max-replicas not multiple 3 for multi-az"***REMOVED***
					MachinePoolArgs = &exe.MachinePoolArgs{
						Token:              token,
						Cluster:            clusterID,
						MinReplicas:        minReplicas,
						MaxReplicas:        invalidMaxReplicas4Mutilcluster,
						Name:               machinepoolName,
						MachineType:        machineType,
						AutoscalingEnabled: true,
			***REMOVED***
					err = mpService.Create(MachinePoolArgs***REMOVED***
					Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
					Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("Multi AZ clusters require that the number of replicas be a\nmultiple of 3"***REMOVED******REMOVED***

					MachinePoolArgs = &exe.MachinePoolArgs{
						Token:              token,
						Cluster:            clusterID,
						MinReplicas:        invalidMinReplicas4Mutilcluster,
						MaxReplicas:        maxReplicas,
						Name:               machinepoolName,
						MachineType:        machineType,
						AutoscalingEnabled: true,
			***REMOVED***
					err = mpService.Create(MachinePoolArgs***REMOVED***
					Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
					Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("Multi AZ clusters require that the number of replicas be a\nmultiple of 3"***REMOVED******REMOVED***

		***REMOVED***

	***REMOVED******REMOVED***
***REMOVED******REMOVED***

	}***REMOVED***
}***REMOVED***
