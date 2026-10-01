# PR Verification Template (`/verified`)

Copy and fill out this checklist when testing and verifying PRs for `terraform-provider-rhcs` resources.

---

### Verification Summary

- **Provider Version Tested:** `local-build` / `vX.Y.Z`
- **Cluster Type:** ROSA HCP / ROSA Classic
- **Cluster ID:** `<redacted_cluster_id>`
- **Tested By:** `@your_github_handle`

---

## Mandatory Proof Requirements

Every checklist item and its required output is mandatory. Do not leave an item unchecked or an output omitted. You may skip a test only when it cannot apply to the resource change; identify the skipped checklist item and provide a specific technical justification in **Note / Blockers**. Redact every sensitive value in posted logs, state output, and CLI output as `REDACTED`.

For every test case, you MUST include the following outputs in your execution logs:
- **Create & Update:** Must include `terraform apply`, `terraform state show`, `rosa describe`, AND a post-apply `terraform plan` showing no drift.
- **Import:** Must include `terraform import`, `terraform state show`, AND `terraform plan` showing no drift (or an expected diff for write-only/sensitive fields like secrets).
- **Destroy:** Must include `terraform destroy` AND `rosa describe` / `rosa list` confirming the resource no longer exists in OCM/cluster.
- **Validation, Re-create, and External Delete:** Must include the command output that proves the expected behavior for each test case.

---

## Verification Checklist & Test Cases

### 1. Initial Creation & Drift Check (`Create`)

- [ ] **Run `terraform plan` & `terraform apply`**
  - **Action:** Apply the initial HCL configuration.
  - **Expected Output:** Plan shows `1 to add, 0 to change, 0 to destroy`. Resource applies successfully without errors.
- [ ] **Verify State (`terraform state show`)**
  - **Action:** Inspect the state file for the created resource.
  - **Expected Output:** Specified HCL attributes match state values; computed `id` is populated.
- [ ] **Cross-check with ROSA CLI (`rosa describe / list`)**
  - **Action:** Run `rosa list <resource>` and `rosa describe <resource> --cluster <cluster_id>`.
  - **Expected Output:** CLI output matches the created resource configuration in OCM.
- [ ] **No-Op Drift Check (`terraform plan`)**
  - **Action:** Run `terraform plan` immediately after apply.
  - **Expected Output:** `No changes. Your infrastructure matches the configuration.`

---

### 2. In-Place Updates & Validation Rules (`Update`)

- [ ] **In-Place Attribute Update**
  - **Action:** Modify updatable fields in `main.tf` and run `terraform apply`.
  - **Expected Terraform Output:** Plan shows `~ update in-place` (`0 to add, 1 to change, 0 to destroy`).
- [ ] **Verify State (`terraform state show`)**
  - **Action:** Inspect updated attributes in state.
  - **Expected Output:** State values reflect the newly applied HCL configuration.
- [ ] **Cross-check with ROSA CLI (`rosa describe`)**
  - **Action:** Run `rosa describe <resource> --cluster <cluster_id>`.
  - **Expected Output:** CLI output reflects updated resource attributes in OCM.
- [ ] **No-Op Drift Check (`terraform plan`)**
  - **Action:** Run `terraform plan` immediately after update apply.
  - **Expected Output:** `No changes. Your infrastructure matches the configuration.`
- [ ] **Validation & Schema Enforcement**
  - **Action:** Pass invalid syntax or constraint-violating values in HCL.
  - **Expected Output:** Terraform fails fast at `plan` or `apply` phase with explicit validation errors.

---

### 3. Force-Replacement Triggers (`Re-create`)

- [ ] **Immutable Field Changes**
  - **Action:** Modify fields that force resource replacement (e.g., `name`, `cluster`).
  - **Expected Output:** `terraform plan` explicitly shows `-/+ destroy and then create replacement` (`1 to add, 0 to change, 1 to destroy`).

