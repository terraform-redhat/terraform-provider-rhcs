// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplaneoperatorroles

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/openshift-online/rosa-seaplane-api/iam"
)

// fakeIAMClient is an in-memory IAMRoleClient for unit tests.
type fakeIAMClient struct {
	accountID string
	partition string

	roles             map[string]*RoleSummary
	instanceProfiles  map[string]bool
	deletedRoles      []string
	detached          []string
	createdRoles      []string
	updatedTrust      []string
	attached          map[string][]string
	tagged            map[string]map[string]string
	listedPrefixCalls int
}

func newFakeIAMClient() *fakeIAMClient {
	return &fakeIAMClient{
		accountID:        "123456789012",
		partition:        "aws",
		roles:            map[string]*RoleSummary{},
		instanceProfiles: map[string]bool{},
		attached:         map[string][]string{},
		tagged:           map[string]map[string]string{},
	}
}

func (f *fakeIAMClient) CallerIdentity(_ context.Context) (string, string, string, error) {
	return f.accountID, f.partition,
		fmt.Sprintf("arn:%s:iam::%s:user/test", f.partition, f.accountID), nil
}

func (f *fakeIAMClient) GetRole(_ context.Context, name string) (*RoleSummary, error) {
	role, ok := f.roles[name]
	if !ok {
		return nil, nil
	}
	copy := *role
	return &copy, nil
}

func (f *fakeIAMClient) CreateRole(_ context.Context, name, trustPolicy, permissionsBoundary string,
	tags map[string]string) (string, error) {
	if _, exists := f.roles[name]; exists {
		return "", fmt.Errorf("role %q already exists", name)
	}
	f.roles[name] = &RoleSummary{
		Name:             name,
		ARN:              fmt.Sprintf("arn:%s:iam::%s:role/%s", f.partition, f.accountID, name),
		TrustPolicy:      trustPolicy,
		Tags:             map[string]string{},
		PermissionsBound: permissionsBoundary,
	}
	for key, value := range tags {
		f.roles[name].Tags[key] = value
		f.tagged[name] = map[string]string{key: value}
	}
	f.createdRoles = append(f.createdRoles, name)
	return f.roles[name].ARN, nil
}

func (f *fakeIAMClient) UpdateTrustPolicy(_ context.Context, name, trustPolicy string) error {
	f.roles[name].TrustPolicy = trustPolicy
	f.updatedTrust = append(f.updatedTrust, name)
	return nil
}

func (f *fakeIAMClient) UpdatePermissionsBoundary(_ context.Context, name, permissionsBoundary string) error {
	f.roles[name].PermissionsBound = permissionsBoundary
	return nil
}

func (f *fakeIAMClient) TagRole(_ context.Context, name string, tags map[string]string) error {
	for key, value := range tags {
		f.roles[name].Tags[key] = value
	}
	return nil
}

func (f *fakeIAMClient) AttachPolicy(_ context.Context, roleName, policyARN string) error {
	if f.roles[roleName] != nil {
		for _, existing := range f.roles[roleName].AttachedPolicies {
			if existing.ARN == policyARN {
				return nil
			}
		}
		f.roles[roleName].AttachedPolicies = append(f.roles[roleName].AttachedPolicies,
			AttachedPolicy{ARN: policyARN})
	}
	f.attached[roleName] = append(f.attached[roleName], policyARN)
	return nil
}

func (f *fakeIAMClient) DetachPolicy(_ context.Context, roleName, policyARN string) error {
	f.detached = append(f.detached, policyARN)
	remaining := f.attached[roleName][:0]
	for _, existing := range f.attached[roleName] {
		if existing != policyARN {
			remaining = append(remaining, existing)
		}
	}
	f.attached[roleName] = remaining
	return nil
}

func (f *fakeIAMClient) EnsureInstanceProfile(_ context.Context, roleName, instanceProfileName string,
	tags map[string]string) error {
	f.instanceProfiles[instanceProfileName] = true
	if f.roles[roleName] != nil {
		f.roles[roleName].InstanceProfiles = append(f.roles[roleName].InstanceProfiles,
			InstanceProfileSummary{Name: instanceProfileName})
	}
	return nil
}

func (f *fakeIAMClient) DeleteInstanceProfile(_ context.Context, roleName, instanceProfileName string) error {
	delete(f.instanceProfiles, instanceProfileName)
	return nil
}

func (f *fakeIAMClient) DeleteRole(_ context.Context, name string) error {
	delete(f.roles, name)
	f.deletedRoles = append(f.deletedRoles, name)
	return nil
}

