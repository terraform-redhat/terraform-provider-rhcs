// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplaneoperatorroles

import (
	"fmt"

	"github.com/openshift-online/rosa-seaplane-api/iam"
)

// desiredRole is the resolved desired state for a single operator role.
type desiredRole struct {
	Spec         iam.OperatorRoleSpec
	Name         string
	TrustPolicy  string
	PolicyARNs   []string
	IsWorkerRole bool
}

// resolveDesired renders the full desired role set for a prefix from the
// embedded definition set. Trust policies are built exclusively through the
// iam/ module builders (never hand-written JSON).
func resolveDesired(prefix, accountID, partition, issuerURL string) ([]desiredRole, error) {
	set := iam.DefaultDefinitionSet()
	roles := make([]desiredRole, 0, len(set.Roles))
	for _, spec := range set.Roles {
		name := iam.RoleName(prefix, spec.Name)
		trustPolicy := ""
		if spec.IsWorkerRole {
			trustPolicy = iam.BuildEC2TrustPolicy()
		} else {
			var err error
			trustPolicy, err = iam.BuildWebIdentityTrustPolicy(iam.TrustPolicyInput{
				Partition: partition,
				AccountID: accountID,
				IssuerURL: issuerURL,
			}, spec)
			if err != nil {
				return nil, fmt.Errorf("building trust policy for role %q: %w", name, err)
			}
		}
		roles = append(roles, desiredRole{
			Spec:         spec,
			Name:         name,
			TrustPolicy:  trustPolicy,
			PolicyARNs:   iam.ManagedPolicyARNs(spec, partition),
			IsWorkerRole: spec.IsWorkerRole,
		})
	}
	return roles, nil
}

// reconcileActions is the set of repairs the provider performs on refresh to
// converge observed AWS state with the resolved definition set.
type reconcileActions struct {
	// CreateRole is true when the role is missing in AWS.
	CreateRole bool
	// UpdateTrustPolicy is true when the trust policy drifted from the
	// expected document.
	UpdateTrustPolicy bool
	// AttachPolicies lists managed policies required by the definition set
	// that are not attached in AWS. Policies are only ever attached (never
	// detached) so roles shared with running V1-managed clusters are never
	// broken (see the design doc migration analysis).
	AttachPolicies []string
	// UpdateTags are tags to merge (e.g. the applied definition version).
	UpdateTags map[string]string
	// EnsureInstanceProfile is true when the worker role has no instance
	// profile matching the expected name.
	EnsureInstanceProfile bool
	// ForeignRole is set when the role exists in AWS but is tagged with a
	// different operator-roles prefix. Such roles are never touched.
	ForeignRole bool
}

// planReconcile diffs one observed role against its desired state and returns
// the repair actions. observed may be nil (role missing in AWS).
func planReconcile(role desiredRole, observed *RoleSummary, prefix string) reconcileActions {
	actions := reconcileActions{}
	if observed == nil {
		actions.CreateRole = true
		// Role creation does not attach policies; apply must attach the full
		// definition set afterwards.
		actions.AttachPolicies = append(actions.AttachPolicies, role.PolicyARNs...)
		return actions
	}

	// Never touch a role that belongs to a different operator-roles prefix
	// (a pre-existing role whose name collides with this prefix).
	if tagged, ok := observed.Tags[iam.TagRolePrefix]; ok && tagged != prefix {
		actions.ForeignRole = true
		return actions
	}

	if normalizePolicyDoc(observed.TrustPolicy) != normalizePolicyDoc(role.TrustPolicy) {
		actions.UpdateTrustPolicy = true
	}

	attached := make(map[string]bool, len(observed.AttachedPolicies))
	for _, policy := range observed.AttachedPolicies {
		attached[policy.ARN] = true
	}
	for _, policyARN := range role.PolicyARNs {
		if !attached[policyARN] {
			actions.AttachPolicies = append(actions.AttachPolicies, policyARN)
		}
	}

	// Refresh the discovery tags (including the applied definition version)
	// whenever they differ from the current definition set.
	expectedTags := iam.RoleTags(prefix)
	updateTags := map[string]string{}
	for key, value := range expectedTags {
		if observed.Tags[key] != value {
			updateTags[key] = value
		}
	}
	actions.UpdateTags = updateTags

	if role.IsWorkerRole {
		for _, profile := range observed.InstanceProfiles {
			if profile.Name == role.Name {
				return actions
			}
		}
		actions.EnsureInstanceProfile = true
	}
	return actions
}

// normalizePolicyDoc compares two IAM policy documents semantically by
// trimming whitespace, so AWS-normalized documents compare equal to the
// documents produced by the iam/ builders.
func normalizePolicyDoc(doc string) string {
	out := make([]byte, 0, len(doc))
	for i := 0; i < len(doc); i++ {
		c := doc[i]
		switch c {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			out = append(out, c)
		}
	}
	return string(out)
}
