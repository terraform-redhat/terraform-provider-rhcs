package external_auth_provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v1 "github.com/openshift-online/ocm-api-model/clientapi/clustersmgmt/v1"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithValidateConfig = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}

type Resource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
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

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.collection = connection.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection, connection)
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
	var plan State
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.Cluster.ValueString()
	providerID := plan.ID.ValueString()

	// Wait till the cluster is ready
	cluster, err := r.clusterWait.WaitForClusterToBeReady(ctx, clusterID, 60)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				clusterID, err,
			),
		)
		return
	}

	// Verify that external auth configuration is enabled on the cluster
	if cluster.ExternalAuthConfig() == nil || !cluster.ExternalAuthConfig().Enabled() {
		resp.Diagnostics.AddError(
			"External authentication not enabled",
			fmt.Sprintf(
				"External authentication configuration is not enabled for cluster '%s'. "+
					"Please enable external authentication on the cluster before creating external auth providers.",
				clusterID,
			),
		)
		return
	}

	// Build the external auth object
	externalAuth, err := r.buildExternalAuth(ctx, &plan, resp)
	if err != nil || resp.Diagnostics.HasError() {
		return
	}

	// Create the external auth provider
	addResp, err := r.collection.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().Add().
		Body(externalAuth).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create external auth provider",
			fmt.Sprintf("Cannot create external authentication provider '%s' for cluster '%s': %v", providerID, clusterID, err),
		)
		return
	}

	// Update state with the created resource
	err = r.populateState(ctx, addResp.Body(), &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to populate state",
			fmt.Sprintf("Cannot populate state after creating external auth provider: %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state State
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Cluster.ValueString()
	providerID := state.ID.ValueString()

	// Get the external auth provider
	getResp, err := r.collection.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().
		ExternalAuth(providerID).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read external auth provider",
			fmt.Sprintf("Cannot read external authentication provider '%s' for cluster '%s': %v", providerID, clusterID, err),
		)
		return
	}

	// Update state with the current resource
	err = r.populateState(ctx, getResp.Body(), &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to populate state",
			fmt.Sprintf("Cannot populate state after reading external auth provider: %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan State
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state State
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.Cluster.ValueString()
	providerID := plan.ID.ValueString()

	// Build the external auth object with updated configuration
	externalAuth, err := r.buildExternalAuth(ctx, &plan, resp)
	if err != nil || resp.Diagnostics.HasError() {
		return
	}

	// Update the external auth provider
	updateResp, err := r.collection.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().
		ExternalAuth(providerID).Update().Body(externalAuth).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update external auth provider",
			fmt.Sprintf("Cannot update external authentication provider '%s' for cluster '%s': %v", providerID, clusterID, err),
		)
		return
	}

	// Update state with the updated resource
	err = r.populateState(ctx, updateResp.Body(), &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to populate state",
			fmt.Sprintf("Cannot populate state after updating external auth provider: %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state State
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Cluster.ValueString()
	providerID := state.ID.ValueString()

	// Delete the external auth provider
	_, err := r.collection.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().
		ExternalAuth(providerID).Delete().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete external auth provider",
			fmt.Sprintf("Cannot delete external authentication provider '%s' for cluster '%s': %v", providerID, clusterID, err),
		)
		return
	}

	// Remove the resource from state
	resp.State.RemoveResource(ctx)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "begin importstate()")
	fields := strings.Split(req.ID, ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"External auth provider to import should be specified as <cluster_id>,<provider_id>",
		)
		return
	}
	clusterID := fields[0]
	providerID := fields[1]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), providerID)...)
}

