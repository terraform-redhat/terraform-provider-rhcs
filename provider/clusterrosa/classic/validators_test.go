// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package classic

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	rosa "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common"
	rosaTypes "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

func cloneBasicState() *ClusterRosaClassicState {
	state := generateBasicRosaClassicClusterState()
	state.Channel = types.StringNull()
	state.Version = types.StringValue("4.14.0")
	state.ExternalID = types.StringNull()
	state.DisableSCPChecks = types.BoolNull()
	state.Tags = types.MapNull(types.StringType)
	state.EtcdEncryption = types.BoolNull()
	state.BaseDNSDomain = types.StringNull()
	state.AWSSubnetIDs = types.ListNull(types.StringType)
	state.FIPS = types.BoolNull()
	state.AWSPrivateLink = types.BoolNull()
	state.Private = types.BoolNull()
	state.MachineCIDR = types.StringNull()
	state.ServiceCIDR = types.StringNull()
	state.PodCIDR = types.StringNull()
	state.HostPrefix = types.Int64Null()
	state.Ec2MetadataHttpTokens = types.StringNull()
	state.AWSAdditionalControlPlaneSecurityGroupIds = types.ListNull(types.StringType)
	state.AWSAdditionalInfraSecurityGroupIds = types.ListNull(types.StringType)
	state.AWSAdditionalComputeSecurityGroupIds = types.ListNull(types.StringType)
	state.AutoScalingEnabled = types.BoolNull()
	state.ComputeMachineType = types.StringNull()
	state.DefaultMPLabels = types.MapNull(types.StringType)
	state.MultiAZ = types.BoolNull()
	state.WorkerDiskSize = types.Int64Null()
	state.CreateAdminUser = types.BoolNull()
	state.AdminCredentials = rosaTypes.AdminCredentialsNull()
	return state
}

var _ = Describe("Channel and channel_group update validation", func() {
	DescribeTable("validateChannelAndChannelGroupChanges",
		func(mutate func(state, plan *ClusterRosaClassicState), expectedErr bool) {
			state := cloneBasicState()
			plan := cloneBasicState()
			mutate(state, plan)

			diags := validateChannelAndChannelGroupChanges(state, plan)
			Expect(diags.HasError()).To(Equal(expectedErr))
		},
		Entry("unchanged -> ok",
			func(_, _ *ClusterRosaClassicState) {},
			false,
		),
		Entry("channel only -> ok",
			func(state, plan *ClusterRosaClassicState) {
				state.Channel = types.StringValue("stable")
				plan.Channel = types.StringValue("candidate")
			},
			false,
		),
		Entry("channel_group only -> ok",
			func(state, plan *ClusterRosaClassicState) {
				state.ChannelGroup = types.StringValue("stable")
				plan.ChannelGroup = types.StringValue("fast")
			},
			false,
		),
		Entry("channel and channel_group together -> error",
			func(state, plan *ClusterRosaClassicState) {
				state.Channel = types.StringValue("stable")
				state.ChannelGroup = types.StringValue("stable")
				plan.Channel = types.StringValue("candidate")
				plan.ChannelGroup = types.StringValue("fast")
			},
			true,
		),
	)
})

var _ = Describe("Channel group and version update validation", func() {
	DescribeTable("validateChannelGroupAndVersionChanges",
		func(mutate func(state, plan *ClusterRosaClassicState), expectedErr bool) {
			state := cloneBasicState()
			plan := cloneBasicState()
			mutate(state, plan)

			diags := validateChannelGroupAndVersionChanges(state, plan)
			Expect(diags.HasError()).To(Equal(expectedErr))
		},
		Entry("unchanged -> ok",
			func(_, _ *ClusterRosaClassicState) {},
			false,
		),
		Entry("channel_group only -> ok",
			func(state, plan *ClusterRosaClassicState) {
				state.ChannelGroup = types.StringValue("stable")
				plan.ChannelGroup = types.StringValue("fast")
			},
			false,
		),
		Entry("version only -> ok",
			func(state, plan *ClusterRosaClassicState) {
				state.Version = types.StringValue("4.14.0")
				plan.Version = types.StringValue("4.15.0")
			},
			false,
		),
		Entry("channel_group and version together -> error",
			func(state, plan *ClusterRosaClassicState) {
				state.ChannelGroup = types.StringValue("stable")
				state.Version = types.StringValue("4.14.0")
				plan.ChannelGroup = types.StringValue("fast")
				plan.Version = types.StringValue("4.15.0")
			},
			true,
		),
		Entry("version set when state version is null -> counts as version change with channel_group",
			func(state, plan *ClusterRosaClassicState) {
				state.ChannelGroup = types.StringValue("stable")
				state.Version = types.StringNull()
				plan.ChannelGroup = types.StringValue("fast")
				plan.Version = types.StringValue("4.15.0")
			},
			true,
		),
	)
})

