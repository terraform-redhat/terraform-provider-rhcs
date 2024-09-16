package registry_config

import (
	"context"
	"net/url"
	"path"

	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type RegistryConfig struct {
	RegistrySources            RegistrySources    `tfsdk:"registry_sources"`
	AllowedRegistriesForImport []RegistryLocation `tfsdk:"allowed_registries_for_import"`
	AdditionalTrustedCa        types.Map          `tfsdk:"additional_trusted_ca"`
	PlatformAllowlistId        types.String       `tfsdk:"platform_allowlist_id"`
}

type RegistrySources struct {
	AllowedRegistries  types.List `tfsdk:"allowed_registries"`
	BlockedRegistries  types.List `tfsdk:"blocked_registries"`
	InsecureRegistries types.List `tfsdk:"insecure_registries"`
}

type RegistryLocation struct {
	DomainName types.String `tfsdk:"domain_name"`
	Insecure   types.Bool   `tfsdk:"insecure"`
}

// CreateRegistryConfigBuilder creates a ClusterRegistryConfigBuilder from a Terraform state
func CreateRegistryConfigBuilder(ctx context.Context,
	state *RegistryConfig) (*cmv1.ClusterRegistryConfigBuilder, error) {
	if state == nil {
		return nil, nil
	}
	registryConfig := cmv1.NewClusterRegistryConfig()
	var registrySourcesBuilder *cmv1.RegistrySourcesBuilder
	hasAllowedRegistriesValue := common.HasValue(state.RegistrySources.AllowedRegistries)
	hasBlockedRegistriesValue := common.HasValue(state.RegistrySources.BlockedRegistries)
	hasInsecureRegistriesValue := common.HasValue(state.RegistrySources.InsecureRegistries)
	hasAllowedRegistriesForImportValue := checkAllowedRegistriesForImportValue(state)
	hasAdditionalTrustedCaValue := common.HasValue(state.AdditionalTrustedCa)

	if hasAllowedRegistriesValue || hasBlockedRegistriesValue ||
		hasInsecureRegistriesValue || hasAllowedRegistriesForImportValue || hasAdditionalTrustedCaValue {
		registrySourcesBuilder = cmv1.NewRegistrySources()
	}
	if hasAllowedRegistriesValue {
		err := fillAllowedRegistries(ctx, state.RegistrySources.AllowedRegistries, registrySourcesBuilder)
		if err != nil {
			return nil, err
		}
	}
	if hasBlockedRegistriesValue {
		err := fillBlockedRegistries(ctx, state.RegistrySources.BlockedRegistries, registrySourcesBuilder)
		if err != nil {
			return nil, err
		}
	}
	if hasInsecureRegistriesValue {
		err := fillInsecureRegistries(ctx, state.RegistrySources.InsecureRegistries, registrySourcesBuilder)
		if err != nil {
			return nil, err
		}
	}
	if hasAllowedRegistriesForImportValue {
		fillAllowedRegistriesForImport(state, registryConfig)
	}
	if hasAdditionalTrustedCaValue {
		err := fillAdditionalCa(ctx, state.AdditionalTrustedCa, registryConfig)
		if err != nil {
			return nil, err
		}
	}

	return registryConfig.RegistrySources(registrySourcesBuilder), nil
}

// UpdateRegistryConfigBuilder returns a ClusterRegistryConfigBuilder or an error starting from the state and a plan
func UpdateRegistryConfigBuilder(ctx context.Context, state *RegistryConfig,
	plan *RegistryConfig) (*cmv1.ClusterRegistryConfigBuilder, error) {
	registryConfig := cmv1.NewClusterRegistryConfig()
	var registrySourcesBuilder *cmv1.RegistrySourcesBuilder

	patchAllowedRegistries, isAllowedRegistriesChanged := shouldPatchAllowedRegistries(ctx, state,
		plan)
	patchBlockedRegistries, isBlockedRegistriesChanged := shouldPatchBlockedRegistries(state,
		plan)
	patchInsecureRegistries, isInsecureRegistriesChanged := shouldPatchInsecureRegistries(state,
		plan)
	isAllowedForImportChanged := shouldPatchRegistryAllowedForImport(state,
		plan)
	patchCa, isAdditionalCaChanged := shouldPatchAdditionalCa(state, plan)
	patchAllowlist, isAllowlistChanged := shouldPatchPlatformAllowlist(state, plan)

	if isAllowedRegistriesChanged || isBlockedRegistriesChanged || isInsecureRegistriesChanged {
		registrySourcesBuilder = cmv1.NewRegistrySources()
	}

	if isAllowedRegistriesChanged {
		err := fillAllowedRegistries(ctx, patchAllowedRegistries, registrySourcesBuilder)
		if err != nil {
			return nil, err
		}
	}
	if isBlockedRegistriesChanged {
		err := fillBlockedRegistries(ctx, patchBlockedRegistries, registrySourcesBuilder)
		if err != nil {
			return nil, err
		}
	}
	if isInsecureRegistriesChanged {
		err := fillInsecureRegistries(ctx, patchInsecureRegistries, registrySourcesBuilder)
		if err != nil {
			return nil, err
		}
	}
	if isAllowedForImportChanged {
		fillAllowedRegistriesForImport(plan, registryConfig)
	}
	if isAdditionalCaChanged {
		err := fillAdditionalCa(ctx, patchCa, registryConfig)
		if err != nil {
			return nil, err
		}
	}

	if isAllowlistChanged {
		registryConfig.PlatformAllowlist(cmv1.NewRegistryAllowlist().ID(patchAllowlist))
	}

	return registryConfig.RegistrySources(registrySourcesBuilder), nil
}

// PopulateRegistryConfigState takes a Cluster object from CS in input and fills a Terraform state with the data in it.
// Returns an error if there is a failure
func PopulateRegistryConfigState(inputCluster *cmv1.Cluster, state *RegistryConfig) error {
	if registryConfig, ok := inputCluster.GetRegistryConfig(); ok {
		if registrySources, ok := registryConfig.GetRegistrySources(); ok {
			if allowedRegistries, ok := registrySources.GetAllowedRegistries(); ok {
				listValue, err := common.StringArrayToList(allowedRegistries)
				if err != nil {
					return err
				}
				state.RegistrySources.AllowedRegistries = listValue
			}

			if blockedRegistries, ok := registrySources.GetBlockedRegistries(); ok {
				listValue, err := common.StringArrayToList(blockedRegistries)
				if err != nil {
					return err
				}
				state.RegistrySources.BlockedRegistries = listValue
			}

			if insecureRegistries, ok := registrySources.GetInsecureRegistries(); ok {
				listValue, err := common.StringArrayToList(insecureRegistries)
				if err != nil {
					return err
				}
				state.RegistrySources.InsecureRegistries = listValue
			}
		}

		if registryLocations, ok := registryConfig.GetAllowedRegistriesForImport(); ok {
			if state == nil {
				state = &RegistryConfig{}
			}
			state.AllowedRegistriesForImport = make([]RegistryLocation, len(registryLocations))
			for i, location := range registryLocations {
				domainName, ok := location.GetDomainName()
				if ok {
					state.AllowedRegistriesForImport[i] = RegistryLocation{
						DomainName: types.StringValue(domainName),
					}
					if insecure, ok := location.GetInsecure(); ok {
						state.AllowedRegistriesForImport[i].Insecure = types.BoolValue(insecure)
					} else {
						state.AllowedRegistriesForImport[i].Insecure = types.BoolValue(false)
					}
				}
			}
		} else {
			if state != nil {
				state.AllowedRegistriesForImport = nil
			}
		}

		if additionalTrustedCa, ok := registryConfig.GetAdditionalTrustedCa(); ok {
			mapValue, err := common.ConvertStringMapToMapType(additionalTrustedCa)
			if err != nil {
				return err
			}
			if state == nil {
				state = &RegistryConfig{}
			}
			state.AdditionalTrustedCa = mapValue
		}

		if platformAllowlist, ok := registryConfig.GetPlatformAllowlist(); ok {
			// try first directly the id, then href is id is not filled
			if id, ok := platformAllowlist.GetID(); ok {
				state.PlatformAllowlistId = types.StringValue(id)
			} else {
				href, ok := platformAllowlist.GetHREF()
				if ok {
					hrefUrl, err := url.Parse(href)
					if err != nil {
						return err
					}
					// Retrieve resource id from the full API href
					state.PlatformAllowlistId = types.StringValue(path.Base(hrefUrl.Path))
				}
			}
		}
	}
	return nil
}
