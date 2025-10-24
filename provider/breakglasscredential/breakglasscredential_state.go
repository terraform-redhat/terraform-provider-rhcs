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

import "github.com/hashicorp/terraform-plugin-framework/types"

type BreakGlassCredential struct {
	Cluster             types.String `tfsdk:"cluster"`
	Id                  types.String `tfsdk:"id"`
	Username            types.String `tfsdk:"username"`
	ExpirationDuration  types.String `tfsdk:"expiration_duration"`
	ExpirationTimestamp types.String `tfsdk:"expiration_timestamp"`
	RevocationTimestamp types.String `tfsdk:"revocation_timestamp"`
	Status              types.String `tfsdk:"status"`
	Kubeconfig          types.String `tfsdk:"kubeconfig"`
}