func (r *Resource) populateState(ctx context.Context, externalAuth *cmv1.ExternalAuth, state *State) error {
	if state == nil {
		state = &State{}
	}

	// Set basic attributes
	state.ID = types.StringValue(externalAuth.ID())

	// Set issuer configuration
	if externalAuth.Issuer() != nil {
		issuer := &TokenIssuer{}
		issuer.URL = types.StringValue(externalAuth.Issuer().URL())

		if audiences := externalAuth.Issuer().Audiences(); len(audiences) > 0 {
			audienceValues := make([]types.String, len(audiences))
			for i, audience := range audiences {
				audienceValues[i] = types.StringValue(audience)
			}
			audienceSet, diag := types.SetValueFrom(ctx, types.StringType, audienceValues)
			if diag.HasError() {
				return fmt.Errorf("failed to populate audiences: %v", diag.Errors())
			}
			issuer.Audiences = audienceSet
		}

		if ca := externalAuth.Issuer().CA(); ca != "" {
			issuer.CA = types.StringValue(ca)
		} else {
			issuer.CA = types.StringNull()
		}

		state.Issuer = issuer
	}

	// Set clients configuration
	if clients := externalAuth.Clients(); len(clients) > 0 {
		clientConfigs := make([]ExternalAuthClientConfig, len(clients))
		for i, client := range clients {
			config := ExternalAuthClientConfig{}

			if client.Component() != nil {
				config.Component = &ClientComponent{
					Name:      types.StringValue(client.Component().Name()),
					Namespace: types.StringValue(client.Component().Namespace()),
				}
			}

			if id := client.ID(); id != "" {
				config.ID = types.StringValue(id)
			} else {
				config.ID = types.StringNull()
			}

			if secret := client.Secret(); secret != "" {
				config.Secret = types.StringValue(secret)
			} else {
				config.Secret = types.StringNull()
			}

			if scopes := client.ExtraScopes(); len(scopes) > 0 {
				scopeValues := make([]types.String, len(scopes))
				for j, scope := range scopes {
					scopeValues[j] = types.StringValue(scope)
				}
				scopeSet, diag := types.SetValueFrom(ctx, types.StringType, scopeValues)
				if diag.HasError() {
					return fmt.Errorf("failed to populate extra scopes: %v", diag.Errors())
				}
				config.ExtraScopes = scopeSet
			} else {
				config.ExtraScopes = types.SetNull(types.StringType)
			}

			// Set computed type based on whether client has ID/secret
			if !config.ID.IsNull() && !config.Secret.IsNull() {
				config.Type = types.StringValue("confidential")
			} else {
				config.Type = types.StringValue("public")
			}

			clientConfigs[i] = config
		}

		clientList, diag := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"component": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":      types.StringType,
						"namespace": types.StringType,
					},
				},
				"id":           types.StringType,
				"secret":       types.StringType,
				"extra_scopes": types.SetType{ElemType: types.StringType},
				"type":         types.StringType,
			},
		}, clientConfigs)
		if diag.HasError() {
			return fmt.Errorf("failed to populate clients: %v", diag.Errors())
		}
		state.Clients = clientList
	} else {
		state.Clients = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"component": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":      types.StringType,
						"namespace": types.StringType,
					},
				},
				"id":           types.StringType,
				"secret":       types.StringType,
				"extra_scopes": types.SetType{ElemType: types.StringType},
				"type":         types.StringType,
			},
		})
	}

	// Set claim configuration
	if externalAuth.Claim() != nil {
		claim := &ExternalAuthClaim{}

		if externalAuth.Claim().Mappings() != nil {
			mappings := &TokenClaimMappings{}

			if externalAuth.Claim().Mappings().UserName() != nil {
				username := &UsernameClaim{}
				if usernameClaim := externalAuth.Claim().Mappings().UserName().Claim(); usernameClaim != "" {
					username.Claim = types.StringValue(usernameClaim)
				} else {
					username.Claim = types.StringNull()
				}
				if prefix := externalAuth.Claim().Mappings().UserName().Prefix(); prefix != "" {
					username.Prefix = types.StringValue(prefix)
				} else {
					username.Prefix = types.StringNull()
				}
				if prefixPolicy := externalAuth.Claim().Mappings().UserName().PrefixPolicy(); prefixPolicy != "" {
					username.PrefixPolicy = types.StringValue(prefixPolicy)
				} else {
					username.PrefixPolicy = types.StringNull()
				}
				mappings.Username = username
			}

			if externalAuth.Claim().Mappings().Groups() != nil {
				groups := &GroupsClaim{}
				if groupsClaim := externalAuth.Claim().Mappings().Groups().Claim(); groupsClaim != "" {
					groups.Claim = types.StringValue(groupsClaim)
				} else {
					groups.Claim = types.StringNull()
				}
				if prefix := externalAuth.Claim().Mappings().Groups().Prefix(); prefix != "" {
					groups.Prefix = types.StringValue(prefix)
				} else {
					groups.Prefix = types.StringNull()
				}
				mappings.Groups = groups
			}

			claim.Mappings = mappings
		}

		if validationRules := externalAuth.Claim().ValidationRules(); len(validationRules) > 0 {
			rules := make([]TokenClaimValidationRule, len(validationRules))
			for i, rule := range validationRules {
				rules[i] = TokenClaimValidationRule{
					Claim:         types.StringValue(rule.Claim()),
					RequiredValue: types.StringValue(rule.RequiredValue()),
				}
			}

			rulesList, diag := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"claim":          types.StringType,
					"required_value": types.StringType,
				},
			}, rules)
			if diag.HasError() {
				return fmt.Errorf("failed to populate validation rules: %v", diag.Errors())
			}
			claim.ValidationRules = rulesList
		} else {
			claim.ValidationRules = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"claim":          types.StringType,
					"required_value": types.StringType,
				},
			})
		}

		state.Claim = claim
	}

	return nil
}

