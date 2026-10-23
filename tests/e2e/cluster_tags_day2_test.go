// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"maps"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Seaplane cluster day-2 tags", ci.SeaplaneValidated, func() {
	It("updates customer AWS tags in place", ci.Day2, ci.High, ci.FeatureClusterMisc, func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profile.Profile().GetClusterType() != constants.SEAPLANE {
			Skip("cluster tag updates require a Seaplane profile")
		}
		if clusterID == "" {
			Skip("an existing profile-managed cluster is required for the tag day-2 workflow")
		}

		tags, err := clusterworkflow.NewSeaplaneClusterTagsUpdater(ctx, profile, clusterID)
		Expect(err).NotTo(HaveOccurred())
		originalClusterID, err := tags.ClusterID(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(originalClusterID).To(Equal(clusterID))
		originalTags, err := tags.Tags(ctx)
		Expect(err).NotTo(HaveOccurred())
		originalTags = maps.Clone(originalTags)
		if originalTags == nil {
			originalTags = map[string]string{}
		}
		DeferCleanup(func(cleanupCtx SpecContext) {
			Expect(tags.UpdateTags(originalTags)).To(Succeed())
			Expect(tags.WaitTagsApplied(cleanupCtx, originalTags)).To(Succeed())
		})

		initialTags := map[string]string{
			"environment": "test",
			"team":        "compute",
			"obsolete":    "remove-me",
		}
		By("Set root-level AWS tags on the existing cluster")
		Expect(tags.UpdateTags(initialTags)).To(Succeed())
		Expect(tags.WaitTagsApplied(ctx, initialTags)).To(Succeed())
		initialActualTags, err := tags.Tags(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(initialActualTags).To(Equal(initialTags))

		updatedTags := map[string]string{
			"environment": "production",
			"team":        "compute",
			"cost-center": "cc-1234",
		}
		By("Update root-level tags, adding and removing keys")
		Expect(tags.UpdateTags(updatedTags)).To(Succeed())
		Expect(tags.WaitTagsApplied(ctx, updatedTags)).To(Succeed())

		By("Verify the updated tags replace the previous map")
		actualTags, err := tags.Tags(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(actualTags).To(Equal(updatedTags))
		Expect(actualTags).NotTo(HaveKey("obsolete"))
		updatedClusterID, err := tags.ClusterID(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(updatedClusterID).To(Equal(originalClusterID), "updating tags should not replace the cluster")
	})
})
