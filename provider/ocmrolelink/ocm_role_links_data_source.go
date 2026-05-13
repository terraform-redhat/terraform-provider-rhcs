package ocmrolelink

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

type OCMRoleLinksDataSource struct {
	connection *sdk.Connection
}

var _ datasource.DataSource = &OCMRoleLinksDataSource{}
var _ datasource.DataSourceWithConfigure = &OCMRoleLinksDataSource{}

func NewDataSource() datasource.DataSource {
	return &OCMRoleLinksDataSource{}
}

func (d *OCMRoleLinksDataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ocm_role_links"
}

func (d *OCMRoleLinksDataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of linked OCM roles for an organization",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Description: "OCM organization ID (automatically determined from the current user)",
				Computed:    true,
			},
			"role_arns": schema.ListAttribute{
				Description: "List of linked AWS IAM role ARNs for the organization",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *OCMRoleLinksDataSource) Configure(ctx context.Context,
	req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)
		return
	}

	d.connection = connection
}

func (d *OCMRoleLinksDataSource) Read(ctx context.Context,
	req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OCMRoleLinksState
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Fetching current account organization")
	acctResp, err := d.connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get current account",
			fmt.Sprintf("Unable to retrieve current user's organization: %s", err.Error()),
		)
		return
	}
	orgID := acctResp.Body().Organization().ID()
	tflog.Debug(ctx, fmt.Sprintf("Using current user's organization: %s", orgID))

	tflog.Debug(ctx, fmt.Sprintf("Querying linked OCM roles for organization %s", orgID))

	labelsResp, err := d.connection.AccountsMgmt().V1().Organizations().
		Organization(orgID).
		Labels().
		Labels(OCMRoleLabel).
		Get().
		Send()

	if err != nil {
		if labelsResp != nil && labelsResp.Status() == http.StatusNotFound {
			tflog.Debug(ctx, fmt.Sprintf("No OCM role label found for organization %s", orgID))
			state.OrganizationID = types.StringValue(orgID)
			state.RoleArns = types.ListValueMust(types.StringType, []attr.Value{})
			diags = resp.State.Set(ctx, &state)
			resp.Diagnostics.Append(diags...)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OCM role links",
			fmt.Sprintf("Unable to retrieve OCM role links for organization %s: %s", orgID, err.Error()),
		)
		return
	}

	existingValue := labelsResp.Body().Value()
	rawArns := strings.Split(existingValue, ",")
	roleArns := make([]attr.Value, 0, len(rawArns))

	for _, arn := range rawArns {
		trimmed := strings.TrimSpace(arn)
		if trimmed != "" {
			roleArns = append(roleArns, types.StringValue(trimmed))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Found %d linked role(s) for organization %s", len(roleArns), orgID))

	state.OrganizationID = types.StringValue(orgID)
	state.RoleArns = types.ListValueMust(types.StringType, roleArns)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
