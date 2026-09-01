// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package notificationcontacts

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"

	rosa "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type NotificationContactsResource struct {
	collection *sdk.Connection
}

type NotificationContactsState struct {
	ClusterID types.String `tfsdk:"cluster_id"`
	ID        types.String `tfsdk:"id"`
	Contacts  types.Set    `tfsdk:"contacts"`
}

func New() resource.Resource {
	return &NotificationContactsResource{}
}

var _ resource.Resource = &NotificationContactsResource{}
var _ resource.ResourceWithImportState = &NotificationContactsResource{}
var _ resource.ResourceWithConfigure = &NotificationContactsResource{}

func (r *NotificationContactsResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_cluster_notification_contacts"
}

func (r *NotificationContactsResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages notification contacts for a ROSA cluster. " +
			"Notification contacts are OCM account usernames that receive " +
			"cluster notification emails via the cluster's subscription.",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Description: "Identifier of the cluster." +
					" " + common.ValueCannotBeChangedStringDescription,
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the notification contacts resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"contacts": schema.SetAttribute{
				Description: "Set of OCM account usernames to receive cluster " +
					"notification emails. Values must be OCM usernames (not " +
					"email addresses). While email addresses are accepted by " +
					"the API, they are resolved to usernames internally, which " +
					"will cause persistent plan diffs and may lead to incorrect " +
					"contact removal on subsequent applies. All contacts must " +
					"belong to the same Red Hat organization as the cluster.",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^\S+$`),
							"notification contact must not contain whitespace",
						),
					),
				},
			},
		},
	}
}

func (r *NotificationContactsResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got: %T. "+
					"Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.collection = connection
}

func (r *NotificationContactsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	plan := &NotificationContactsState{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.ClusterID.ValueString()

	cluster, err := r.collection.ClustersMgmt().V1().
		Clusters().Cluster(clusterID).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				clusterID, err,
			),
		)
		return
	}

	subID, ok := rosa.GetSubscriptionID(cluster.Body())
	if !ok {
		resp.Diagnostics.AddError(
			"Can't set notification contacts",
			fmt.Sprintf(
				"Cluster '%s' does not have a subscription ID "+
					"available yet. Ensure the cluster has been "+
					"fully created before managing notification "+
					"contacts.", clusterID,
			),
		)
		return
	}

	var usernames []string
	diags = plan.Contacts.ElementsAs(ctx, &usernames, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf(
		"Setting %d notification contacts (subscription '%s')",
		len(usernames), subID,
	))

	err = rosa.UpdateNotificationContacts(
		ctx, r.collection, subID, usernames,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't set notification contacts",
			fmt.Sprintf(
				"Can't set notification contacts for cluster '%s': %v",
				clusterID, err,
			),
		)
		return
	}

	plan.ID = plan.ClusterID

	actualContacts, fetchErr := rosa.FetchNotificationContacts(
		ctx, r.collection, subID,
	)
	if fetchErr != nil {
		resp.Diagnostics.AddWarning(
			"Can't verify notification contacts",
			fmt.Sprintf(
				"Contacts were set but could not be verified: %v",
				fetchErr,
			),
		)
	} else {
		plan.Contacts = contactsToSet(actualContacts)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationContactsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	state := &NotificationContactsState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterID.ValueString()

	clusterResp, err := r.collection.ClustersMgmt().V1().
		Clusters().Cluster(clusterID).Get().SendContext(ctx)
	if err != nil {
		if clusterResp != nil &&
			clusterResp.Status() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				clusterID, err,
			),
		)
		return
	}

	subID, ok := rosa.GetSubscriptionID(clusterResp.Body())
	if !ok {
		resp.Diagnostics.AddError(
			"Can't read notification contacts",
			fmt.Sprintf(
				"Cluster '%s' subscription ID is not available.",
				clusterID,
			),
		)
		return
	}

	usernames, err := rosa.FetchNotificationContacts(
		ctx, r.collection, subID,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't read notification contacts",
			fmt.Sprintf(
				"Can't read notification contacts for cluster '%s': %v",
				clusterID, err,
			),
		)
		return
	}

	state.Contacts = contactsToSet(usernames)
	state.ID = state.ClusterID

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationContactsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	state := &NotificationContactsState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := &NotificationContactsState{}
	diags = req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.ClusterID.ValueString()

	cluster, err := r.collection.ClustersMgmt().V1().
		Clusters().Cluster(clusterID).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				clusterID, err,
			),
		)
		return
	}

	subID, ok := rosa.GetSubscriptionID(cluster.Body())
	if !ok {
		resp.Diagnostics.AddError(
			"Can't update notification contacts",
			fmt.Sprintf(
				"Cluster '%s' does not have a subscription ID available.",
				clusterID,
			),
		)
		return
	}

	var usernames []string
	diags = plan.Contacts.ElementsAs(ctx, &usernames, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf(
		"Updating to %d notification contacts (subscription '%s')",
		len(usernames), subID,
	))

	err = rosa.UpdateNotificationContacts(
		ctx, r.collection, subID, usernames,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't update notification contacts",
			fmt.Sprintf(
				"Can't update notification contacts for cluster '%s': %v",
				clusterID, err,
			),
		)
		actualContacts, fetchErr := rosa.FetchNotificationContacts(
			ctx, r.collection, subID,
		)
		if fetchErr != nil {
			diags = resp.State.Set(ctx, state)
			resp.Diagnostics.Append(diags...)
			return
		}
		plan.Contacts = contactsToSet(actualContacts)
		plan.ID = plan.ClusterID
		diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.ID = plan.ClusterID

	actualContacts, fetchErr := rosa.FetchNotificationContacts(
		ctx, r.collection, subID,
	)
	if fetchErr != nil {
		resp.Diagnostics.AddWarning(
			"Can't verify notification contacts",
			fmt.Sprintf(
				"Contacts were updated but could not be verified: %v",
				fetchErr,
			),
		)
	} else {
		plan.Contacts = contactsToSet(actualContacts)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationContactsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	state := &NotificationContactsState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterID.ValueString()

	clusterResp, err := r.collection.ClustersMgmt().V1().
		Clusters().Cluster(clusterID).Get().SendContext(ctx)
	if err != nil {
		if clusterResp != nil &&
			clusterResp.Status() == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster '%s' to remove "+
					"notification contacts: %v",
				clusterID, err,
			),
		)
		return
	}

	subID, ok := rosa.GetSubscriptionID(clusterResp.Body())
	if !ok {
		resp.Diagnostics.AddError(
			"Can't remove notification contacts",
			fmt.Sprintf(
				"Cluster '%s' subscription ID is not available.",
				clusterID,
			),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf(
		"Removing all notification contacts (subscription '%s')",
		subID,
	))

	err = rosa.UpdateNotificationContacts(
		ctx, r.collection, subID, []string{},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't remove notification contacts",
			fmt.Sprintf(
				"Can't remove notification contacts for cluster '%s': %v",
				clusterID, err,
			),
		)
	}
}

func (r *NotificationContactsResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(
		ctx, path.Root("cluster_id"), req, resp,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...,
	)
}

func contactsToSet(usernames []string) types.Set {
	if len(usernames) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, len(usernames))
	for i, u := range usernames {
		vals[i] = types.StringValue(u)
	}
	return types.SetValueMust(types.StringType, vals)
}
