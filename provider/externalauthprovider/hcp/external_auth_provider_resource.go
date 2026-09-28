// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ExternalAuthProviderResource struct {
	clustersClient *cmv1.ClustersClient
}

type ExternalAuthProviderState struct {
	Cluster                   types.String `tfsdk:"cluster"`
	Name                      types.String `tfsdk:"name"`
	IssuerURL                 types.String `tfsdk:"issuer_url"`
	IssuerAudiences           types.Set    `tfsdk:"issuer_audiences"`
	IssuerCA                  types.String `tfsdk:"issuer_ca"`
	ClaimMappingGroupsClaim   types.String `tfsdk:"claim_mapping_groups_claim"`
	ClaimMappingUsernameClaim types.String `tfsdk:"claim_mapping_username_claim"`
	ClaimValidationRule       types.List   `tfsdk:"claim_validation_rule"`
	ConsoleClientID           types.String `tfsdk:"console_client_id"`
	ConsoleClientSecret       types.String `tfsdk:"console_client_secret"`
	ExtraScopes               types.Set    `tfsdk:"console_extra_scopes"`
	ID                        types.String `tfsdk:"id"`
}

var _ resource.Resource = &ExternalAuthProviderResource{}
var _ resource.ResourceWithValidateConfig = &ExternalAuthProviderResource{}
var _ resource.ResourceWithConfigure = &ExternalAuthProviderResource{}
var _ resource.ResourceWithImportState = &ExternalAuthProviderResource{}

func New() resource.Resource {
	return &ExternalAuthProviderResource{}
}

func (r *ExternalAuthProviderResource) Metadata(
	_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_external_auth_provider"
}

func (r *ExternalAuthProviderResource) Schema(
	_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "External authentication provider configuration for a ROSA HCP cluster.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "ID of the ROSA HCP cluster. The cluster must have external authentication enabled.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the external authentication provider; used as its OCM ID.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "name may not be empty/blank string"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"issuer_url": schema.StringAttribute{
				Description: "HTTPS URL of the token issuer.",
				Required:    true,
				Validators: []validator.String{
					issuerURLValidator(),
				},
			},
			"issuer_audiences": schema.SetAttribute{
				Description: "Audiences for which tokens from the issuer are valid (one to ten values).",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 10),
					setvalidator.NoNullValues(),
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "issuer audience may not be empty/blank string"),
					),
				},
			},
			"issuer_ca": schema.StringAttribute{
				Description: "PEM-encoded CA certificate bundle content for the issuer " +
					"(for example, file(\"ca.pem\")). Set an empty string to clear it; " +
					"When omitted, the current API value is recorded in state.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					issuerCAValidator(),
				},
			},
			"claim_mapping_groups_claim": schema.StringAttribute{
				Description: "ID token claim used to map groups. Set an empty string to clear it; " +
					"When omitted, the current API value is recorded in state.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^$|.*\S.*`),
						"groups claim may not be blank; use an empty string to clear it",
					),
				},
			},
			"claim_mapping_username_claim": schema.StringAttribute{
				Description: "ID token claim used to construct usernames. Set an empty string to clear it; " +
					"When omitted, the current API value is recorded in state.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^$|.*\S.*`),
						"username claim may not be blank; use an empty string to clear it",
					),
				},
			},
			"claim_validation_rule": schema.ListAttribute{
				Description: "Rules for token claims in claim:required_value format. " +
					"The first colon separates the claim; later colons belong to required_value. " +
					"Set an empty list to clear rules. When omitted, the current API value " +
					"is recorded in state.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.NoNullValues(),
					listvalidator.ValueStringsAre(claimValidationRuleValidator()),
				},
			},
			"console_client_id": schema.StringAttribute{
				Description: "Client ID used by the OpenShift console. " +
					"Must be set with console_client_secret and appear in issuer_audiences. " +
					"Set both client fields to empty strings to remove the client; " +
					"When omitted, the current API client ID is recorded in state.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^$|.*\S.*`),
						"console client ID may not be blank; use empty strings on both client fields to remove the client",
					),
				},
			},
			"console_client_secret": schema.StringAttribute{
				Description: "Client secret for the OpenShift console application. " +
					"Must be set with console_client_id. The API requires a non-empty secret " +
					"for an existing console client; set both client fields empty to " +
					"remove it. The API does not return the secret; import leaves it null in state.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^$|.*\S.*`),
						"console client secret may not be blank; use empty strings on both client fields to remove the client",
					),
				},
			},
			"console_extra_scopes": schema.SetAttribute{
				Description: "Additional scopes requested by the OpenShift console client. " +
					"Requires console_client_id and console_client_secret when configured. " +
					"When omitted, API scopes are recorded in state; an empty set clears them.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
					setvalidator.ValueStringsAre(scopeValidator()),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the external authentication provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ExternalAuthProviderResource) Configure(
	_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	r.clustersClient = connection.ClustersMgmt().V1().Clusters()
}

