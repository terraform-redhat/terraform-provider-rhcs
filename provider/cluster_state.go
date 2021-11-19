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

package provider

***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/types"
***REMOVED***

type ClusterState struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CloudProvider types.String `tfsdk:"cloud_provider"`
	CloudRegion   types.String `tfsdk:"cloud_region"`
	MultiAZ       types.Bool   `tfsdk:"multi_az"`
	Properties    types.Map    `tfsdk:"properties"`
	APIURL        types.String `tfsdk:"api_url"`
	ConsoleURL    types.String `tfsdk:"console_url"`
	State         types.String `tfsdk:"state"`
	Wait          types.Bool   `tfsdk:"wait"`
}
