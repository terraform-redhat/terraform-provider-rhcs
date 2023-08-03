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

package machinepool

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func MachinePoolFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster": {
			Description: "Identifier of the cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description:      "Name of the machine pool.Must consist of lower-case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character.",
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: NameValidators,
		},
		"machine_type": {
			Description: "Identifier of the machine type used by the nodes, " +
				"for example `r5.xlarge`. Use the `rhcs_machine_types` data " +
				"source to find the possible values.",
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"replicas": {
			Description:   "The number of machines of the pool",
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"autoscaling_enabled", "min_replicas", "max_replicas"},
		},
		"use_spot_instances": {
			Description: "Use Spot Instances.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"max_spot_price": {
			Description: "Max Spot price.",
			Type:        schema.TypeFloat,
			Optional:    true,
			ForceNew:    true,
		},
		"autoscaling_enabled": {
			Description:   "Enables autoscaling. This variable requires you to set a maximum and minimum replicas range using the `max_replicas` and `min_replicas` variables.",
			Type:          schema.TypeBool,
			Optional:      true,
			ConflictsWith: []string{"replicas"},
			RequiredWith:  []string{"max_replicas", "max_replicas"},
		},
		"min_replicas": {
			Description:   "The minimum number of replicas for autoscaling.",
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"replicas"},
			RequiredWith:  []string{"autoscaling_enabled", "max_replicas"},
		},
		"max_replicas": {
			Description:   "The maximum number of replicas for autoscaling functionality.",
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"replicas"},
			RequiredWith:  []string{"min_replicas", "autoscaling_enabled"},
		},
		"taints": {
			Description: "Taint for machine pool. Format should be a comma-separated " +
				"list of 'key=value:ScheduleType'. This list will overwrite any modifications " +
				"made to node taints on an ongoing basis.\n",
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: TaintsSchema(),
			},
			Optional: true,
		},
		"labels": {
			Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
				" This list will overwrite any modifications made to node labels on an ongoing basis.",
			Type:     schema.TypeMap,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
		"multi_availability_zone": {
			Description: "Create a multi-AZ machine pool for a multi-AZ cluster (default true)",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"availability_zone": {
			Description:   "Select availability zone to create a single AZ machine pool for a multi-AZ cluster",
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"subnet_id"},
		},
		"subnet_id": {
			Description:   "Select subnet to create a single AZ machine pool for BYOVPC cluster",
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"availability_zone"},
		},
	}
}

type MachinePoolState struct {
	// Required
	Cluster     string `tfsdk:"cluster"`
	MachineType string `tfsdk:"machine_type"`
	Name        string `tfsdk:"name"`

	//Optional
	Replicas              *int              `tfsdk:"replicas"`
	UseSpotInstances      *bool             `tfsdk:"use_spot_instances"`
	MaxSpotPrice          *float64          `tfsdk:"max_spot_price"`
	AutoScalingEnabled    *bool             `tfsdk:"autoscaling_enabled"`
	MinReplicas           *int              `tfsdk:"min_replicas"`
	MaxReplicas           *int              `tfsdk:"max_replicas"`
	Taints                []Taint           `tfsdk:"taints"`
	Labels                map[string]string `tfsdk:"labels"`
	MultiAvailabilityZone *bool             `tfsdk:"multi_availability_zone"`
	AvailabilityZone      *string           `tfsdk:"availability_zone"`
	SubnetID              *string           `tfsdk:"subnet_id"`

	// Computed
	ID string `tfsdk:"id"`
}

func NameValidators(i interface{}, path cty.Path) diag.Diagnostics {
	machinepoolName, ok := i.(string)
	if !ok {
		return diag.Errorf("Invalid type for attribute `name`. Expected type `string`")
	}

	if !machinepoolNameRE.MatchString(machinepoolName) {
		return diag.Errorf(
			fmt.Sprintf("Can't create machine pool: Can't create machine pool with name '%s'. Expected a valid value for 'name' matching %s",
				machinepoolName, machinepoolNameRE,
			),
		)
	}

	return nil
}