func (r *ExternalAuthProviderResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	clusterID, name, ok := parseExternalAuthImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Use <cluster_id>,<name> to import an external authentication provider.",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), name)...)
}

func parseExternalAuthImportID(id string) (clusterID, name string, ok bool) {
	parts := strings.Split(id, ",")
	if len(parts) != 2 {
		return "", "", false
	}
	clusterID, name = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	return clusterID, name, clusterID != "" && name != ""
}

const clusterIDLogKey = "cluster_id"

const (
	pendingDeletionRetryInterval = 5 * time.Second
	pendingDeletionRetryTimeout  = 30 * time.Second
	pendingDeletionMaxRetries    = 3
)

func safeAPIError(err error, secrets ...string) string {
	message := err.Error()
	for _, secret := range secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[REDACTED]")
		}
	}
	return message
}

func isPendingDeletionPatchResponse(response *cmv1.ExternalAuthUpdateResponse) bool {
	return response != nil && response.Status() == http.StatusBadRequest && response.Error() != nil &&
		strings.Contains(response.Error().Reason(), "external authentication pending deletion")
}

// retryPendingDeletionPatch returns after
// - a successful send
// - a response that is not HTTP 400 with a reason containing "external authentication pending deletion",
// - or a matching failure after maxRetries retries.
// maxRetries excludes the initial send.
// Between matching failures, it waits interval before retrying.
// If the retry context is done before a send, it returns nil and the context error;
// If the context ends while waiting, it returns the last response and the context error.
func retryPendingDeletionPatch(
	ctx context.Context,
	send func(context.Context) (*cmv1.ExternalAuthUpdateResponse, error),
	interval, timeout time.Duration,
	maxRetries int,
) (*cmv1.ExternalAuthUpdateResponse, error) {
	retryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for retries := 0; ; retries++ {
		if err := retryCtx.Err(); err != nil {
			return nil, fmt.Errorf("external authentication PATCH retry window ended: %w", err)
		}
		response, err := send(retryCtx)
		if err == nil || !isPendingDeletionPatchResponse(response) || retries >= maxRetries {
			return response, err
		}
		// OCM may reject PATCH while a delete/recreate ManifestWork is pending.
		select {
		case <-time.After(interval):
		case <-retryCtx.Done():
			return response, fmt.Errorf("external authentication PATCH retry window ended: %w", retryCtx.Err())
		}
	}
}

func (r *ExternalAuthProviderResource) checkCluster(ctx context.Context, clusterID string) error {
	response, err := r.clustersClient.Cluster(clusterID).Get().SendContext(ctx)
	if err != nil {
		return fmt.Errorf("cannot retrieve cluster %q: %w", clusterID, err)
	}
	if response == nil || response.Body() == nil {
		return fmt.Errorf("OCM returned no cluster details for %q", clusterID)
	}
	cluster := response.Body()
	if cluster.State() != cmv1.ClusterStateReady {
		return fmt.Errorf("cluster %q is not ready", clusterID)
	}
	if cluster.Hypershift() == nil || !cluster.Hypershift().Enabled() {
		return fmt.Errorf("external authentication providers require a ROSA Hosted Control Plane cluster")
	}
	if cluster.ExternalAuthConfig() == nil || !cluster.ExternalAuthConfig().Enabled() {
		return fmt.Errorf("external authentication configuration is not enabled for cluster %q", clusterID)
	}
	return nil
}

