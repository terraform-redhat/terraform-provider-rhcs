/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wifconfig

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// WifConfigDataSource implements the rhcs_wif_config data source.
type WifConfigDataSource struct {
	wifConfigs *cmv1.WifConfigsClient
}

var _ datasource.DataSource = &WifConfigDataSource{}
var _ datasource.DataSourceWithConfigure = &WifConfigDataSource{}

// NewDataSource creates a new WIF config data source.
func NewDataSource() datasource.DataSource {
	return &WifConfigDataSource{}
}

func (d *WifConfigDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_wif_config"
}

func (d *WifConfigDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing WIF config by display_name or id. " +
			"Use with the GCP WIF wiring HCL under examples/ to provision GCP IAM resources. " +
			"Display name is preferred since it is known at plan time without " +
			"passing IDs between Terraform configs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "WIF config ID. Specify to look up by ID. Exactly one of id or display_name is required.",
				Optional:    true,
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "WIF config display name. Specify to look up by name. Exactly one of id or display_name is required.",
				Optional:    true,
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "OCM organization ID owning this config.",
				Computed:    true,
			},
			"openshift_version": schema.StringAttribute{
				Description: "OpenShift version (x.y.z) used to scope WIF IAM resources.",
				Computed:    true,
			},
			"gcp": schema.SingleNestedAttribute{
				Description: "GCP-specific WIF configuration (blueprint from OCM).",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"project_id": schema.StringAttribute{
						Description: "GCP project ID where WIF resources are configured.",
						Computed:    true,
					},
					"project_number": schema.StringAttribute{
						Description: "GCP project number for WIF resources.",
						Computed:    true,
					},
					"role_prefix": schema.StringAttribute{
						Description: "Prefix for GCP custom role names.",
						Computed:    true,
					},
					"federated_project_id": schema.StringAttribute{
						Description: "GCP project ID where WorkloadIdentityPool resources are configured.",
						Computed:    true,
					},
					"federated_project_number": schema.StringAttribute{
						Description: "GCP project number for WorkloadIdentityPool.",
						Computed:    true,
					},
					"impersonator_email": schema.StringAttribute{
						Description: "Service account email used by OCM for impersonation.",
						Computed:    true,
					},
					"workload_identity_pool": schema.SingleNestedAttribute{
						Description: "Workload identity pool blueprint from OCM. Use with the GCP WIF wiring HCL under examples/.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"pool_id": schema.StringAttribute{
								Description: "GCP workload identity pool ID.",
								Computed:    true,
							},
							"identity_provider": schema.SingleNestedAttribute{
								Description: "OIDC identity provider configuration for the pool.",
								Computed:    true,
								Attributes: map[string]schema.Attribute{
									"identity_provider_id": schema.StringAttribute{
										Description: "Identity provider ID (e.g., oidc).",
										Computed:    true,
									},
									"issuer_url": schema.StringAttribute{
										Description: "OIDC issuer URL.",
										Computed:    true,
									},
									"jwks": schema.StringAttribute{
										Description: "JWKS JSON string for token validation.",
										Computed:    true,
									},
									"allowed_audiences": schema.ListAttribute{
										Description: "Allowed OIDC audiences.",
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
					"service_accounts": schema.ListNestedAttribute{
						Description: "Service accounts blueprint from OCM. Each entry defines a GCP SA with roles and access method.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"service_account_id": schema.StringAttribute{
									Description: "GCP service account ID.",
									Computed:    true,
								},
								"access_method": schema.StringAttribute{
									Description: "Access method: impersonate, wif, or vm.",
									Computed:    true,
								},
								"osd_role": schema.StringAttribute{
									Description: "OSD role name.",
									Computed:    true,
								},
								"roles": schema.ListNestedAttribute{
									Description: "IAM roles for this service account.",
									Computed:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"role_id": schema.StringAttribute{
												Description: "Role ID (predefined or custom).",
												Computed:    true,
											},
											"predefined": schema.BoolAttribute{
												Description: "True if this is a predefined GCP role.",
												Computed:    true,
											},
											"permissions": schema.ListAttribute{
												Description: "Permissions for custom roles.",
												Computed:    true,
												ElementType: types.StringType,
											},
											"resource_bindings": schema.ListNestedAttribute{
												Description: "Resource-level bindings (target SA for iam.serviceAccountUser etc.).",
												Computed:    true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"type": schema.StringAttribute{
															Description: "Resource type (e.g., iam.serviceAccounts).",
															Computed:    true,
														},
														"name": schema.StringAttribute{
															Description: "Target resource name.",
															Computed:    true,
														},
													},
												},
											},
										},
									},
								},
								"credential_request": schema.SingleNestedAttribute{
									Description: "OpenShift credential request (namespace and SA names for WIF principals).",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"namespace": schema.StringAttribute{
											Description: "OpenShift namespace for the credential secret.",
											Computed:    true,
										},
										"service_account_names": schema.ListAttribute{
											Description: "OpenShift service account names that can assume this GCP SA.",
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
							},
						},
					},
					"support": schema.SingleNestedAttribute{
						Description: "Support access configuration from OCM.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"principal": schema.StringAttribute{
								Description: "Support group principal (e.g., sd-sre-platform-gcp-access@redhat.com).",
								Computed:    true,
							},
							"roles": schema.ListNestedAttribute{
								Description: "Roles bound to the support group.",
								Computed:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"role_id": schema.StringAttribute{Description: "Role ID.", Computed: true},
										"predefined": schema.BoolAttribute{
											Description: "True if predefined.",
											Computed:    true,
										},
										"permissions": schema.ListAttribute{
											Description: "Custom role permissions.",
											Computed:    true,
											ElementType: types.StringType,
										},
										"resource_bindings": schema.ListNestedAttribute{
											Description: "Resource bindings.",
											Computed:    true,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"type": schema.StringAttribute{Computed: true},
													"name": schema.StringAttribute{Computed: true},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *WifConfigDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	d.wifConfigs = conn.ClustersMgmt().V1().GCP().WifConfigs()
}

func (d *WifConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WifConfigState
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && strings.TrimSpace(config.ID.ValueString()) != ""
	hasDisplayName := !config.DisplayName.IsNull() && strings.TrimSpace(config.DisplayName.ValueString()) != ""

	if hasID && hasDisplayName {
		resp.Diagnostics.AddError(
			"Invalid configuration",
			"Cannot specify both id and display_name. Specify exactly one.",
		)
		return
	}
	if !hasID && !hasDisplayName {
		resp.Diagnostics.AddError(
			"Invalid configuration",
			"Either id or display_name must be specified.",
		)
		return
	}

	var obj *cmv1.WifConfig

	if hasID {
		id := config.ID.ValueString()
		getResp, err := d.wifConfigs.WifConfig(id).Get().SendContext(ctx)
		if err != nil {
			if getResp != nil && getResp.Status() == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"WIF config not found",
					fmt.Sprintf("No WIF config found with id %q.", id),
				)
				return
			}
			resp.Diagnostics.AddError("failed to get WIF config", err.Error())
			return
		}
		obj = getResp.Body()
	} else {
		displayName := config.DisplayName.ValueString()
		searchExpr := fmt.Sprintf("display_name = '%s'", strings.ReplaceAll(displayName, "'", "''"))
		listResp, err := d.wifConfigs.List().Search(searchExpr).Size(2).SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to list WIF configs", err.Error())
			return
		}
		items := listResp.Items()
		if items == nil {
			resp.Diagnostics.AddError(
				"WIF config not found",
				fmt.Sprintf("No WIF config found with display_name %q.", displayName),
			)
			return
		}
		var count int
		items.Each(func(item *cmv1.WifConfig) bool {
			if count == 0 {
				obj = item
			}
			count++
			return true
		})
		if count == 0 {
			resp.Diagnostics.AddError(
				"WIF config not found",
				fmt.Sprintf("No WIF config found with display_name %q.", displayName),
			)
			return
		}
		if count > 1 {
			resp.Diagnostics.AddError(
				"Multiple WIF configs found",
				fmt.Sprintf("Found %d WIF configs with display_name %q. Use id for unique lookup.", count, displayName),
			)
			return
		}
	}

	state := WifConfigState{}
	populator := &WifConfigResource{}
	populator.populateState(obj, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
