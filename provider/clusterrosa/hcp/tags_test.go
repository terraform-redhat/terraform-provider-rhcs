// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var _ = Describe("observedTags", func() {
	ctx := context.Background()
	clusterWithTags := func(rawVersion string, tags map[string]string) *cmv1.Cluster {
		cluster, err := cmv1.NewCluster().
			Version(cmv1.NewVersion().RawID(rawVersion)).
			AWS(cmv1.NewAWS().Tags(tags)).
			Build()
		Expect(err).NotTo(HaveOccurred())
		return cluster
	}
	tagMap := func(tags map[string]string) types.Map {
		elements := map[string]attr.Value{}
		for key, value := range tags {
			elements[key] = types.StringValue(value)
		}
		return types.MapValueMust(types.StringType, elements)
	}

	It("keeps the prior value on clusters older than 4.20", func() {
		prior := tagMap(map[string]string{"team": "platform"})
		observed, err := observedTags(ctx, prior, clusterWithTags("4.19.12", map[string]string{"team": "data"}))
		Expect(err).NotTo(HaveOccurred())
		Expect(observed).To(Equal(prior))
	})

	It("reports out-of-band tag changes on 4.20 and later", func() {
		prior := tagMap(map[string]string{"team": "platform"})
		observed, err := observedTags(ctx, prior,
			clusterWithTags("4.20.1", map[string]string{"team": "data", "Cost Center": "Finance Team"}))
		Expect(err).NotTo(HaveOccurred())
		Expect(observed).To(Equal(tagMap(map[string]string{"team": "data", "Cost Center": "Finance Team"})))
	})

	It("leaves unmanaged tags out of state, such as after import without tags", func() {
		observed, err := observedTags(ctx, types.MapNull(types.StringType),
			clusterWithTags("4.20.1", map[string]string{"team": "platform"}))
		Expect(err).NotTo(HaveOccurred())
		Expect(observed.IsNull()).To(BeTrue())
	})

	It("reports tags added outside Terraform to a managed empty map", func() {
		observed, err := observedTags(ctx, tagMap(map[string]string{}),
			clusterWithTags("4.20.1", map[string]string{"team": "platform"}))
		Expect(err).NotTo(HaveOccurred())
		Expect(observed).To(Equal(tagMap(map[string]string{"team": "platform"})))
	})

	It("reports removal of all managed tags as an empty map that stays managed", func() {
		observed, err := observedTags(ctx, tagMap(map[string]string{"team": "platform"}),
			clusterWithTags("4.20.1", nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(observed).To(Equal(tagMap(map[string]string{})))
	})

	It("keeps a configured empty map when there are no tags", func() {
		prior := tagMap(map[string]string{})
		observed, err := observedTags(ctx, prior, clusterWithTags("4.20.1", nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(observed).To(Equal(prior))
	})
})
