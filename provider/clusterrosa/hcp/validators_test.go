// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint

	rosa "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

func cloneBasicState() *ClusterRosaHcpState {
	state := generateBasicRosaHcpClusterState()
	state.Channel = types.StringNull()
	state.Version = types.StringValue("4.14.0")
	state.ExternalID = types.StringNull()
	state.Tags = types.MapNull(types.StringType)
	state.EtcdEncryption = types.BoolNull()
	state.FIPS = types.BoolNull()
	state.EtcdKmsKeyArn = types.StringNull()
	state.Private = types.BoolNull()
	state.MachineCIDR = types.StringNull()
	state.ServiceCIDR = types.StringNull()
	state.PodCIDR = types.StringNull()
	state.HostPrefix = types.Int64Null()
	state.AWSAdditionalComputeSecurityGroupIds = types.ListNull(types.StringType)
	state.AutoScalingEnabled = types.BoolNull()
	state.MinReplicas = types.Int64Null()
	state.MaxReplicas = types.Int64Null()
	state.ComputeMachineType = types.StringNull()
	state.Ec2MetadataHttpTokens = types.StringNull()
	state.WorkerDiskSize = types.Int64Null()
	state.CreateAdminUser = types.BoolNull()
	state.BaseDNSDomain = types.StringNull()
	state.ExternalAuthProvidersEnabled = types.BoolNull()
	state.LogForwardersAtClusterCreation = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{}})
	return state
}

