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

package dnsdomain

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type DNSDomainState struct {
	// computed
	ID    string `tfsdk:"id"`
	DnsID string `tfsdk:"dns_id"`
}

func DNSDomainFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"dns_id": {
			Description: "Unique identifier of the DNS Domain",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
