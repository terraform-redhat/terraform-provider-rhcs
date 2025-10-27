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
package breakglasscredential

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

type BreakGlassCredentialResource struct {
	collection    *cmv1.ClustersClient
	clusterClient common.ClusterClient
}

func New() resource.Resource {
	return &BreakGlassCredentialResource{}
}

var _ resource.Resource = &BreakGlassCredentialResource{}
var _ resource.ResourceWithConfigure = &BreakGlassCredentialResource{}
var _ resource.ResourceWithImportState = &BreakGlassCredentialResource{}

var expirationDurationValidator = attrvalidators.NewStringValidator("The expiration duration needs to be at least 10 minutes from now and to be at maximum 24 hours.",
	func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
			return
		}
		duration, err := time.ParseDuration(req.ConfigValue.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid configuration", err.Error())
			return
		}
		if duration < 10*time.Minute {
			resp.Diagnostics.AddError("Invalid configuration", "The expiration duration needs to be at least 10 minutes")
			return
		}
		if duration > 24*time.Hour {
			resp.Diagnostics.AddError("Invalid configuration", "The expiration duration needs to be at maximum 24 hours")
			return
		}
	})

func (b *BreakGlassCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_break_glass_credential"
}

func (b *BreakGlassCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Edit a cluster break glass credential",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
			},
			"username": schema.StringAttribute{
				Description: "User name of the break glass credential. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9-.]*$`), "The username '%s' must respect the regexp '^[a-zA-Z0-9-.]*$'"),
					// Maximum length for generated common name is 64 characters
					// 35 characters for "system:customer-break-glass:" + username
					stringvalidator.LengthAtMost(35),
				},
			},
			"expiration_duration": schema.StringAttribute{
				Description: "Expire the break glass credential after a relative duration like 2h, 8h." + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Validators: []validator.String{
					expirationDurationValidator,
				},
			},
			"id": schema.StringAttribute{
				Description: "Identifier of the break glass credential.",
				Optional:    true,
				Computed:    true,
			},
			"expiration_timestamp": schema.StringAttribute{
				Description: "Expiration timestamp of the break glass credential.",
				Optional:    true,
				Computed:    true,
			},
			"revocation_timestamp": schema.StringAttribute{
				Description: "Revocation timestamp of the break glass credential.",
				Optional:    true,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the break glass credential.",
				Optional:    true,
				Computed:    true,
			},
			"kubeconfig": schema.StringAttribute{
				Description: "Kubeconfig of the break glass credential.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (b *BreakGlassCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	b.collection = connection.ClustersMgmt().V1().Clusters()
	b.clusterClient = common.NewClusterClient(b.collection)
}

func (b *BreakGlassCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &BreakGlassCredential{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := plan.Cluster.ValueString()
	cluster, err := b.clusterClient.FetchCluster(ctx, clusterId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't check retrieve cluster",
			err.Error(),
		)
		return
	}
	if !cluster.Hypershift().Enabled() {
		resp.Diagnostics.AddError(
			"Unsupported Cluster Type",
			"Break glass credentials are only supported on Hosted Control Plane clusters",
		)
		return
	}
	if !cluster.ExternalAuthConfig().Enabled() {
		resp.Diagnostics.AddError(
			"External Authentication Configuration is not enabled",
			fmt.Sprintf("External Authentication Configuration is not enabled for cluster '%s'",
				clusterId),
		)
		return
	}

	err = b.createBreakGlassCredential(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed creating cluster break glass credential",
			fmt.Sprintf(
				"Failed creating break glass credential for cluster '%s': %v",
				clusterId, err,
			),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (b *BreakGlassCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &BreakGlassCredential{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := b.getBreakGlassCredential(state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed getting cluster break glass credential",
			fmt.Sprintf(
				"Failed getting break glass credential for cluster '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (b *BreakGlassCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fields := strings.Split(req.ID, ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Break glass credential to import should be specified as <cluster_id>,<break_glass_credential_id>",
		)
		return
	}
	clusterID := fields[0]
	breakGlassCredentialId := fields[1]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), breakGlassCredentialId)...)
}

func (b *BreakGlassCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &BreakGlassCredential{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"Cannot delete Break Glass Credential",
		fmt.Sprintf(
			"Cannot delete the break glass credential for cluster '%s'. "+
				"It is being removed from the Terraform state only. "+
				"To resume managing the break glass credential, import it again. ",
			state.Cluster.ValueString(),
		),
	)

	resp.State.RemoveResource(ctx)
}

func (b *BreakGlassCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Until we support. return an informative error
	resp.Diagnostics.AddError("Can't update Break Glass Credential", "Update is currently not supported.")
}

func (b *BreakGlassCredentialResource) getBreakGlassCredential(
	state *BreakGlassCredential) error {

	breakGlassCredential, err := b.collection.Cluster(state.Cluster.ValueString()).BreakGlassCredentials().BreakGlassCredential(state.Id.ValueString()).Get().Send()
	if err != nil {
		return err
	}

	b.populateState(breakGlassCredential.Body(), state)
	return nil
}

func (b *BreakGlassCredentialResource) createBreakGlassCredential(ctx context.Context,
	state *BreakGlassCredential) error {

	builder := cmv1.NewBreakGlassCredential()
	if !state.Username.IsNull() && !state.Username.IsUnknown() && state.Username.ValueString() != "" {
		builder.Username(state.Username.ValueString())
	}
	if !state.ExpirationDuration.IsNull() && !state.ExpirationDuration.IsUnknown() && state.ExpirationDuration.ValueString() != "" {
		expirationDuration, err := time.ParseDuration(state.ExpirationDuration.ValueString())
		if err != nil {
			return err
		}
		expirationTimestamp := time.Now().Add(expirationDuration).Round(time.Second)
		builder.ExpirationTimestamp(expirationTimestamp)
	}
	breakGlassCredential, err := builder.Build()
	if err != nil {
		return err
	}
	resp, err := b.collection.Cluster(state.Cluster.ValueString()).BreakGlassCredentials().Add().Body(breakGlassCredential).Send()
	if err != nil {
		return err
	}

	// Get the created break glass credential which includes the kubeconfig
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	pollResp, err := b.collection.Cluster(state.Cluster.ValueString()).BreakGlassCredentials().BreakGlassCredential(resp.Body().ID()).
		Poll().Interval(5 * time.Second).Predicate(
		func(response *cmv1.BreakGlassCredentialGetResponse) bool {
			return response.Body().Kubeconfig() != ""
		}).StartContext(pollCtx)
	if err != nil {
		return err
	}

	b.populateState(pollResp.Body(), state)
	return nil
}

func (b *BreakGlassCredentialResource) populateState(credential *cmv1.BreakGlassCredential, state *BreakGlassCredential) {
	if state == nil {
		state = &BreakGlassCredential{}
	}
	state.Id = types.StringValue(credential.ID())
	state.Username = types.StringValue(credential.Username())
	if !credential.ExpirationTimestamp().IsZero() {
		state.ExpirationTimestamp = types.StringValue(credential.ExpirationTimestamp().Format(time.RFC3339))
	} else {
		state.ExpirationTimestamp = types.StringNull()
	}
	if !credential.RevocationTimestamp().IsZero() {
		state.RevocationTimestamp = types.StringValue(credential.RevocationTimestamp().Format(time.RFC3339))
	} else {
		state.RevocationTimestamp = types.StringNull()
	}
	state.Status = types.StringValue(string(credential.Status()))
	state.Kubeconfig = types.StringValue(credential.Kubeconfig())
}
