package registry_config

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	dsschemadsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func RegistryConfigResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"registry_sources": schema.SingleNestedAttribute{
			Description: "registry_sources contains configuration that determines how the container runtime should treat individual registries when accessing images for builds+pods. (e.g. whether or not to allow insecure access).  It does not contain configuration for the internal cluster registry.",
			Attributes:  RegistrySourcesResource(),
			Optional:    true,
		},
		"allowed_registries_for_import": schema.ListNestedAttribute{
			Description: "allowed_registries_for_import limits the container image registries that normal users may import images from. Set this list to the registries that you trust to contain valid Docker images and that you want applications to be able to import from.",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: RegistryLocationResource(),
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"additional_trusted_ca": schema.MapAttribute{
			Description: "additional_trusted_ca is a map containing the registry hostname as the key, and the PEM-encoded certificate as the value, for each additional registry CA to trust.",
			ElementType: types.StringType,
			Optional:    true,
		},
		"platform_allowlist_id": schema.StringAttribute{
			Description: "platform_allowlist_id contains a reference to a RegistryAllowlist which is a list of internal registries which needs to be whitelisted for the platform to work. It can be omitted at creation and updating and its lifecycle can be managed separately if needed.",
			Optional:    true,
			Computed:    true,
		},
	}
}

func RegistryConfigDatasource() map[string]dsschemadsschema.Attribute {
	return map[string]dsschemadsschema.Attribute{
		"registry_sources": schema.SingleNestedAttribute{
			Description: "registry_sources contains configuration that determines how the container runtime should treat individual registries when accessing images for builds+pods. (e.g. whether or not to allow insecure access).  It does not contain configuration for the internal cluster registry.",
			Attributes:  RegistrySourcesResource(),
			Optional:    true,
		},
		"allowed_registries_for_import": schema.ListNestedAttribute{
			Description: "allowed_registries_for_import limits the container image registries that normal users may import images from. Set this list to the registries that you trust to contain valid Docker images and that you want applications to be able to import from.",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: RegistryLocationResource(),
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"additional_trusted_ca": schema.MapAttribute{
			Description: "additional_trusted_ca is a map containing the registry hostname as the key, and the PEM-encoded certificate as the value, for each additional registry CA to trust.",
			ElementType: types.StringType,
			Optional:    true,
		},
		"platform_allowlist_id": schema.StringAttribute{
			Description: "platform_allowlist_id contains a reference to a RegistryAllowlist which is a list of internal registries which needs to be whitelisted for the platform to work. It can be omitted at creation and updating and its lifecycle can be managed separately if needed.",
			Optional:    true,
			Computed:    true,
		},
	}
}

func RegistrySourcesResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allowed_registries": schema.ListAttribute{
			Description: "allowed_registries: registries for which image pull and push actions are allowed. To specify all subdomains, add the asterisk (*) wildcard character as a prefix to the domain name. For example, *.example.com. You can specify an individual repository within a registry. For example: reg1.io/myrepo/myapp:latest. All other registries are blocked. Mutually exclusive with `BlockedRegistries`",
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("blocked_registries")),
			},
		},
		"blocked_registries": schema.ListAttribute{
			Description: "blocked_registries: registries for which image pull and push actions are denied. To specify all subdomains, add the asterisk (*) wildcard character as a prefix to the domain name. For example, *.example.com. You can specify an individual repository within a registry. For example: reg1.io/myrepo/myapp:latest. All other registries are allowed. Mutually exclusive with `AllowedRegistries`",
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("allowed_registries")),
			},
		},
		"insecure_registries": schema.ListAttribute{
			Description: "insecure_registries are registries which do not have a valid TLS certificate or only support HTTP connections. To specify all subdomains, add the asterisk (*) wildcard character as a prefix to the domain name. For example, *.example.com. You can specify an individual repository within a registry. For example: reg1.io/myrepo/myapp:latest.",
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func RegistryLocationResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"domain_name": schema.StringAttribute{
			Description: "domain_name specifies a domain name for the registry",
			Optional:    true,
		},
		"insecure": schema.BoolAttribute{
			Description: "insecure indicates whether the registry is secure (https) or insecure (http). By default (if not specified) the registry is assumed as secure.",
			Optional:    true,
		},
	}
}
