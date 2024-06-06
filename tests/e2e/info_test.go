package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("RHCS Info", func() {
	var rhcsInfoService exec.RhcsInfoService

	BeforeEach(func() {
		var err error
		rhcsInfoService, err = exec.NewRhcsInfoService(constants.RhcsInfoDir)
		Expect(err).ToNot(HaveOccurred())
	})

	It("can verify the state of the rhcs_info data source - [id:68301]",
		ci.Day2, ci.Medium, ci.FeatureIDP, func() {

			By("Creating/Applying rhcs-info resource by terraform")
			rhcsInfoArgs := &exec.RhcsInfoArgs{}
			_, err := rhcsInfoService.Apply(rhcsInfoArgs)
			Expect(err).ShouldNot(HaveOccurred())

			By("Comparing rhcs-info state output to OCM API output")
			currentAccountInfo, err := cms.RetrieveCurrentAccount(ci.RHCSConnection)
			Expect(err).ShouldNot(HaveOccurred())

			// Address the resource's kind and name for the state command
			currentResourceState, err := rhcsInfoService.ShowState("data.rhcs_info.info")
			Expect(err).ShouldNot(HaveOccurred())

			// convert given string to a map of values
			resourceStateMap, err := helper.ParseStringToMap(currentResourceState)
			Expect(err).ToNot(HaveOccurred())

			// comparsion between rhcs_info source to backend api
			Expect(resourceStateMap["account_email"]).To(Equal(currentAccountInfo.Body().Email()))
			Expect(resourceStateMap["account_id"]).To(Equal(currentAccountInfo.Body().ID()))
			Expect(resourceStateMap["account_username"]).To(Equal(currentAccountInfo.Body().Username()))
			Expect(resourceStateMap["organization_id"]).To(Equal(currentAccountInfo.Body().Organization().ID()))
			Expect(resourceStateMap["organization_external_id"]).To(Equal(currentAccountInfo.Body().Organization().ExternalID()))
			Expect(resourceStateMap["organization_name"]).To(Equal(currentAccountInfo.Body().Organization().Name()))
		})
})