var _ = Describe("Channel and channel_group update validation", func() {
	DescribeTable("validateChannelAndChannelGroupChanges",
		func(mutate func(state, plan *ClusterRosaHcpState), expectedErr bool) {
			state := cloneBasicState()
			plan := cloneBasicState()
			mutate(state, plan)

			diags := validateChannelAndChannelGroupChanges(state, plan)
			Expect(diags.HasError()).To(Equal(expectedErr))
		},
		Entry("unchanged -> ok",
			func(_, _ *ClusterRosaHcpState) {},
			false,
		),
		Entry("channel only -> ok",
			func(state, plan *ClusterRosaHcpState) {
				state.Channel = types.StringValue("stable")
				plan.Channel = types.StringValue("candidate")
			},
			false,
		),
		Entry("channel_group only -> ok",
			func(state, plan *ClusterRosaHcpState) {
				state.ChannelGroup = types.StringValue("stable")
				plan.ChannelGroup = types.StringValue("fast")
			},
			false,
		),
		Entry("channel and channel_group together -> error",
			func(state, plan *ClusterRosaHcpState) {
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
		func(mutate func(state, plan *ClusterRosaHcpState), expectedErr bool) {
			state := cloneBasicState()
			plan := cloneBasicState()
			mutate(state, plan)

			diags := validateChannelGroupAndVersionChanges(state, plan)
			Expect(diags.HasError()).To(Equal(expectedErr))
		},
		Entry("unchanged -> ok",
			func(_, _ *ClusterRosaHcpState) {},
			false,
		),
		Entry("channel_group only -> ok",
			func(state, plan *ClusterRosaHcpState) {
				state.ChannelGroup = types.StringValue("stable")
				plan.ChannelGroup = types.StringValue("fast")
			},
			false,
		),
		Entry("version only -> ok",
			func(state, plan *ClusterRosaHcpState) {
				state.Version = types.StringValue("4.14.0")
				plan.Version = types.StringValue("4.15.0")
			},
			false,
		),
		Entry("channel_group and version together -> error",
			func(state, plan *ClusterRosaHcpState) {
				state.ChannelGroup = types.StringValue("stable")
				state.Version = types.StringValue("4.14.0")
				plan.ChannelGroup = types.StringValue("fast")
				plan.Version = types.StringValue("4.15.0")
			},
			true,
		),
		Entry("version set when state version is null -> counts as version change with channel_group",
			func(state, plan *ClusterRosaHcpState) {
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
		func(mutate func(state, plan *ClusterRosaHcpState), expectedErr bool) {
			state := cloneBasicState()
			plan := cloneBasicState()
			mutate(state, plan)

			diags := validateChannelAndVersionChanges(state, plan)
			Expect(diags.HasError()).To(Equal(expectedErr))
		},
		Entry("unchanged -> ok",
			func(_, _ *ClusterRosaHcpState) {},
			false,
		),
		Entry("channel only -> ok",
			func(state, plan *ClusterRosaHcpState) {
				state.Channel = types.StringValue("stable")
				plan.Channel = types.StringValue("candidate")
			},
			false,
		),
		Entry("version only -> ok",
			func(state, plan *ClusterRosaHcpState) {
				state.Version = types.StringValue("4.14.0")
				plan.Version = types.StringValue("4.15.0")
			},
			false,
		),
		Entry("channel and version together -> error",
			func(state, plan *ClusterRosaHcpState) {
				state.Channel = types.StringValue("stable")
				state.Version = types.StringValue("4.14.0")
				plan.Channel = types.StringValue("candidate")
				plan.Version = types.StringValue("4.15.0")
			},
			true,
		),
		Entry("version set when state version is null -> counts as version change with channel",
			func(state, plan *ClusterRosaHcpState) {
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
})

var _ = Describe("Auto node role ARN validator", func() {
	DescribeTable("should validate correctly",
		func(request validator.StringRequest, expectedErr bool) {
			response := validator.StringResponse{}
			validateAutoNodeRoleARN().ValidateString(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("null -> ok",
			validator.StringRequest{
				Path:           path.Root("auto_node").AtName("role_arn"),
				PathExpression: path.MatchRoot("auto_node").AtName("role_arn"),
				ConfigValue:    types.StringNull(),
			},
			false,
		),
		Entry("unknown -> ok",
			validator.StringRequest{
				Path:           path.Root("auto_node").AtName("role_arn"),
				PathExpression: path.MatchRoot("auto_node").AtName("role_arn"),
				ConfigValue:    types.StringUnknown(),
			},
			false,
		),
		Entry("valid IAM role ARN -> ok",
			validator.StringRequest{
				Path:           path.Root("auto_node").AtName("role_arn"),
				PathExpression: path.MatchRoot("auto_node").AtName("role_arn"),
				ConfigValue:    types.StringValue(autoNodeRoleArn),
			},
			false,
		),
		Entry("empty string -> error",
			validator.StringRequest{
				Path:           path.Root("auto_node").AtName("role_arn"),
				PathExpression: path.MatchRoot("auto_node").AtName("role_arn"),
				ConfigValue:    types.StringValue(""),
			},
			true,
		),
		Entry("single quote -> error",
			validator.StringRequest{
				Path:           path.Root("auto_node").AtName("role_arn"),
				PathExpression: path.MatchRoot("auto_node").AtName("role_arn"),
				ConfigValue:    types.StringValue("arn:aws:iam::123456789012:role/karpenter'"),
			},
			true,
		),
		Entry("non-IAM ARN -> error",
			validator.StringRequest{
				Path:           path.Root("auto_node").AtName("role_arn"),
				PathExpression: path.MatchRoot("auto_node").AtName("role_arn"),
				ConfigValue:    types.StringValue("arn:aws:s3:::my-bucket"),
			},
			true,
		),
	)
})

var _ = Describe("Auto node helpers", func() {
	It("getAutoNodeMode returns null when auto_node is nil", func() {
		Expect(getAutoNodeMode(nil)).To(Equal(types.StringNull()))
	})

	It("getAutoNodeMode returns mode from auto_node", func() {
		autoNode := &AutoNode{Mode: types.StringValue(autoNodeModeEnabled)}
		Expect(getAutoNodeMode(autoNode)).To(Equal(types.StringValue(autoNodeModeEnabled)))
	})

	It("getAutoNodeRoleARN returns null when auto_node is nil", func() {
		Expect(getAutoNodeRoleARN(nil)).To(Equal(types.StringNull()))
	})

	It("getAutoNodeRoleARN returns role ARN from auto_node", func() {
		autoNode := &AutoNode{RoleARN: types.StringValue(autoNodeRoleArn)}
		Expect(getAutoNodeRoleARN(autoNode)).To(Equal(types.StringValue(autoNodeRoleArn)))
	})
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
