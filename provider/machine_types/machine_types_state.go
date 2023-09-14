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

package machine_types

type MachineTypesState struct {
	Items []*MachineTypeState `tfsdk:"items"`
}

type MachineTypeState struct {
	CloudProvider string `tfsdk:"cloud_provider"`
	ID            string `tfsdk:"id"`
	Name          string `tfsdk:"name"`
	CPU           int64  `tfsdk:"cpu"`
	RAM           int64  `tfsdk:"ram"`
}