var _ = Describe("Channel and version update validation", func() {
	DescribeTable("validateChannelAndVersionChanges",
		func(mutate func(state, plan *ClusterRosaClassicState), expectedErr bool) {
			state := cloneBasicState()
			plan := cloneBasicState()
			mutate(state, plan)

			diags := validateChannelAndVersionChanges(state, plan)
			Expect(diags.HasError()).To(Equal(expectedErr))
		},
		Entry("unchanged -> ok",
			func(_, _ *ClusterRosaClassicState) {},
			false,
		),
		Entry("channel only -> ok",
			func(state, plan *ClusterRosaClassicState) {
				state.Channel = types.StringValue("stable")
				plan.Channel = types.StringValue("candidate")
			},
			false,
		),
		Entry("version only -> ok",
			func(state, plan *ClusterRosaClassicState) {
				state.Version = types.StringValue("4.14.0")
				plan.Version = types.StringValue("4.15.0")
			},
			false,
		),
		Entry("channel and version together -> error",
			func(state, plan *ClusterRosaClassicState) {
				state.Channel = types.StringValue("stable")
				state.Version = types.StringValue("4.14.0")
				plan.Channel = types.StringValue("candidate")
				plan.Version = types.StringValue("4.15.0")
			},
			true,
		),
		Entry("version set when state version is null -> counts as version change with channel",
			func(state, plan *ClusterRosaClassicState) {
				state.Channel = types.StringValue("stable")
				state.Version = types.StringNull()
				plan.Channel = types.StringValue("candidate")
				plan.Version = types.StringValue("4.15.0")
			},
			true,
		),
	)
})

var _ = Describe("Immutable attribute update validation", func() {
	It("returns no diagnostics when state and plan match", func() {
		state := cloneBasicState()
		plan := cloneBasicState()

		diags := validateNoImmutableAttChange(state, plan)
		Expect(diags.HasError()).To(BeFalse())
	})

	It("errors when an immutable attribute changes", func() {
		state := cloneBasicState()
		plan := cloneBasicState()
		plan.Name = types.StringValue("renamed-cluster")

		diags := validateNoImmutableAttChange(state, plan)
		Expect(diags.HasError()).To(BeTrue())
		Expect(diags.Errors()[0].Summary()).To(Equal(common.AssertionErrorSummaryMessage))
	})

	It("errors when an immutable STS field changes", func() {
		state := cloneBasicState()
		plan := cloneBasicState()
		plan.Sts.RoleARN = types.StringValue(roleArn)

		diags := validateNoImmutableAttChange(state, plan)
		Expect(diags.HasError()).To(BeTrue())
		Expect(diags.Errors()[0].Summary()).To(Equal(common.AssertionErrorSummaryMessage))
	})
})

var _ = Describe("HTTP tokens version validation", func() {
	DescribeTable("validateHttpTokensVersion",
		func(mutate func(state *ClusterRosaClassicState), version string, expectErr bool) {
			state := cloneBasicState()
			mutate(state)

			err := validateHttpTokensVersion(context.Background(), state, version)
			if expectErr {
				Expect(err).To(HaveOccurred())
				return
			}
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("null ec2_metadata_http_tokens -> ok",
			func(state *ClusterRosaClassicState) {
				state.Ec2MetadataHttpTokens = types.StringNull()
			},
			"not-a-version",
			false,
		),
		Entry("optional ec2_metadata_http_tokens -> ok",
			func(state *ClusterRosaClassicState) {
				state.Ec2MetadataHttpTokens = types.StringValue(string(cmv1.Ec2MetadataHttpTokensOptional))
			},
			"not-a-version",
			false,
		),
		Entry("invalid version string -> error",
			func(state *ClusterRosaClassicState) {
				state.Ec2MetadataHttpTokens = types.StringValue(string(cmv1.Ec2MetadataHttpTokensRequired))
			},
			"not-a-version",
			true,
		),
	)
})

var _ = Describe("getOcmVersionMinor", func() {
	DescribeTable("extracts major.minor",
		func(input, expected string) {
			Expect(getOcmVersionMinor(input)).To(Equal(expected))
		},
		Entry("semver version", "4.14.5", "4.14"),
		Entry("two-segment version", "4.14", "4.14"),
		Entry("prefixed version falls back to split", "openshift-v4.14.5", "openshift-v4.14"),
	)
})

var _ = Describe("shouldPatchProperties", func() {
	It("returns true when user properties change", func() {
		state := cloneBasicState()
		plan := cloneBasicState()
		plan.Properties = types.MapValueMust(types.StringType, map[string]attr.Value{
			"rosa_creator_arn": types.StringValue("arn:aws:iam::123456789012:user/other"),
		})

		Expect(shouldPatchProperties(state, plan)).To(BeTrue())
	})

	It("returns false when properties and OCM defaults are unchanged", func() {
		state := cloneBasicState()
		plan := cloneBasicState()
		state.OCMProperties = types.MapValueMust(types.StringType, map[string]attr.Value{
			rosa.PropertyRosaTfVersion: types.StringValue(rosa.OCMProperties[rosa.PropertyRosaTfVersion]),
			rosa.PropertyRosaTfCommit:  types.StringValue(rosa.OCMProperties[rosa.PropertyRosaTfCommit]),
		})
		plan.OCMProperties = state.OCMProperties

		Expect(shouldPatchProperties(state, plan)).To(BeFalse())
	})

	It("returns true when OCM default property values drift", func() {
		state := cloneBasicState()
		plan := cloneBasicState()
		state.OCMProperties = types.MapValueMust(types.StringType, map[string]attr.Value{
			rosa.PropertyRosaTfVersion: types.StringValue("stale-version"),
			rosa.PropertyRosaTfCommit:  types.StringValue(rosa.OCMProperties[rosa.PropertyRosaTfCommit]),
		})
		plan.OCMProperties = state.OCMProperties

		Expect(shouldPatchProperties(state, plan)).To(BeTrue())
	})
})
