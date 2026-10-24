# TF E2E Coverage Plan — Seaplane V2 HCP Flow (HostedCluster Capabilities)

**Epic:** ROSAENG-65243 (Image Registry Enable/Disable)
**Parent:** ROSA-1174 — Seaplane API passthrough features for Engineering Preview 1
**Status:** e2e validated 2026-10-17 on the ROSAENG-68171 branch (see below)

> This is the capabilities slice of the V2 HCP passthrough coverage plan,
> split out of ROSAENG-68171 so the feature can be tested, verified, and
> peer-reviewed independently. The sibling passthrough features (custom node
> pool tags ROSAENG-65173, custom network CIDRs ROSAENG-65120, GPU machine
> pools ROSAENG-65204) remain on the ROSAENG-68171 branch.

---

## Implementation status

| Epic | Provider | Unit | Subsystem | E2E | Status |
|------|----------|------|-----------|-----|--------|
| ROSAENG-65243 capabilities | ✅ `enabled/disabled_capabilities` JSON-string attributes (pathbind passthrough — the SDK stores `[]OptionalCapability`, a named type pathbind cannot convert from `[]string`, so the repo's JSON-string convention for arrays applies, same as `componentRoutes`) + whitelist/overlap validation | ✅ (via subsystem) | ✅ 4 specs (create body, replacement plan, invalid capability, enabled+disabled overlap) | ✅ `cluster_capabilities_test.go` + profile `sp-no-caps` (pinned by the spec) | **e2e validated 2026-10-17** (day1 filter, ~18m: create → Ready → TF output + Platform API assertions → clean teardown) |

---

## Feature reference

| Feature | V2 (Seaplane) API path | Mutability | rosa CLI (V2) surface |
|---------|------------------------|------------|------------------------|
| Image Registry Enable/Disable | `spec.hostedCluster.capabilities.{enabled,disabled}` — `ImageRegistry` is one of 7 valid capabilities | Create-only, immutable | `--enable-cluster-capabilities` / `--disable-cluster-capabilities` (comma-separated; validated in `cmd/create/cluster/seaplane.go` `encodeSeaplaneClusterCapabilities`) |

Related rosa CLI work: rosa CLI V2 e2e already validated —
`tests/e2e/seaplane_capabilities_test.go` (ROSAENG-68175, profile
`rosa-seaplane-disable-capabilities`): creates a cluster with
`--disable-cluster-capabilities ImageRegistry` and asserts the describe JSON
reports `capabilities.disabled: [ImageRegistry]` with no `enabled` key.

**Scope note (from JIRA comments):** the ephemeral test environment's registry
is enabled-by-default but not backed by storage (`image-registry` CO
Degraded, `StorageExists=False`). The TF test validates the *capability
setting* and cluster readiness — not registry pod health — matching the rosa
CLI test's scope.

---

## E2E spec

`tests/e2e/seaplane/cluster_capabilities_test.go`
(`[id:ROSAENG-65243]`, labels: `day1`, `seaplane-validated`, `seaplane-one`):

1. Profile `sp-no-caps` (pinned by the spec via
   `profilehandler.NewProfileHandlerByName`) with
   `disable_cluster_capabilities: ["ImageRegistry"]`.
2. Create cluster via `clusterworkflow` backend; assert the TF output echoes
   the setting.
3. Verify via Platform API: `cluster.Spec.HostedCluster.Capabilities.Disabled`
   contains `ImageRegistry` and `Enabled` is empty.
4. Verify the cluster reaches `Ready` with the capability disabled.

Negative and immutability coverage lives in the subsystem
(`subsystem/seaplane/cluster_resource_test.go`, "Cluster capabilities"
context): POST body assertion, invalid capability name rejected with the
valid-values list, enabled+disabled overlap rejected, and changing the
capabilities after create plans a replacement.

E2E requires the real V2 environment: `SEAPLANE_URL`,
`SEAPLANE_RELEASE_IMAGE`, and AWS credentials. `CLUSTER_PROFILE` is not
needed for this spec (the profile is pinned); any ambient profile only
affects the other specs in the suite.

---

## Verification

- `make generate-seaplane` after pathbind changes + `go build ./...`
- `make pre-push-checks` (fmt, lint, unit, subsystem, gitleaks,
  subsystem-registry)
- Regenerate docs (`make docs`) and review `docs/resources/cluster_seaplane.md`
- Local e2e: `SEAPLANE_URL=… SEAPLANE_RELEASE_IMAGE=… make e2e-seaplane`
  (or focus the spec with ginkgo `--label-filter` / `-focus`)
