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

import "github.com/hashicorp/terraform-plugin-framework/types"

type OCMInfoState struct {
	AccountID              types.String `tfsdk:"account_id"`
	AccountName            types.String `tfsdk:"account_name"`
	AccountUsername        types.String `tfsdk:"account_username"`
	AccountEmail           types.String `tfsdk:"account_email"`
	OrganizationID         types.String `tfsdk:"organization_id"`
	OrganizationExternalID types.String `tfsdk:"organization_external_id"`
	OrganizationName       types.String `tfsdk:"organization_name"`

	OCMAWSAccountID types.String `tfsdk:"ocm_aws_account_id"`
	OCMAPI          types.String `tfsdk:"ocm_api"`
}