func (r *Resource) buildExternalAuth(ctx context.Context, plan *State, resp interface{}) (*cmv1.ExternalAuth, error) {
	// Build the external auth object
	externalAuthBuilder := cmv1.NewExternalAuth().ID(plan.ID.ValueString())

	// Set issuer configuration
	if plan.Issuer != nil {
		issuerBuilder := cmv1.NewTokenIssuer()

		if !plan.Issuer.URL.IsNull() {
			issuerBuilder.URL(plan.Issuer.URL.ValueString())
		}

		if !plan.Issuer.Audiences.IsNull() {
			var audiences []string
			if createResp, ok := resp.(*resource.CreateResponse); ok {
				createResp.Diagnostics.Append(plan.Issuer.Audiences.ElementsAs(ctx, &audiences, false)...)
				if createResp.Diagnostics.HasError() {
					return nil, fmt.Errorf("failed to process audiences")
				}
			} else if updateResp, ok := resp.(*resource.UpdateResponse); ok {
				updateResp.Diagnostics.Append(plan.Issuer.Audiences.ElementsAs(ctx, &audiences, false)...)
				if updateResp.Diagnostics.HasError() {
					return nil, fmt.Errorf("failed to process audiences")
				}
			}
			issuerBuilder.Audiences(audiences...)
		}

		if !plan.Issuer.CA.IsNull() {
			issuerBuilder.CA(plan.Issuer.CA.ValueString())
		}

		externalAuthBuilder.Issuer(issuerBuilder)
	}

	// Set clients configuration
	if !plan.Clients.IsNull() {
		var clients []ExternalAuthClientConfig
		if createResp, ok := resp.(*resource.CreateResponse); ok {
			createResp.Diagnostics.Append(plan.Clients.ElementsAs(ctx, &clients, false)...)
			if createResp.Diagnostics.HasError() {
				return nil, fmt.Errorf("failed to process clients")
			}
		} else if updateResp, ok := resp.(*resource.UpdateResponse); ok {
			updateResp.Diagnostics.Append(plan.Clients.ElementsAs(ctx, &clients, false)...)
			if updateResp.Diagnostics.HasError() {
				return nil, fmt.Errorf("failed to process clients")
			}
		}

		clientBuilders := make([]*v1.ExternalAuthClientConfigBuilder, 0)
		for _, client := range clients {
			clientBuilder := cmv1.NewExternalAuthClientConfig()

			if client.Component != nil {
				compBuilder := cmv1.NewClientComponent()
				if !client.Component.Name.IsNull() {
					compBuilder.Name(client.Component.Name.ValueString())
				}
				if !client.Component.Namespace.IsNull() {
					compBuilder.Namespace(client.Component.Namespace.ValueString())
				}
				clientBuilder.Component(compBuilder)
			}

			if !client.ID.IsNull() {
				clientBuilder.ID(client.ID.ValueString())
			}

			if !client.Secret.IsNull() {
				clientBuilder.Secret(client.Secret.ValueString())
			}

			if !client.ExtraScopes.IsNull() {
				var scopes []string
				if createResp, ok := resp.(*resource.CreateResponse); ok {
					createResp.Diagnostics.Append(client.ExtraScopes.ElementsAs(ctx, &scopes, false)...)
					if createResp.Diagnostics.HasError() {
						return nil, fmt.Errorf("failed to process extra scopes")
					}
				} else if updateResp, ok := resp.(*resource.UpdateResponse); ok {
					updateResp.Diagnostics.Append(client.ExtraScopes.ElementsAs(ctx, &scopes, false)...)
					if updateResp.Diagnostics.HasError() {
						return nil, fmt.Errorf("failed to process extra scopes")
					}
				}
				clientBuilder.ExtraScopes(scopes...)
			}
			clientBuilders = append(clientBuilders, clientBuilder)
		}
		externalAuthBuilder.Clients(clientBuilders...)
	}

	// Set claim configuration
	if plan.Claim != nil {
		claimBuilder := cmv1.NewExternalAuthClaim()

		if plan.Claim.Mappings != nil {
			mappingsBuilder := cmv1.NewTokenClaimMappings()

			if plan.Claim.Mappings.Username != nil {
				usernameBuilder := cmv1.NewUsernameClaim()
				if !plan.Claim.Mappings.Username.Claim.IsNull() {
					usernameBuilder.Claim(plan.Claim.Mappings.Username.Claim.ValueString())
				}
				if !plan.Claim.Mappings.Username.Prefix.IsNull() {
					usernameBuilder.Prefix(plan.Claim.Mappings.Username.Prefix.ValueString())
				}
				if !plan.Claim.Mappings.Username.PrefixPolicy.IsNull() {
					usernameBuilder.PrefixPolicy(plan.Claim.Mappings.Username.PrefixPolicy.ValueString())
				}
				mappingsBuilder.UserName(usernameBuilder)
			}

			if plan.Claim.Mappings.Groups != nil {
				groupsBuilder := cmv1.NewGroupsClaim()
				if !plan.Claim.Mappings.Groups.Claim.IsNull() {
					groupsBuilder.Claim(plan.Claim.Mappings.Groups.Claim.ValueString())
				}
				if !plan.Claim.Mappings.Groups.Prefix.IsNull() {
					groupsBuilder.Prefix(plan.Claim.Mappings.Groups.Prefix.ValueString())
				}
				mappingsBuilder.Groups(groupsBuilder)
			}
			claimBuilder.Mappings(mappingsBuilder)
		}

		if !plan.Claim.ValidationRules.IsNull() {
			var rules []TokenClaimValidationRule
			if createResp, ok := resp.(*resource.CreateResponse); ok {
				createResp.Diagnostics.Append(plan.Claim.ValidationRules.ElementsAs(ctx, &rules, false)...)
				if createResp.Diagnostics.HasError() {
					return nil, fmt.Errorf("failed to process validation rules")
				}
			} else if updateResp, ok := resp.(*resource.UpdateResponse); ok {
				updateResp.Diagnostics.Append(plan.Claim.ValidationRules.ElementsAs(ctx, &rules, false)...)
				if updateResp.Diagnostics.HasError() {
					return nil, fmt.Errorf("failed to process validation rules")
				}
			}

			for _, rule := range rules {
				ruleBuilder := cmv1.NewTokenClaimValidationRule()
				if !rule.Claim.IsNull() {
					ruleBuilder.Claim(rule.Claim.ValueString())
				}
				if !rule.RequiredValue.IsNull() {
					ruleBuilder.RequiredValue(rule.RequiredValue.ValueString())
				}
				claimBuilder.ValidationRules(ruleBuilder)
			}
		}
		externalAuthBuilder.Claim(claimBuilder)
	}

	// Build the external auth object
	externalAuth, err := externalAuthBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("cannot build external authentication: %v", err)
	}

	return externalAuth, nil
}
