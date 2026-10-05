// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// ClassicCreateOnlyReplicasWarning is returned when Classic cluster replicas in
// config differ from Terraform state. The attribute is create-only: day-2
// changes update state only and do not PATCH OCM machine pools.
const ClassicCreateOnlyReplicasWarning = "replicas is only used on cluster creation and has no effect " +
	"after the cluster exists; use the rhcs_machine_pool resource to manage node replicas"

// HCPCreateOnlyReplicasWarning is returned when HCP cluster replicas in config
// differ from Terraform state. The attribute is create-only: day-2 changes
// update state only and do not PATCH OCM node pools.
const HCPCreateOnlyReplicasWarning = "replicas is only used on cluster creation and has no effect " +
	"after the cluster exists; use the rhcs_hcp_machine_pool resource to manage node replicas"

// ReconcileCreateOnlyReplicas applies create-only replicas state policy for
// Classic and HCP cluster resources. It never implies an OCM nodes PATCH.
//
// configHasReplicas must reflect Terraform config (request.Config), not plan:
// CreateOnlyOptionalInt64 plans null when omitted, but callers should still
// read Config so plan/state copies are never mistaken for an explicit setting.
//
// mismatchWarning is returned when state and config differ (for example
// ClassicCreateOnlyReplicasWarning or HCPCreateOnlyReplicasWarning).
//
// Rules:
//  1. Config omits replicas → state null, no warning
//  2. State null/unknown and config set → state = config, no warning
//  3. State equals config → keep, no warning
//  4. State differs from config → state = config + mismatchWarning
func ReconcileCreateOnlyReplicas(
	configHasReplicas bool,
	stateReplicas, configReplicas types.Int64,
	mismatchWarning string,
) (types.Int64, string) {
	if !configHasReplicas {
		return types.Int64Null(), ""
	}
	if !common.HasValue(configReplicas) {
		// Config claimed present but value unavailable; leave state unchanged.
		return stateReplicas, ""
	}
	if !common.HasValue(stateReplicas) {
		return configReplicas, ""
	}
	if stateReplicas.Equal(configReplicas) {
		return stateReplicas, ""
	}
	return configReplicas, mismatchWarning
}
