// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplaneoperatorroles

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/openshift-online/rosa-seaplane-api/iam"
)

// rolePrefixTag is the tag key recording the operator-roles prefix a role
// belongs to (used to refuse managing foreign roles).
const rolePrefixTag = iam.TagRolePrefix

// roleDefinition aliases the iam module's OperatorRoleSpec for readability in
// discovery code.
type roleDefinition = iam.OperatorRoleSpec

// defaultDefinitionSet returns a copy of the embedded iam definition set.
func defaultDefinitionSet() iam.DefinitionSet {
	return iam.DefaultDefinitionSet()
}

// iamRoleTags returns the discovery tags applied to every role created from
// the definition set, including the applied definition version.
func iamRoleTags(prefix string) map[string]string {
	return iam.RoleTags(prefix)
}

// iamDefinitionVersion returns the version of the embedded definition set.
func iamDefinitionVersion() string {
	return iam.DefinitionVersion()
}

// stringList converts a Go string slice to a Terraform string-element slice.
func stringList(values []string) []types.String {
	if len(values) == 0 {
		return nil
	}
	result := make([]types.String, 0, len(values))
	for _, value := range values {
		result = append(result, types.StringValue(value))
	}
	return result
}
