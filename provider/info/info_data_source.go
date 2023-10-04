/*
Copyright (c) 2021 Red Hat, Inc.

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
package info

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

type OCMInfoDataSource struct {
	collection      *cmv1.CurrentAccountClient
	ocmAPI          string
	ocmAwsAccountID string
}

var _ datasource.DataSource = &OCMInfoDataSource{}
var _ datasource.DataSourceWithConfigure = &OCMInfoDataSource{}

func New() datasource.DataSource {
	return &OCMInfoDataSource{}
}

func (d *OCMInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_info"
}

func (d *OCMInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "OCM user account ID",
				Computed:    true,
			},
			"account_name": schema.StringAttribute{
				Description: "OCM account User full name",
				Computed:    true,
			},
			"account_username": schema.StringAttribute{
				Description: "OCM account username",
				Computed:    true,
			},
			"account_email": schema.StringAttribute{
				Description: "OCM account email",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "OCM account organization id",
				Computed:    true,
			},
			"organization_external_id": schema.StringAttribute{
				Description: "OCM account organization external id",
				Computed:    true,
			},
			"organization_name": schema.StringAttribute{
				Description: "OCM account organization name",
				Computed:    true,
			},
			"ocm_api": schema.StringAttribute{
				Description: "OCM API url",
				Computed:    true,
			},
			"ocm_aws_account_id": schema.StringAttribute{
				Description: "OCM AWS account ID",
				Computed:    true,
			},
		},
	}
}

func (d *OCMInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	d.collection = connection.AccountsMgmt().V1().CurrentAccount()
	d.ocmAPI = connection.URL()

}

func (d *OCMInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the state:
	state := &OCMInfoState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get account information
	accountResp, err := d.collection.Get().Send()
	if err != nil {
		resp.Diagnostics.AddError("Failed to connect to OCM", "Please verify your OCM offline token")
		return
	}
	obj, ok := accountResp.GetBody()
	if !ok {
		resp.Diagnostics.AddError("Parsing Error", "Failed to parse OCM response")
		return
	}

	state.AccountID = types.StringValue(obj.ID())
	state.AccountName = types.StringValue(fmt.Sprintf("%s %s", obj.FirstName(), obj.LastName()))
	state.AccountUsername = types.StringValue(obj.Username())
	state.AccountEmail = types.StringValue(obj.Email())
	state.OrganizationID = types.StringValue(obj.Organization().ID())
	state.OrganizationExternalID = types.StringValue(obj.Organization().ExternalID())
	state.OrganizationName = types.StringValue(obj.Organization().Name())

	state.OCMAPI = types.StringValue(d.ocmAPI)
	state.OCMAWSAccountID = types.StringValue(extractOCMAWSAccount(d.ocmAPI))

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func extractOCMAWSAccount(url string) string {
	env := ocmEnvProd
	if strings.Contains(url, "stage") {
		env = ocmEnvStage
	} else if strings.Contains(url, "integration") {
		env = ocmEnvInt
	}

	return ocmAWSAccounts[env]
}
