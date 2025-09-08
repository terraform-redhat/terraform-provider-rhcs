package imagemirror

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

type ImageMirrorResource struct {
	clustersClient *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &ImageMirrorResource{}
var _ resource.ResourceWithImportState = &ImageMirrorResource{}

func New() resource.Resource {
	return &ImageMirrorResource{}
}

func (r *ImageMirrorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_mirror"
}

func (r *ImageMirrorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages image mirror configurations for ROSA HCP clusters",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Description: "The ID of the ROSA cluster",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of mirror (only 'digest' is currently supported)",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("digest"),
				Validators: []validator.String{
					attrvalidators.EnumValueValidator([]string{"digest"}),
				},
			},
			"source": schema.StringAttribute{
				Description: "The source registry that will be mirrored. Cannot be changed after creation",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mirrors": schema.ListAttribute{
				Description: "List of mirror registries that will serve content for the source",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the image mirror",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_timestamp": schema.StringAttribute{
				Description: "Timestamp when the image mirror was created",
				Computed:    true,
			},
			"last_update_timestamp": schema.StringAttribute{
				Description: "Timestamp when the image mirror was last updated",
				Computed:    true,
			},
		},
	}
}

func (r *ImageMirrorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.clustersClient = connection.ClustersMgmt().V1().Clusters()
}

