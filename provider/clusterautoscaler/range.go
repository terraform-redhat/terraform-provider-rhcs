/*
Copyright (c***REMOVED*** 2023 Red Hat, Inc.

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

package clusterautoscaler

***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
***REMOVED***

func rangeAttribute(description string, required bool, optional bool***REMOVED*** schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Required:    required,
		Optional:    optional,
		Attributes: map[string]schema.Attribute{
			"min": schema.Int64Attribute{
				Required: true,
	***REMOVED***,
			"max": schema.Int64Attribute{
				Required: true,
	***REMOVED***,
***REMOVED***,
		Validators: []validator.Object{
			rangeValidator(description***REMOVED***,
***REMOVED***,
	}
}