---

### 4. Resource Import (`Import`)

- [ ] **Import Pre-existing Resource**
  - **Action:** Run `terraform import <resource_type>.<name> <cluster_id>,<resource_id>`.
  - **Expected Output:** Import succeeds (`Import successful!`).
- [ ] **Post-Import State Inspection (`terraform state show`)**
  - **Action:** Inspect imported resource state.
  - **Expected Output:** All OCM attributes are correctly populated into state.
- [ ] **Post-Import Drift Check (`terraform plan`)**
  - **Action:** Run `terraform plan` right after importing.
  - **Expected Output:** `No changes` or acceptable diffs for write-only/sensitive attributes (e.g., `console_client_secret`).

---

### 5. Out-Of-Band Drift Detection (`External Delete`)

- [ ] **Delete Resource via ROSA CLI**
  - **Action:** Delete resource directly via CLI (`rosa delete <resource> --cluster <cluster_id>`).
  - **Expected Output:** Resource is removed from OCM/cluster.
- [ ] **Detect Missing Resource in Terraform**
  - **Action:** Run `terraform plan`.
  - **Expected Output:** Terraform detects resource absence during refresh and plans recreation (`+ create` / `1 to add`).

---

### 6. Resource Teardown (`Destroy`)

- [ ] **Destroy Resource via Terraform**
  - **Action:** Run `terraform destroy -auto-approve`.
  - **Expected Output:** Resource is destroyed (`1 destroyed`).
- [ ] **CLI Cleanup Verification (`rosa describe / list`)**
  - **Action:** Run `rosa describe <resource> --cluster <cluster_id>`.
  - **Expected Output:** CLI confirms resource not found (`E: <resource> '<id>' not found`).
- [ ] **Clean State Check (`terraform plan`)**
  - **Action:** Run `terraform plan`.
  - **Expected Output:** Plan shows `1 to add` (to re-create from `main.tf`) with zero orphaned state entries.

---

## Execution Logs / Proof

<details>
<summary>Click to expand execution logs (Plan / Apply / State / CLI Outputs / Destroy)</summary>

#### 1. Initial Creation
```bash
$ terraform plan
# Paste plan output here

$ terraform apply
# Paste apply output here

$ terraform plan
# Expected: No changes. Your infrastructure matches the configuration.

$ terraform state show <resource>
# Paste state show output here

$ rosa describe <resource> --cluster <cluster_id>
# Paste CLI output here
```

#### 2. In-Place Updates
```bash
$ terraform apply
# Paste update apply output here

$ terraform plan
# Expected: No changes. Your infrastructure matches the configuration.

$ terraform state show <resource>
# Paste updated state output here

$ rosa describe <resource> --cluster <cluster_id>
# Paste updated CLI output here
```

#### 2a. Validation & Schema Enforcement
```bash
$ terraform plan
# Paste validation error output here
```

#### 3. Force-Replacement Triggers
```bash
$ terraform plan
# Paste replacement plan output here
```

#### 4. Resource Import
```bash
$ terraform import <resource_type>.<name> <cluster_id>,<resource_id>
# Paste import output here

$ terraform state show <resource>
# Paste post-import state output here

$ terraform plan
# Expected: No changes (or expected sensitive/secret diffs)
```

#### 5. Out-Of-Band Drift Detection
```bash
$ rosa delete <resource> --cluster <cluster_id>
# Paste external delete output here

$ terraform plan
# Paste drift-detection plan output here
```

#### 6. Resource Destroy
```bash
$ terraform destroy
# Paste destroy output here

$ rosa describe <resource> --cluster <cluster_id>
# Expected: E: <resource> '<id>' not found
```

</details>

---

> **Note / Blockers (Optional):**
> Document any known backend constraints, pending dependency MRs, or API edge-cases here.
> For every skipped mandatory test, identify the checklist item and provide the specific technical justification for the skip.
