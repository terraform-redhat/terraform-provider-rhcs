package imagemirror

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ImageMirrorDataSource struct {
	clustersClient *cmv1.ClustersClient
}

var _ datasource.DataSourceWithConfigure = &ImageMirrorDataSource{}

func NewDataSource() datasource.DataSource {
	return &ImageMirrorDataSource{}
}

func (d *ImageMirrorDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_mirrors"
}

func (d *ImageMirrorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches image mirrors for a ROSA HCP cluster",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Description: "The ID of the ROSA cluster",
				Required:    true,
			},
			"image_mirrors": schema.ListNestedAttribute{
				Description: "List of image mirrors for the cluster",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the image mirror",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of mirror",
							Computed:    true,
						},
						"source": schema.StringAttribute{
							Description: "The source registry that will be mirrored",
							Computed:    true,
						},
						"mirrors": schema.ListAttribute{
							Description: "List of mirror registries",
							ElementType: types.StringType,
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *ImageMirrorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.clustersClient = connection.ClustersMgmt().V1().Clusters()
}

type ImageMirrorDataSourceModel struct {
	ClusterID    types.String       `tfsdk:"cluster_id"`
	ImageMirrors []ImageMirrorModel `tfsdk:"image_mirrors"`
}

type ImageMirrorModel struct {
	ID                  types.String `tfsdk:"id"`
	Type                types.String `tfsdk:"type"`
	Source              types.String `tfsdk:"source"`
	Mirrors             types.List   `tfsdk:"mirrors"`
	CreationTimestamp   types.String `tfsdk:"creation_timestamp"`
	LastUpdateTimestamp types.String `tfsdk:"last_update_timestamp"`
}

func (d *ImageMirrorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImageMirrorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := data.ClusterID.ValueString()

	// Get all image mirrors for the cluster
	response, err := d.clustersClient.Cluster(clusterId).ImageMirrors().List().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to List Image Mirrors",
			fmt.Sprintf("Could not list image mirrors for cluster '%s': %s", clusterId, err.Error()),
		)
		return
	}

	// Convert OCM image mirrors to our model
	imageMirrors := make([]ImageMirrorModel, 0, response.Size())
	response.Items().Each(func(imageMirror *cmv1.ImageMirror) bool {
		mirrors, diags := types.ListValueFrom(ctx, types.StringType, imageMirror.Mirrors())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return false
		}

		model := ImageMirrorModel{
			ID:      types.StringValue(imageMirror.ID()),
			Type:    types.StringValue(imageMirror.Type()),
			Source:  types.StringValue(imageMirror.Source()),
			Mirrors: mirrors,
		}

		if !imageMirror.CreationTimestamp().IsZero() {
			model.CreationTimestamp = types.StringValue(imageMirror.CreationTimestamp().Format("2006-01-02T15:04:05Z"))
		}
		if !imageMirror.LastUpdateTimestamp().IsZero() {
			model.LastUpdateTimestamp = types.StringValue(imageMirror.LastUpdateTimestamp().Format("2006-01-02T15:04:05Z"))
		}

		imageMirrors = append(imageMirrors, model)
		return true
	})

	data.ImageMirrors = imageMirrors

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
