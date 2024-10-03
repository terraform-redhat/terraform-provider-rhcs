package registry_config

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

func shouldPatchAllowedRegistries(ctx context.Context, state, plan *RegistryConfig) (types.List, bool) {
	if state == nil && plan == nil {
		return types.List{}, false
	}
	var allowedRegistriesState types.List
	if state != nil {
		allowedRegistriesState = state.RegistrySources.AllowedRegistries
	}
	var allowedRegistriesPlan types.List
	if plan != nil {
		allowedRegistriesPlan = plan.RegistrySources.AllowedRegistries
	}
	return common.ShouldPatchList(allowedRegistriesState, allowedRegistriesPlan)
}

func shouldPatchBlockedRegistries(state, plan *RegistryConfig) (types.List, bool) {
	if state == nil && plan == nil {
		return types.List{}, false
	}
	var blockedRegistriesState types.List
	if state != nil {
		blockedRegistriesState = state.RegistrySources.BlockedRegistries
	}
	var blockedRegistriesPlan types.List
	if plan != nil {
		blockedRegistriesPlan = plan.RegistrySources.BlockedRegistries
	}
	return common.ShouldPatchList(blockedRegistriesState, blockedRegistriesPlan)
}

func shouldPatchInsecureRegistries(state, plan *RegistryConfig) (types.List, bool) {
	if state == nil && plan == nil {
		return types.List{}, false
	}
	var insecureRegistriesState types.List
	if state != nil {
		insecureRegistriesState = state.RegistrySources.InsecureRegistries
	}
	var insecureRegistriesPlan types.List
	if plan != nil {
		insecureRegistriesPlan = plan.RegistrySources.InsecureRegistries
	}
	return common.ShouldPatchList(insecureRegistriesState, insecureRegistriesPlan)
}

func shouldPatchPlatformAllowlist(state, plan *RegistryConfig) (string, bool) {
	if state == nil && plan == nil {
		return "", false
	}
	var allowlistState types.String
	if state != nil {
		allowlistState = state.PlatformAllowlistId
	}
	var allowlistPlan types.String
	if plan != nil {
		allowlistPlan = plan.PlatformAllowlistId
	}
	return common.ShouldPatchString(allowlistState, allowlistPlan)
}

func fillAllowedRegistries(ctx context.Context, stateInput types.List, builder *cmv1.RegistrySourcesBuilder) error {
	allowedRegistries, err := common.StringListToArray(ctx, stateInput)
	if err != nil {
		return err
	}
	builder.AllowedRegistries(allowedRegistries...)
	return nil
}

func fillBlockedRegistries(ctx context.Context, stateInput types.List, builder *cmv1.RegistrySourcesBuilder) error {
	blockedRegistries, err := common.StringListToArray(ctx, stateInput)
	if err != nil {
		return err
	}
	builder.BlockedRegistries(blockedRegistries...)
	return nil
}

func fillInsecureRegistries(ctx context.Context, stateInput types.List, builder *cmv1.RegistrySourcesBuilder) error {
	insecureRegistries, err := common.StringListToArray(ctx, stateInput)
	if err != nil {
		return err
	}
	if insecureRegistries == nil {
		return nil
	}
	builder.InsecureRegistries(insecureRegistries...)
	return nil
}

func fillAllowedRegistriesForImport(plan *RegistryConfig, registryConfig *cmv1.ClusterRegistryConfigBuilder) {
	var locationBuilder []*cmv1.RegistryLocationBuilder
	for _, location := range plan.AllowedRegistriesForImport {
		locationBuilder = append(locationBuilder, cmv1.NewRegistryLocation().
			DomainName(location.DomainName.ValueString()).
			Insecure(location.Insecure.ValueBool()))
	}
	registryConfig.AllowedRegistriesForImport(locationBuilder...)
}

func fillAdditionalCa(ctx context.Context, stateInput types.Map,
	registryConfig *cmv1.ClusterRegistryConfigBuilder) error {
	additionalCa := map[string]string{}
	elements, err := common.OptionalMap(ctx, stateInput)
	if err != nil {
		return err
	}
	for k, v := range elements {
		additionalCa[k] = v
	}
	registryConfig.AdditionalTrustedCa(additionalCa)
	return nil
}

func checkAllowedRegistriesForImportValue(state *RegistryConfig) bool {
	return state.AllowedRegistriesForImport != nil
}

func shouldPatchRegistryAllowedForImport(state, plan *RegistryConfig) bool {
	if state == nil && plan == nil {
		return false
	}

	var locationsState []RegistryLocation
	if state != nil {
		locationsState = state.AllowedRegistriesForImport
	}
	var locationsPlan []RegistryLocation
	if plan != nil {
		locationsPlan = plan.AllowedRegistriesForImport
	}

	if (locationsState == nil && locationsPlan != nil) || (locationsState != nil && locationsPlan == nil) {
		return true
	}
	if len(locationsState) != len(locationsPlan) {
		return true
	}
	for i := range locationsState {
		if !locationsState[i].DomainName.Equal(locationsPlan[i].DomainName) ||
			!locationsState[i].Insecure.Equal(locationsPlan[i].Insecure) {
			return true
		}
	}
	return false
}

func shouldPatchAdditionalCa(state, plan *RegistryConfig) (types.Map, bool) {
	if state == nil && plan == nil {
		return types.Map{}, false
	}
	var additionalCaState types.Map
	if state != nil {
		additionalCaState = state.AdditionalTrustedCa
	}
	var additionalCaPlan types.Map
	if plan != nil {
		additionalCaPlan = plan.AdditionalTrustedCa
	}
	return common.ShouldPatchMap(additionalCaState, additionalCaPlan)
}
