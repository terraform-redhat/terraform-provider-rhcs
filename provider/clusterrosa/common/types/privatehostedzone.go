// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package types

import "github.com/hashicorp/terraform-plugin-framework/types"

type PrivateHostedZone struct {
	ID      types.String `tfsdk:"id"`
	RoleARN types.String `tfsdk:"role_arn"`
}