func (r *ImageMirrorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ImageMirrorState
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := plan.ClusterID.ValueString()

	// Validate cluster state and type (consistent with ROSA CLI)
	cluster, err := r.clustersClient.Cluster(clusterId).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Get Cluster",
			fmt.Sprintf("Could not retrieve cluster '%s': %s", clusterId, err.Error()),
		)
		return
	}

	// Check if cluster is ready (consistent with ROSA CLI validation)
	if cluster.Body().State() != cmv1.ClusterStateReady {
		resp.Diagnostics.AddError(
			"Cluster Not Ready",
			fmt.Sprintf("Cluster '%s' is not ready", clusterId),
		)
		return
	}

	// Check if cluster is HCP (consistent with ROSA CLI validation)
	if cluster.Body().Hypershift().Enabled() == false {
		resp.Diagnostics.AddError(
			"Unsupported Cluster Type",
			"Image mirrors are only supported on Hosted Control Plane clusters",
		)
		return
	}

	// Convert mirrors list
	mirrors := make([]string, 0, len(plan.Mirrors.Elements()))
	diags = plan.Mirrors.ElementsAs(ctx, &mirrors, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the image mirror
	imageMirrorBuilder := cmv1.NewImageMirror().
		Type(plan.Type.ValueString()).
		Source(plan.Source.ValueString()).
		Mirrors(mirrors...)

	imageMirror, err := imageMirrorBuilder.Build()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Build Image Mirror",
			fmt.Sprintf("Could not build image mirror: %s", err.Error()),
		)
		return
	}

	// Create the image mirror
	response, err := r.clustersClient.Cluster(clusterId).ImageMirrors().Add().Body(imageMirror).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Image Mirror",
			fmt.Sprintf("Could not create image mirror: %s", err.Error()),
		)
		return
	}

	// Update state with response
	plan.ID = types.StringValue(response.Body().ID())
	if !response.Body().CreationTimestamp().IsZero() {
		plan.CreationTimestamp = types.StringValue(response.Body().CreationTimestamp().Format("2006-01-02T15:04:05Z"))
	}
	if !response.Body().LastUpdateTimestamp().IsZero() {
		plan.LastUpdateTimestamp = types.StringValue(response.Body().LastUpdateTimestamp().Format("2006-01-02T15:04:05Z"))
	}

	tflog.Debug(ctx, "Created image mirror", map[string]interface{}{
		"cluster_id": clusterId,
		"id":         plan.ID.ValueString(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ImageMirrorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ImageMirrorState
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.ClusterID.ValueString()
	imageMirrorId := state.ID.ValueString()

	// Get the image mirror
	response, err := r.clustersClient.Cluster(clusterId).ImageMirrors().ImageMirror(imageMirrorId).Get().SendContext(ctx)
	if err != nil {
		if response != nil && response.Status() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Read Image Mirror",
			fmt.Sprintf("Could not read image mirror '%s' for cluster '%s': %s", imageMirrorId, clusterId, err.Error()),
		)
		return
	}

	imageMirror := response.Body()

	// Update state with current values
	state.Type = types.StringValue(imageMirror.Type())
	state.Source = types.StringValue(imageMirror.Source())

	mirrors := make([]types.String, 0, len(imageMirror.Mirrors()))
	for _, mirror := range imageMirror.Mirrors() {
		mirrors = append(mirrors, types.StringValue(mirror))
	}
	var listDiags diag.Diagnostics
	state.Mirrors, listDiags = types.ListValueFrom(ctx, types.StringType, mirrors)
	resp.Diagnostics.Append(listDiags...)

	if !imageMirror.CreationTimestamp().IsZero() {
		state.CreationTimestamp = types.StringValue(imageMirror.CreationTimestamp().Format("2006-01-02T15:04:05Z"))
	}
	if !imageMirror.LastUpdateTimestamp().IsZero() {
		state.LastUpdateTimestamp = types.StringValue(imageMirror.LastUpdateTimestamp().Format("2006-01-02T15:04:05Z"))
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ImageMirrorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ImageMirrorState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := plan.ClusterID.ValueString()
	imageMirrorId := state.ID.ValueString()

	// Convert mirrors list
	mirrors := make([]string, 0, len(plan.Mirrors.Elements()))
	diags := plan.Mirrors.ElementsAs(ctx, &mirrors, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the update - only include type and mirrors fields as cluster_id and source require replacement
	imageMirrorBuilder := cmv1.NewImageMirror().
		Type(plan.Type.ValueString()).
		Mirrors(mirrors...)

	imageMirror, err := imageMirrorBuilder.Build()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Build Image Mirror Update",
			fmt.Sprintf("Could not build image mirror update: %s", err.Error()),
		)
		return
	}

	// Update the image mirror
	response, err := r.clustersClient.Cluster(clusterId).ImageMirrors().ImageMirror(imageMirrorId).Update().Body(imageMirror).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update Image Mirror",
			fmt.Sprintf("Could not update image mirror '%s': %s", imageMirrorId, err.Error()),
		)
		return
	}

	// Update state with response
	plan.ID = state.ID // Keep existing ID
	if !response.Body().CreationTimestamp().IsZero() {
		plan.CreationTimestamp = types.StringValue(response.Body().CreationTimestamp().Format("2006-01-02T15:04:05Z"))
	}
	if !response.Body().LastUpdateTimestamp().IsZero() {
		plan.LastUpdateTimestamp = types.StringValue(response.Body().LastUpdateTimestamp().Format("2006-01-02T15:04:05Z"))
	}

	tflog.Debug(ctx, "Updated image mirror", map[string]interface{}{
		"cluster_id": clusterId,
		"id":         plan.ID.ValueString(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ImageMirrorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ImageMirrorState
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.ClusterID.ValueString()
	imageMirrorId := state.ID.ValueString()

	// Delete the image mirror
	response, err := r.clustersClient.Cluster(clusterId).ImageMirrors().ImageMirror(imageMirrorId).Delete().SendContext(ctx)
	if err != nil {
		if response != nil && response.Status() == http.StatusNotFound {
			// Already deleted
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Delete Image Mirror",
			fmt.Sprintf("Could not delete image mirror '%s' for cluster '%s': %s", imageMirrorId, clusterId, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Deleted image mirror", map[string]interface{}{
		"cluster_id": clusterId,
		"id":         imageMirrorId,
	})
}

func (r *ImageMirrorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: cluster_id:image_mirror_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format 'cluster_id:image_mirror_id'",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