func (r *ExternalAuthProviderResource) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan, config ExternalAuthProviderState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	requestPlan := configuredPlan(plan, config)
	validateCrossFields(ctx, requestPlan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	clusterID, name := plan.Cluster.ValueString(), plan.Name.ValueString()
	if err := r.checkCluster(ctx, clusterID); err != nil {
		resp.Diagnostics.AddError("Cannot create external authentication provider", err.Error())
		return
	}
	body, err := buildCreate(ctx, requestPlan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid external authentication provider", err.Error())
		return
	}
	response, err := r.clustersClient.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().Add().
		Body(body).SendContext(ctx)
	if err != nil {
		if response != nil && response.Status() == http.StatusConflict {
			resp.Diagnostics.AddError(
				"External authentication provider already exists",
				fmt.Sprintf("Provider %q on cluster %q already exists; import it instead.", name, clusterID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Cannot create external authentication provider",
			fmt.Sprintf("OCM could not create provider %q on cluster %q: %s", name, clusterID,
				safeAPIError(err, plan.ConsoleClientSecret.ValueString())),
		)
		return
	}
	resp.Diagnostics.Append(populateFromAPI(ctx, &plan, response.Body())...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "Created external authentication provider", map[string]any{clusterIDLogKey: clusterID, "id": name})
}

func (r *ExternalAuthProviderResource) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ExternalAuthProviderState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clusterID, name := state.Cluster.ValueString(), state.Name.ValueString()
	response, err := r.clustersClient.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().
		ExternalAuth(name).Get().SendContext(ctx)
	if err != nil {
		if response != nil && response.Status() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Cannot read external authentication provider",
			fmt.Sprintf("OCM could not read provider %q on cluster %q: %s", name, clusterID,
				safeAPIError(err, state.ConsoleClientSecret.ValueString())),
		)
		return
	}
	if response.Body() == nil {
		resp.Diagnostics.AddError("Cannot read external authentication provider", "OCM returned no provider body.")
		return
	}
	resp.Diagnostics.Append(populateFromAPI(ctx, &state, response.Body())...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ExternalAuthProviderResource) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan, prior, config ExternalAuthProviderState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	requestPlan := configuredPlan(plan, config)
	validateCrossFields(ctx, requestPlan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	clusterID, name := plan.Cluster.ValueString(), plan.Name.ValueString()
	if err := r.checkCluster(ctx, clusterID); err != nil {
		resp.Diagnostics.AddError("Cannot update external authentication provider", err.Error())
		return
	}
	getResponse, err := r.clustersClient.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().
		ExternalAuth(name).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot update external authentication provider",
			fmt.Sprintf("OCM could not inspect provider %q on cluster %q: %s", name, clusterID,
				safeAPIError(err, prior.ConsoleClientSecret.ValueString())),
		)
		return
	}
	if getResponse.Body() == nil {
		resp.Diagnostics.AddError("Cannot update external authentication provider", "OCM returned no provider body.")
		return
	}
	remote := getResponse.Body()
	body, changed, err := buildPatch(ctx, requestPlan, prior, remote)
	if err != nil {
		resp.Diagnostics.AddError("Cannot safely update external authentication provider", err.Error())
		return
	}
	if changed {
		response, err := retryPendingDeletionPatch(ctx, func(ctx context.Context) (*cmv1.ExternalAuthUpdateResponse, error) {
			return r.clustersClient.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().
				ExternalAuth(name).Update().Body(body).SendContext(ctx)
		}, pendingDeletionRetryInterval, pendingDeletionRetryTimeout, pendingDeletionMaxRetries)
		if err != nil {
			resp.Diagnostics.AddError(
				"Cannot update external authentication provider",
				fmt.Sprintf("OCM could not update provider %q on cluster %q: %s", name, clusterID,
					safeAPIError(err, plan.ConsoleClientSecret.ValueString(), prior.ConsoleClientSecret.ValueString())),
			)
			return
		}
		remote = response.Body()
	}
	resp.Diagnostics.Append(populateFromAPI(ctx, &plan, remote)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = prior.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "Updated external authentication provider", map[string]any{clusterIDLogKey: clusterID, "id": name})
}

func (r *ExternalAuthProviderResource) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ExternalAuthProviderState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clusterID, name := state.Cluster.ValueString(), state.Name.ValueString()
	response, err := r.clustersClient.Cluster(clusterID).ExternalAuthConfig().ExternalAuths().
		ExternalAuth(name).Delete().SendContext(ctx)
	if err != nil {
		if response != nil && response.Status() == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Cannot delete external authentication provider",
			fmt.Sprintf("OCM could not delete provider %q on cluster %q: %s", name, clusterID,
				safeAPIError(err, state.ConsoleClientSecret.ValueString())),
		)
		return
	}
	tflog.Debug(ctx, "Deleted external authentication provider", map[string]any{clusterIDLogKey: clusterID, "id": name})
}
