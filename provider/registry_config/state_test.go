package registry_config

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var _ = Describe("Registry Config resource", func() {
	allowed1 := "allowed1.io"
	insecureRegistry := "insecure.io"
	registry1 := "registry1.io"
	pemValue := "PEM"
	Context("CreateRegistryConfigBuilder", func() {
		DescribeTable("should produce the right output",
			func(inputState *RegistryConfig, expectedConfig *cmv1.ClusterRegistryConfigBuilder, expectedErr error) {
				createdConfig, err := CreateRegistryConfigBuilder(context.TODO(), inputState)
				if expectedErr == nil {
					Expect(err).To(Not(HaveOccurred()))
					Expect(createdConfig).To(BeEquivalentTo(expectedConfig))
				} else {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(BeEquivalentTo(expectedErr.Error()))
				}
			},
			Entry("nil input -> empty config",
				nil,
				nil,
				nil,
			),
			Entry("allowed registries -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(allowed1)},
				},
				cmv1.NewClusterRegistryConfig().
					RegistrySources(
						cmv1.NewRegistrySources().AllowedRegistries(allowed1)),
				nil,
			),
			Entry("allowed and insecure registries -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{
						AllowedRegistries:  getListTypeValue(allowed1),
						InsecureRegistries: getListTypeValue(insecureRegistry)},
				},
				cmv1.NewClusterRegistryConfig().
					RegistrySources(
						cmv1.NewRegistrySources().AllowedRegistries(allowed1).InsecureRegistries(insecureRegistry)),
				nil,
			),
			Entry("blocked and allowed registries for import -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(registry1),
					},
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(true)},
					},
				},
				cmv1.NewClusterRegistryConfig().
					RegistrySources(
						cmv1.NewRegistrySources().BlockedRegistries(registry1)).
					AllowedRegistriesForImport(
						cmv1.NewRegistryLocation().DomainName(allowed1).Insecure(false),
						cmv1.NewRegistryLocation().DomainName(registry1).Insecure(true)),
				nil,
			),
			Entry("blocked, allowed registries for import and additional CA -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(registry1),
					},
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(true)},
					},
					AdditionalTrustedCa: getMapTypeValue(registry1, pemValue),
				},
				cmv1.NewClusterRegistryConfig().
					RegistrySources(
						cmv1.NewRegistrySources().BlockedRegistries(registry1)).
					AdditionalTrustedCa(
						map[string]string{registry1: pemValue}).
					AllowedRegistriesForImport(
						cmv1.NewRegistryLocation().DomainName(allowed1).Insecure(false),
						cmv1.NewRegistryLocation().DomainName(registry1).Insecure(true)),
				nil,
			),
		)
	})
	Context("UpdateRegistryConfigBuilder", func() {
		DescribeTable("should produce the right output",
			func(inputState, inputPlan *RegistryConfig,
				expectedConfig *cmv1.ClusterRegistryConfigBuilder, expectedErr error) {
				createdConfig, err := UpdateRegistryConfigBuilder(context.TODO(), inputState, inputPlan)
				if expectedErr == nil {
					Expect(err).To(Not(HaveOccurred()))
					Expect(createdConfig).To(BeEquivalentTo(expectedConfig))
				} else {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(BeEquivalentTo(expectedErr.Error()))
				}
			},
			Entry("nil input -> empty config",
				nil, nil,
				cmv1.NewClusterRegistryConfig(),
				nil,
			),
			Entry("no change between state and plan -> empty config",
				&RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(allowed1)},
				}, &RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(allowed1)},
				},
				cmv1.NewClusterRegistryConfig(),
				nil,
			),
			Entry("change between nil state and plan -> reflected config",
				nil,
				&RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(allowed1)},
				},
				cmv1.NewClusterRegistryConfig().RegistrySources(cmv1.NewRegistrySources().AllowedRegistries(allowed1)),
				nil,
			),
			Entry("change allowed registries between state and plan -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(allowed1)},
				}, &RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(registry1)},
				},
				cmv1.NewClusterRegistryConfig().RegistrySources(cmv1.NewRegistrySources().AllowedRegistries(registry1)),
				nil,
			),
			Entry("change allowed and insecure registries between state and plan -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(allowed1)},
				}, &RegistryConfig{
					RegistrySources: RegistrySources{
						AllowedRegistries:  getListTypeValue(registry1),
						InsecureRegistries: getListTypeValue(insecureRegistry)},
				},
				cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().AllowedRegistries(registry1).InsecureRegistries(insecureRegistry)),
				nil,
			),
			Entry("change blocked and insecure registries between state and plan -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{AllowedRegistries: getListTypeValue(allowed1)},
				}, &RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries:  getListTypeValue(registry1),
						InsecureRegistries: getListTypeValue(insecureRegistry)},
				},
				cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().BlockedRegistries(registry1).
						InsecureRegistries(insecureRegistry).AllowedRegistries()),
				nil,
			),
			Entry("change blocked and AllowedRegistriesForImport registries between state and plan -> reflected in config",
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(allowed1),
					},
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(true)},
					},
				},
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(registry1),
					},
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(true)},
					},
				},
				cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().BlockedRegistries(registry1)).AllowedRegistriesForImport(
					cmv1.NewRegistryLocation().DomainName(registry1).Insecure(true)),
				nil,
			),
			Entry("change additionalCA and AllowedRegistriesForImport registries between state and plan -> reflected in config",
				&RegistryConfig{
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(true)},
					},
				},
				&RegistryConfig{
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
					},
					AdditionalTrustedCa: getMapTypeValue(registry1, pemValue),
				},
				cmv1.NewClusterRegistryConfig().
					AdditionalTrustedCa(
						map[string]string{registry1: pemValue}).
					AllowedRegistriesForImport(
						cmv1.NewRegistryLocation().DomainName(allowed1).Insecure(false)),
				nil,
			),
			Entry("nil state and additionalCA and AllowedRegistriesForImport in plan -> reflected in config",
				nil,
				&RegistryConfig{
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
					},
					AdditionalTrustedCa: getMapTypeValue(registry1, pemValue),
				},
				cmv1.NewClusterRegistryConfig().
					AdditionalTrustedCa(
						map[string]string{registry1: pemValue}).
					AllowedRegistriesForImport(
						cmv1.NewRegistryLocation().DomainName(allowed1).Insecure(false)),
				nil,
			),
			Entry("change AllowedRegistriesForImport entries between state and plan -> reflected in config",
				&RegistryConfig{
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(false)},
					},
				},
				&RegistryConfig{
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1)},
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(true)},
					},
				},
				cmv1.NewClusterRegistryConfig().
					AllowedRegistriesForImport(
						cmv1.NewRegistryLocation().DomainName(allowed1).Insecure(false),
						cmv1.NewRegistryLocation().DomainName(registry1).Insecure(true),
					),
				nil,
			),
			Entry("change allowlist between state and plan -> reflected in config",
				&RegistryConfig{
					PlatformAllowlistId: types.StringValue("id1"),
				},
				&RegistryConfig{
					PlatformAllowlistId: types.StringValue("id2"),
				},
				cmv1.NewClusterRegistryConfig().
					PlatformAllowlist(cmv1.NewRegistryAllowlist().ID("id2")),
				nil,
			),
		)
	})
	Context("PopulateRegistryConfigState", func() {
		DescribeTable("should produce the right output",
			func(inputCluster *cmv1.Cluster, expectedState *RegistryConfig, expectedErr error) {
				state := &RegistryConfig{}
				err := PopulateRegistryConfigState(inputCluster, state)
				if expectedErr == nil {
					Expect(err).To(Not(HaveOccurred()))
					Expect(state).To(BeEquivalentTo(expectedState))
				} else {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(BeEquivalentTo(expectedErr.Error()))
				}
			},
			Entry("nil input -> empty config",
				buildClusterWithRegistryConfig(nil),
				&RegistryConfig{},
				nil,
			),
			Entry("allowed registries in input -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().AllowedRegistries(allowed1))),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						AllowedRegistries: getListTypeValue(allowed1),
					},
				},
				nil,
			),
			Entry("allowed and insecure registries in input -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().AllowedRegistries(allowed1).
						InsecureRegistries(insecureRegistry))),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						AllowedRegistries:  getListTypeValue(allowed1),
						InsecureRegistries: getListTypeValue(insecureRegistry),
					},
				},
				nil,
			),
			Entry("blocked and insecure registries in input -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().BlockedRegistries(allowed1).
						InsecureRegistries(insecureRegistry))),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries:  getListTypeValue(allowed1),
						InsecureRegistries: getListTypeValue(insecureRegistry),
					},
				},
				nil,
			),
			Entry("blocked and allowed registries for import -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().BlockedRegistries(allowed1)).AllowedRegistriesForImport(
					cmv1.NewRegistryLocation().DomainName(allowed1)),
				),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(allowed1),
					},
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1), Insecure: types.BoolValue(false)},
					},
				},
				nil,
			),
			Entry("blocked and multiple allowed registries for import -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().BlockedRegistries(allowed1)).AllowedRegistriesForImport(
					cmv1.NewRegistryLocation().DomainName(allowed1),
					cmv1.NewRegistryLocation().DomainName(registry1).Insecure(true)),
				),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(allowed1),
					},
					AllowedRegistriesForImport: []RegistryLocation{
						{DomainName: types.StringValue(allowed1), Insecure: types.BoolValue(false)},
						{DomainName: types.StringValue(registry1), Insecure: types.BoolValue(true)},
					},
				},
				nil,
			),
			Entry("blocked registries and additional CA in input -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().RegistrySources(
					cmv1.NewRegistrySources().BlockedRegistries(allowed1)).AdditionalTrustedCa(
					map[string]string{registry1: pemValue}),
				),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(allowed1),
					},
					AdditionalTrustedCa: getMapTypeValue(registry1, pemValue),
				},
				nil,
			),
			Entry("allowlist in input with only id -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().
					RegistrySources(
						cmv1.NewRegistrySources().BlockedRegistries(allowed1)).
					PlatformAllowlist(
						cmv1.NewRegistryAllowlist().ID("id1"))),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(allowed1),
					},
					PlatformAllowlistId: types.StringValue("id1"),
				},
				nil,
			),
			Entry("allowlist in input with only href -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().
					RegistrySources(
						cmv1.NewRegistrySources().BlockedRegistries(allowed1)).
					PlatformAllowlist(
						cmv1.NewRegistryAllowlist().
							HREF("/api/clusters_mgmt/v1/registry_allowlists/id1"))),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(allowed1),
					},
					PlatformAllowlistId: types.StringValue("id1"),
				},
				nil,
			),
			Entry("allowlist in input with id and href -> in output too",
				buildClusterWithRegistryConfig(cmv1.NewClusterRegistryConfig().
					RegistrySources(
						cmv1.NewRegistrySources().BlockedRegistries(allowed1)).
					PlatformAllowlist(
						cmv1.NewRegistryAllowlist().
							HREF("/api/clusters_mgmt/v1/registry_allowlists/id1").ID("id1"))),
				&RegistryConfig{
					RegistrySources: RegistrySources{
						BlockedRegistries: getListTypeValue(allowed1),
					},
					PlatformAllowlistId: types.StringValue("id1"),
				},
				nil,
			),
		)
	})
})

func buildClusterWithRegistryConfig(config *cmv1.ClusterRegistryConfigBuilder) *cmv1.Cluster {
	cluster, err := cmv1.NewCluster().ID("clusterId").RegistryConfig(config).Build()
	Expect(err).To(BeNil())
	return cluster
}

func getListTypeValue(value string) types.List {
	list, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue(value),
	})
	Expect(diags).To(BeNil())
	return list
}

func getMapTypeValue(key, value string) types.Map {
	mapValue, diags := types.MapValue(types.StringType, map[string]attr.Value{
		key: types.StringValue(value),
	})
	Expect(diags).To(BeNil())
	return mapValue
}
