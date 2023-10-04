/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"context"
***REMOVED***
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
***REMOVED***

type OCMInfoDataSource struct {
	collection      *cmv1.CurrentAccountClient
	ocmAPI          string
	ocmAwsAccountID string
}

var _ datasource.DataSource = &OCMInfoDataSource{}
var _ datasource.DataSourceWithConfigure = &OCMInfoDataSource{}

func New(***REMOVED*** datasource.DataSource {
	return &OCMInfoDataSource{}
}

func (d *OCMInfoDataSource***REMOVED*** Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_info"
}

func (d *OCMInfoDataSource***REMOVED*** Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "OCM user account ID",
				Computed:    true,
	***REMOVED***,
			"account_name": schema.StringAttribute{
				Description: "OCM account User full name",
				Computed:    true,
	***REMOVED***,
			"account_username": schema.StringAttribute{
				Description: "OCM account username",
				Computed:    true,
	***REMOVED***,
			"account_email": schema.StringAttribute{
				Description: "OCM account email",
				Computed:    true,
	***REMOVED***,
			"organization_id": schema.StringAttribute{
				Description: "OCM account organization id",
				Computed:    true,
	***REMOVED***,
			"organization_external_id": schema.StringAttribute{
				Description: "OCM account organization external id",
				Computed:    true,
	***REMOVED***,
			"organization_name": schema.StringAttribute{
				Description: "OCM account organization name",
				Computed:    true,
	***REMOVED***,
			"ocm_api": schema.StringAttribute{
				Description: "OCM API url",
				Computed:    true,
	***REMOVED***,
			"ocm_aws_account_id": schema.StringAttribute{
				Description: "OCM AWS account ID",
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
}

func (d *OCMInfoDataSource***REMOVED*** Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection***REMOVED***

	// Get the collection of cloud providers:
	d.collection = connection.AccountsMgmt(***REMOVED***.V1(***REMOVED***.CurrentAccount(***REMOVED***
	d.ocmAPI = connection.URL(***REMOVED***

}

func (d *OCMInfoDataSource***REMOVED*** Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse***REMOVED*** {
	// Get the state:
	state := &OCMInfoState{}
	diags := req.Config.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}
	// Get account information
	accountResp, err := d.collection.Get(***REMOVED***.Send(***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError("Failed to connect to OCM", "Please verify your OCM offline token"***REMOVED***
		return
	}
	obj, ok := accountResp.GetBody(***REMOVED***
	if !ok {
		resp.Diagnostics.AddError("Parsing Error", "Failed to parse OCM response"***REMOVED***
		return
	}

	state.AccountID = types.StringValue(obj.ID(***REMOVED******REMOVED***
	state.AccountName = types.StringValue(fmt.Sprintf("%s %s", obj.FirstName(***REMOVED***, obj.LastName(***REMOVED******REMOVED******REMOVED***
	state.AccountUsername = types.StringValue(obj.Username(***REMOVED******REMOVED***
	state.AccountEmail = types.StringValue(obj.Email(***REMOVED******REMOVED***
	state.OrganizationID = types.StringValue(obj.Organization(***REMOVED***.ID(***REMOVED******REMOVED***
	state.OrganizationExternalID = types.StringValue(obj.Organization(***REMOVED***.ExternalID(***REMOVED******REMOVED***
	state.OrganizationName = types.StringValue(obj.Organization(***REMOVED***.Name(***REMOVED******REMOVED***

	state.OCMAPI = types.StringValue(d.ocmAPI***REMOVED***
	state.OCMAWSAccountID = types.StringValue(extractOCMAWSAccount(d.ocmAPI***REMOVED******REMOVED***

	// Save the state:
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func extractOCMAWSAccount(url string***REMOVED*** string {
	env := ocmEnvProd
	if strings.Contains(url, "stage"***REMOVED*** {
		env = ocmEnvStage
	} else if strings.Contains(url, "integration"***REMOVED*** {
		env = ocmEnvInt
	}

	return ocmAWSAccounts[env]
}
