package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Trust policy external ID", func() {
	var (
		clusterService exec.ClusterService
		newClusterID   string
		externalID     = fmt.Sprintf("rhcs-e2e-%s", helper.GenerateRandomStringWithSymbols(8))
		err            error
		profileHandler profilehandler.ProfileHandler
	)

	AfterEach(func() {
		if newClusterID != "" {
			_, err := cms.DeleteCluster(cms.RHCSConnection, newClusterID)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("creates cluster with external ID - [id:trust_policy_external_id_create]", ci.Day1, ci.Medium, func() {
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		clusterService, err = profileHandler.Services().GetClusterService()
		Expect(err).ToNot(HaveOccurred())

		clusterName := helper.GenerateRandomName("trust-policy-extid", 2)
		args := &exec.ClusterArgs{
			ClusterName:              helper.StringPointer(clusterName),
			StsTrustPolicyExternalID: helper.StringPointer(externalID),
		}

		_, err := clusterService.Apply(args)
		Expect(err).ToNot(HaveOccurred())

		output, err := clusterService.Output()
		Expect(err).ToNot(HaveOccurred())
		newClusterID = output.ClusterID

		resp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, newClusterID)
		Expect(err).ToNot(HaveOccurred())

		sts, ok := resp.Body().AWS().GetSTS()
		Expect(ok).To(BeTrue())

		actualExternalID, hasExternalID := sts.GetExternalID()
		Expect(hasExternalID).To(BeTrue())
		Expect(actualExternalID).To(Equal(externalID))
	})
})
