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
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Create Classic or HCP MachinePool", ci.Day2, ci.FeatureMachinepool, func() {
	defer GinkgoRecover()

	var (
		mpService      exec.MachinePoolService
		vpcOutput      *exec.VPCOutput
		profileHandler profilehandler.ProfileHandler
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		mpService, err = profileHandler.Services().GetMachinePoolsService()
		Expect(err).ToNot(HaveOccurred())

		if profileHandler.Profile().IsHCP() {
			By("Get vpc output")
			vpcService, err := profileHandler.Services().GetVPCService()
			Expect(err).ToNot(HaveOccurred())
			vpcOutput, err = vpcService.Output()
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		mpService.Destroy()
	})

	getDefaultMPArgs := func(name string, isHCP bool) *exec.MachinePoolArgs {
		replicas := 3
		machineType := "m5.2xlarge"
		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
		}

		if isHCP {
			subnetId := vpcOutput.PrivateSubnets[0]
			mpArgs.AutoscalingEnabled = helper.BoolPointer(false)
			mpArgs.SubnetID = helper.StringPointer(subnetId)
			mpArgs.AutoRepair = helper.BoolPointer(true)
		}
		return mpArgs
	}

	It("can create machinepool with disk size - [id:69144]", ci.Critical, func() {
		By("Create additional machinepool with disk size specified")
		replicas := 3
		machineType := "r5.xlarge"
		name := helper.GenerateRandomName("mp-69144", 2)
		diskSize := 249
		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
			DiskSize:    helper.IntPointer(diskSize),
		}

		if profileHandler.Profile().IsHCP() {
			subnetId := vpcOutput.PrivateSubnets[0]
			mpArgs.AutoscalingEnabled = helper.BoolPointer(false)
			mpArgs.SubnetID = helper.StringPointer(subnetId)
			mpArgs.AutoRepair = helper.BoolPointer(true)
		}

		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		if profileHandler.Profile().IsHCP() {
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.AWSNodePool().RootVolume().Size()).To(Equal(diskSize))
			Expect(mpResponseBody.AWSNodePool().InstanceType()).To(Equal(machineType))
		} else {
			mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.RootVolume().AWS().Size()).To(Equal(diskSize))
			Expect(mpResponseBody.InstanceType()).To(Equal(machineType))
		}

		By("Destroy machinepool")
		_, err = mpService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create another machinepool without disksize set will be created with default value")
		name = helper.GenerateRandomName("mp-69144", 2)
		mpArgs = getDefaultMPArgs(name, profileHandler.Profile().IsHCP())

		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		if profileHandler.Profile().IsHCP() {
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.AWSNodePool().RootVolume().Size()).To(Equal(300))
			Expect(mpResponseBody.AWSNodePool().InstanceType()).To(Equal("m5.2xlarge"))
		} else {
			mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.RootVolume().AWS().Size()).To(Equal(300))
			Expect(mpResponseBody.InstanceType()).To(Equal("m5.2xlarge"))
		}
	})

	It("will validate well for worker disk size field - [id:76345]", ci.Low, func() {
		By("Try to create a machine pool with invalid worker disk size")
		mpName := helper.GenerateRandomName("mp-76345", 2)
		mpArgs := getDefaultMPArgs(mpName, profileHandler.Profile().IsHCP())
		maxDiskSize := constants.MaxDiskSize
		minDiskSize := constants.MinClassicDiskSize
		if profileHandler.Profile().IsHCP() {
			minDiskSize = constants.MinHCPDiskSize
		}

		errMsg := fmt.Sprintf("Must be between %d GiB and %d GiB", minDiskSize, maxDiskSize)

		mpArgs.DiskSize = helper.IntPointer(minDiskSize - 1)
		_, err := mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		helper.ExpectTFErrorContains(err, errMsg)

		mpArgs.DiskSize = helper.IntPointer(maxDiskSize + 1)
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		helper.ExpectTFErrorContains(err, errMsg)

		// TODO OCM-11521 terraform plan doesn't have validation

		By("Create a successful machine pool with disk size specified")
		mpName = helper.GenerateRandomName("mp-76345", 2)
		mpArgs = getDefaultMPArgs(mpName, profileHandler.Profile().IsHCP())
		mpArgs.DiskSize = helper.IntPointer(249)

		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Update disk size of the created machine pool is not allowed")
		mpArgs.DiskSize = helper.IntPointer(320)
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		helper.ExpectTFErrorContains(err, "disk_size, cannot be changed from 249 to 320")
	})
})
