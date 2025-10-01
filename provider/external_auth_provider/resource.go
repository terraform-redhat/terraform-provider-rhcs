package external_auth_provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithValidateConfig = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}

type Resource struct {
	collection *cmv1.ClustersClient
}

func New() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_auth_provider"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "External authentication provider for ROSA HCP clusters.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the external authentication provider.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "provider ID may not be empty/blank string"),
				},
			},
			"issuer": schema.SingleNestedAttribute{
				Description: "Token issuer configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL of the token issuer.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^https://.*`), "issuer URL must use HTTPS"),
						},
					},
					"audiences": schema.SetAttribute{
						Description: "List of audiences for the token issuer.",
						Required:    true,
						ElementType: types.StringType,
					},
					"ca": schema.StringAttribute{
						Description: "Certificate Authority (CA) certificate content.",
						Optional:    true,
					},
				},
			},
			"clients": schema.ListNestedAttribute{
				Description: "Client configurations for the external authentication provider.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"component": schema.SingleNestedAttribute{
							Description: "Component configuration.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Component name.",
									Optional:    true,
								},
								"namespace": schema.StringAttribute{
									Description: "Component namespace.",
									Optional:    true,
								},
							},
						},
						"id": schema.StringAttribute{
							Description: "Client identifier.",
							Optional:    true,
						},
						"secret": schema.StringAttribute{
							Description: "Client secret (required if client ID is provided).",
							Optional:    true,
							Sensitive:   true,
						},
						"extra_scopes": schema.SetAttribute{
							Description: "Additional OAuth scopes.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"type": schema.StringAttribute{
							Description: "Client type (confidential or public).",
							Computed:    true,
						},
					},
				},
			},
			"claim": schema.SingleNestedAttribute{
				Description: "Claim configuration for token validation and mapping.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"mappings": schema.SingleNestedAttribute{
						Description: "Token claim mappings.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"username": schema.SingleNestedAttribute{
								Description: "Username claim mapping.",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"claim": schema.StringAttribute{
										Description: "Token claim to extract username from.",
										Optional:    true,
									},
									"prefix": schema.StringAttribute{
										Description: "Prefix to apply to username.",
										Optional:    true,
									},
									"prefix_policy": schema.StringAttribute{
										Description: "Policy for applying the prefix.",
										Optional:    true,
									},
								},
							},
							"groups": schema.SingleNestedAttribute{
								Description: "Groups claim mapping.",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"claim": schema.StringAttribute{
										Description: "Token claim to extract groups from.",
										Optional:    true,
									},
									"prefix": schema.StringAttribute{
										Description: "Prefix to apply to group names.",
										Optional:    true,
									},
								},
							},
						},
					},
					"validation_rules": schema.ListNestedAttribute{
						Description: "Token claim validation rules.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"claim": schema.StringAttribute{
									Description: "Token claim to validate.",
									Required:    true,
								},
								"required_value": schema.StringAttribute{
									Description: "Required value for the claim.",
									Required:    true,
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(0),
						},
					},
				},
			},
		},
	}
}

func (r *Resource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		// Validation rules claim and required_value must be specified together
		resourcevalidator.RequiredTogether(
			path.MatchRoot("claim").AtName("validation_rules").AtAnyListIndex().AtName("claim"),
			path.MatchRoot("claim").AtName("validation_rules").AtAnyListIndex().AtName("required_value"),
		),
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	collection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.collection = collection.ClustersMgmt().V1().Clusters()
}

func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config State
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate client configuration - this requires iterating through dynamic lists
	if !config.Clients.IsNull() && !config.Clients.IsUnknown() {
		var clients []ExternalAuthClientConfig
		resp.Diagnostics.Append(config.Clients.ElementsAs(ctx, &clients, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for i, client := range clients {
			// If client ID is provided, secret must also be provided
			if !client.ID.IsNull() && !client.ID.IsUnknown() && client.ID.ValueString() != "" {
				if client.Secret.IsNull() || client.Secret.IsUnknown() || client.Secret.ValueString() == "" {
					resp.Diagnostics.AddAttributeError(
						path.Root("clients").AtListIndex(i).AtName("secret"),
						"Missing Required Field",
						"Client secret is required when client ID is provided.",
					)
				}
			}
		}
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Implementation will be added in OCM-18658
	resp.Diagnostics.AddError("Not Implemented", "Create operation will be implemented in OCM-18658")
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Implementation will be added in OCM-18658
	resp.Diagnostics.AddError("Not Implemented", "Read operation will be implemented in OCM-18658")
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implementation will be added in OCM-18658
	resp.Diagnostics.AddError("Not Implemented", "Update operation will be implemented in OCM-18658")
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Implementation will be added in OCM-18658
	resp.Diagnostics.AddError("Not Implemented", "Delete operation will be implemented in OCM-18658")
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Implementation will be added in OCM-18659
	resp.Diagnostics.AddError("Not Implemented", "Import functionality will be implemented in OCM-18659")
}