func (f *fakeIAMClient) ListRolesByPrefix(_ context.Context, prefix string) ([]RoleSummary, error) {
	f.listedPrefixCalls++
	var result []RoleSummary
	for name, role := range f.roles {
		if len(name) > len(prefix)+1 && name[:len(prefix)+1] == prefix+"-" {
			result = append(result, *role)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

const (
	testPrefix   = "test-prefix"
	testAccount  = "123456789012"
	testIssuer   = "https://oidc.s3.us-east-1.amazonaws.com/test-id"
	testBoundary = "arn:aws:iam::123456789012:policy/boundary"
)

func TestResolveDesired(t *testing.T) {
	desired, err := resolveDesired(testPrefix, testAccount, "aws", testIssuer)
	if err != nil {
		t.Fatal(err)
	}
	set := iam.DefaultDefinitionSet()
	if len(desired) != len(set.Roles) {
		t.Fatalf("expected %d roles, got %d", len(set.Roles), len(desired))
	}
	seenWorker := false
	for i, role := range desired {
		if role.Name != iam.RoleName(testPrefix, set.Roles[i].Name) {
			t.Errorf("role %d: expected name %q, got %q", i,
				iam.RoleName(testPrefix, set.Roles[i].Name), role.Name)
		}
		if role.TrustPolicy == "" {
			t.Errorf("role %q has no trust policy", role.Name)
		}
		if len(role.PolicyARNs) == 0 {
			t.Errorf("role %q has no managed policies", role.Name)
		}
		if set.Roles[i].IsWorkerRole {
			seenWorker = true
			if role.TrustPolicy != iam.BuildEC2TrustPolicy() {
				t.Errorf("worker role trust policy is not the EC2 trust policy")
			}
		} else if !containsSub(role.TrustPolicy, "sts:AssumeRoleWithWebIdentity") {
			t.Errorf("role %q trust policy is not a web-identity policy", role.Name)
		}
	}
	if !seenWorker {
		t.Error("definition set has no worker role")
	}
}

func containsSub(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

func TestPlanReconcile(t *testing.T) {
	desired, err := resolveDesired(testPrefix, testAccount, "aws", testIssuer)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]desiredRole{}
	for _, role := range desired {
		byName[role.Name] = role
	}
	ingress := byName[iam.RoleName(testPrefix, iam.SuffixIngress)]
	worker := byName[iam.RoleName(testPrefix, iam.SuffixWorkerRole)]

	t.Run("missing role is created", func(t *testing.T) {
		actions := planReconcile(ingress, nil, testPrefix)
		if !actions.CreateRole {
			t.Errorf("expected CreateRole, got %+v", actions)
		}
	})

	t.Run("matching role needs no update", func(t *testing.T) {
		observed := &RoleSummary{
			Name:        ingress.Name,
			TrustPolicy: ingress.TrustPolicy,
			Tags:        iam.RoleTags(testPrefix),
			AttachedPolicies: []AttachedPolicy{
				{ARN: ingress.PolicyARNs[0]},
			},
		}
		actions := planReconcile(ingress, observed, testPrefix)
		if actions.UpdateTrustPolicy || actions.CreateRole || actions.EnsureInstanceProfile ||
			len(actions.AttachPolicies) > 0 || len(actions.UpdateTags) > 0 {
			t.Errorf("expected no update, got %+v", actions)
		}
	})

	t.Run("drifted trust policy is repaired", func(t *testing.T) {
		observed := &RoleSummary{
			Name:        ingress.Name,
			TrustPolicy: "{\n  \"Version\": \"2012-10-17\"\n}",
			Tags:        iam.RoleTags(testPrefix),
			AttachedPolicies: []AttachedPolicy{
				{ARN: ingress.PolicyARNs[0]},
			},
		}
		actions := planReconcile(ingress, observed, testPrefix)
		if !actions.UpdateTrustPolicy {
			t.Errorf("expected UpdateTrustPolicy, got %+v", actions)
		}
	})

	t.Run("missing policy and stale tag are reconciled", func(t *testing.T) {
		tags := iam.RoleTags(testPrefix)
		tags[iam.TagDefinitionVersion] = "0.0.0"
		observed := &RoleSummary{
			Name:        ingress.Name,
			TrustPolicy: ingress.TrustPolicy,
			Tags:        tags,
		}
		actions := planReconcile(ingress, observed, testPrefix)
		if len(actions.AttachPolicies) != 1 || actions.AttachPolicies[0] != ingress.PolicyARNs[0] {
			t.Errorf("expected one policy to attach, got %v", actions.AttachPolicies)
		}
		if actions.UpdateTags[iam.TagDefinitionVersion] != iam.DefinitionVersion() {
			t.Errorf("expected definition-version tag update, got %v", actions.UpdateTags)
		}
	})

	t.Run("foreign role is never touched", func(t *testing.T) {
		observed := &RoleSummary{
			Name:        ingress.Name,
			TrustPolicy: "{}",
			Tags:        map[string]string{iam.TagRolePrefix: "other-prefix"},
		}
		actions := planReconcile(ingress, observed, testPrefix)
		if !actions.ForeignRole || actions.CreateRole || actions.UpdateTrustPolicy ||
			len(actions.AttachPolicies) > 0 || len(actions.UpdateTags) > 0 || actions.EnsureInstanceProfile {
			t.Errorf("expected ForeignRole with no other actions, got %+v", actions)
		}
	})

	t.Run("missing worker instance profile is ensured", func(t *testing.T) {
		observed := &RoleSummary{
			Name:        worker.Name,
			TrustPolicy: worker.TrustPolicy,
			Tags:        iam.RoleTags(testPrefix),
			AttachedPolicies: []AttachedPolicy{
				{ARN: worker.PolicyARNs[0]},
				{ARN: worker.PolicyARNs[1]},
			},
		}
		actions := planReconcile(worker, observed, testPrefix)
		if !actions.EnsureInstanceProfile {
			t.Errorf("expected EnsureInstanceProfile, got %+v", actions)
		}
		observed.InstanceProfiles = []InstanceProfileSummary{{Name: worker.Name}}
		actions = planReconcile(worker, observed, testPrefix)
		if actions.EnsureInstanceProfile {
			t.Errorf("unexpected EnsureInstanceProfile, got %+v", actions)
		}
	})
}

func TestApplyCreatesFullRoleSet(t *testing.T) {
	client := newFakeIAMClient()
	r := &SeaplaneOperatorRolesResource{}
	state := &ResourceModel{
		OperatorRolesPrefix: types.StringValue(testPrefix),
		OIDCIssuerURL:       types.StringValue(testIssuer),
		PermissionsBoundary: types.StringValue(testBoundary),
	}

	if err := r.apply(context.Background(), client, state, testAccount, "aws", false); err != nil {
		t.Fatal(err)
	}

	set := iam.DefaultDefinitionSet()
	if len(state.Roles) != len(set.Roles) {
		t.Fatalf("expected %d roles in state, got %d", len(set.Roles), len(state.Roles))
	}
	for _, role := range state.Roles {
		observed, err := client.GetRole(context.Background(), role.Name.ValueString())
		if err != nil || observed == nil {
			t.Fatalf("role %q was not created", role.Name.ValueString())
		}
		if observed.PermissionsBound != testBoundary {
			t.Errorf("role %q has wrong permissions boundary %q", role.Name, observed.PermissionsBound)
		}
		if observed.Tags[iam.TagDefinitionVersion] != iam.DefinitionVersion() {
			t.Errorf("role %q missing definition-version tag", role.Name)
		}
	}
	if state.InstanceProfileName.ValueString() != iam.RoleName(testPrefix, iam.SuffixWorkerRole) {
		t.Errorf("unexpected instance profile name %q", state.InstanceProfileName)
	}
	if !client.instanceProfiles[state.InstanceProfileName.ValueString()] {
		t.Error("worker instance profile was not created")
	}
	if state.DefinitionVersion.ValueString() != iam.DefinitionVersion() {
		t.Errorf("unexpected definition version %q", state.DefinitionVersion)
	}
}

func TestApplyReconcilesDrift(t *testing.T) {
	client := newFakeIAMClient()
	r := &SeaplaneOperatorRolesResource{}
	state := &ResourceModel{
		OperatorRolesPrefix: types.StringValue(testPrefix),
		OIDCIssuerURL:       types.StringValue(testIssuer),
	}
	ctx := context.Background()
	if err := r.apply(ctx, client, state, testAccount, "aws", false); err != nil {
		t.Fatal(err)
	}

	// Simulate drift: stale definition-version tag, a missing policy, a
	// rewritten trust policy, and a removed instance profile.
	ingressName := iam.RoleName(testPrefix, iam.SuffixIngress)
	workerName := iam.RoleName(testPrefix, iam.SuffixWorkerRole)
	client.roles[ingressName].Tags[iam.TagDefinitionVersion] = "0.0.1"
	client.roles[ingressName].AttachedPolicies = nil
	client.roles[ingressName].TrustPolicy = "{}"
	delete(client.instanceProfiles, workerName)
	client.roles[workerName].InstanceProfiles = nil

	if err := r.apply(ctx, client, state, testAccount, "aws", false); err != nil {
		t.Fatal(err)
	}

	if got := client.roles[ingressName].Tags[iam.TagDefinitionVersion]; got != iam.DefinitionVersion() {
		t.Errorf("definition-version tag not reconciled: %q", got)
	}
	if len(client.roles[ingressName].AttachedPolicies) != 1 {
		t.Errorf("policy not re-attached, got %v", client.roles[ingressName].AttachedPolicies)
	}
	if client.roles[ingressName].TrustPolicy == "{}" {
		t.Error("trust policy not repaired")
	}
	if !client.instanceProfiles[workerName] {
		t.Error("worker instance profile not re-created")
	}
}

func TestApplyRefusesForeignRole(t *testing.T) {
	client := newFakeIAMClient()
	ingressName := iam.RoleName(testPrefix, iam.SuffixIngress)
	if _, err := client.CreateRole(context.Background(), ingressName, "{}", "",
		map[string]string{iam.TagRolePrefix: "other-prefix"}); err != nil {
		t.Fatal(err)
	}
	r := &SeaplaneOperatorRolesResource{}
	state := &ResourceModel{
		OperatorRolesPrefix: types.StringValue(testPrefix),
		OIDCIssuerURL:       types.StringValue(testIssuer),
	}
	err := r.apply(context.Background(), client, state, testAccount, "aws", false)
	if err == nil {
		t.Fatal("expected an error for a foreign role")
	}
}

func TestDeleteDetachesOnlyDefinitionPolicies(t *testing.T) {
	client := newFakeIAMClient()
	r := &SeaplaneOperatorRolesResource{}
	state := &ResourceModel{
		OperatorRolesPrefix: types.StringValue(testPrefix),
		OIDCIssuerURL:       types.StringValue(testIssuer),
	}
	ctx := context.Background()
	if err := r.apply(ctx, client, state, testAccount, "aws", false); err != nil {
		t.Fatal(err)
	}

	// Attach an out-of-band policy that must never be detached, and create a
	// foreign role that must not be deleted.
	ingressName := iam.RoleName(testPrefix, iam.SuffixIngress)
	outOfBand := "arn:aws:iam::aws:policy/OutOfBand"
	if err := client.AttachPolicy(ctx, ingressName, outOfBand); err != nil {
		t.Fatal(err)
	}
	foreignName := iam.RoleName("other-prefix", iam.SuffixIngress)
	if _, err := client.CreateRole(ctx, foreignName, "{}", "",
		map[string]string{iam.TagRolePrefix: "other-prefix"}); err != nil {
		t.Fatal(err)
	}

	desired, err := resolveDesired(testPrefix, testAccount, "aws", testIssuer)
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range desired {
		observed, err := client.GetRole(ctx, role.Name)
		if err != nil || observed == nil {
			t.Fatalf("role %q not found", role.Name)
		}
		if err := r.deleteRole(ctx, client, role, observed); err != nil {
			t.Fatal(err)
		}
	}

	for _, detached := range client.detached {
		if detached == outOfBand {
			t.Error("out-of-band policy must not be detached")
		}
	}
	if len(client.detached) == 0 {
		t.Error("expected definition policies to be detached before deletion")
	}
	if client.instanceProfiles[iam.RoleName(testPrefix, iam.SuffixWorkerRole)] {
		t.Error("expected the worker instance profile to be deleted")
	}
	if observed, _ := client.GetRole(ctx, foreignName); observed == nil {
		t.Error("foreign role must not be deleted")
	}
}

func TestSuffixAfterPrefix(t *testing.T) {
	cases := []struct {
		name, prefix, input, suffix string
		ok                          bool
	}{
		{"match", "p", "p-ingress", "ingress", true},
		{"other prefix", "p", "q-ingress", "", false},
		{"exact prefix without suffix", "p", "p", "", false},
		{"worker", "p", "p-ROSA-Worker-Role", "ROSA-Worker-Role", true},
	}
	for _, c := range cases {
		suffix, ok := suffixAfterPrefix(c.input, c.prefix)
		if suffix != c.suffix || ok != c.ok {
			t.Errorf("%s: got (%q,%v), want (%q,%v)", c.name, suffix, ok, c.suffix, c.ok)
		}
	}
}
