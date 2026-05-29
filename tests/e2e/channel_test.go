package e2e

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/openshift-online/ocm-sdk-go/testing"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Cluster channel", func() {
	defer GinkgoRecover()

	var (
		profileHandler profilehandler.ProfileHandler
		profile        profilehandler.ProfileSpec
		clusterService exec.ClusterService
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())
		profile = profileHandler.Profile()

		if !profile.IsUseChannel() {
			Skip("Cluster channel parameter is not configured for this profile")
		}

		clusterService, err = profileHandler.Services().GetClusterService()
		Expect(err).ToNot(HaveOccurred())
	})

	It("channel is correctly set on cluster creation - [OCM-22441]", ci.Day1Post, ci.High, ci.FeatureClusterChannel,
		func() {
			By("Verify channel in Terraform variables used for cluster creation")
			clusterArgs, err := clusterService.ReadTFVars()
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterArgs.Channel).ToNot(BeNil())
			Expect(clusterArgs.ChannelGroup).To(BeNil())
			expectedChannel := *clusterArgs.Channel

			By("Verify channel on the cluster in OCM")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			channel, ok := clusterResp.Body().GetChannel()
			Expect(ok).To(BeTrue())
			Expect(channel).To(Equal(expectedChannel))

			channelGroup, ok := clusterResp.Body().Version().GetChannelGroup()
			Expect(ok).To(BeTrue())
			Expect(channelGroup).To(Equal(strings.Split(expectedChannel, "-")[0]))

			By("Verify channel in Terraform state")
			resource, err := clusterService.GetStateResource("rhcs_cluster_rosa_hcp", "rosa_hcp_cluster")
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(MatchJQ(`.instances[0].attributes.channel`, expectedChannel))
		})
})
