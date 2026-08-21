# OSD WIF GCP Module

Creates the GCP-side resources for Workload Identity Federation (WIF) used by OpenShift Dedicated (OSD) clusters.
Mirrors what `ocm gcp create wif-config --mode auto` provisions: workload identity pool, OIDC provider, service accounts, custom roles, and IAM bindings.

## Prerequisites

- An `rhcs_wif_config` resource must be created first so OCM returns the GCP blueprint.
- The blueprint (pool, service_accounts, support) is exposed as computed attributes on the WIF config.

## Two-phase apply

The module's `for_each` (service accounts, custom roles, support bindings) depends on OCM's blueprint.
Terraform requires `for_each` keys to be known at plan time, but the blueprint is only available after `rhcs_wif_config` is created in OCM.

**Phase 1:** Create the WIF config so OCM returns the blueprint:

```bash
terraform apply -target=rhcs_wif_config.wif
```

**Phase 2:** With the blueprint in state, run a full apply:

```bash
terraform apply
```

From the provider repo root, use the Makefile target:

```bash
make example.cluster
```

## Resources

- **Workload Identity Pool** – IAM workload identity pool
- **OIDC Provider** – Identity provider with issuer, JWKS, and audiences from OCM
- **Service Accounts** – One per entry in the OCM blueprint
- **Custom IAM Roles** – For non-predefined roles in the blueprint
- **Project IAM Bindings** – Roles bound to service accounts at project level
- **SA-level IAM Bindings** – `workloadIdentityUser` (WIF), `serviceAccountTokenCreator` (impersonation), resource bindings
- **Support Access** – Red Hat support group roles (when configured)

## Usage

```hcl
resource "rhcs_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
  gcp = {
    project_id     = var.gcp_project_id
    project_number = tostring(data.google_project.project.number)
    role_prefix    = var.role_prefix
  }
}

module "wif_gcp" {
  source = "./modules/wif-gcp"

  project_id               = var.gcp_project_id
  display_name             = rhcs_wif_config.wif.display_name
  pool_id                  = rhcs_wif_config.wif.gcp.workload_identity_pool.pool_id
  identity_provider        = {
    identity_provider_id = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.identity_provider_id
    issuer_url           = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.issuer_url
    jwks                 = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.jwks
    allowed_audiences    = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.allowed_audiences
  }
  service_accounts         = rhcs_wif_config.wif.gcp.service_accounts
  support                  = rhcs_wif_config.wif.gcp.support
  impersonator_email       = rhcs_wif_config.wif.gcp.impersonator_email
  federated_project_id     = rhcs_wif_config.wif.gcp.federated_project_id
  federated_project_number = tostring(coalesce(rhcs_wif_config.wif.gcp.federated_project_number, data.google_project.project.number))
}

resource "rhcs_cluster_osd_gcp" "cluster" {
  depends_on     = [module.wif_gcp]
  name           = var.cluster_name
  wif_config_id  = rhcs_wif_config.wif.id
  # ... other cluster attributes
}
```

## Variables

| Name | Description |
|------|-------------|
| `project_id` | GCP project ID |
| `display_name` | WIF config display name |
| `pool_id` | Workload identity pool ID from OCM |
| `identity_provider` | OIDC provider config (issuer_url, jwks, allowed_audiences) |
| `service_accounts` | Service accounts blueprint from OCM |
| `support` | Support access config (optional) |
| `impersonator_email` | Impersonator SA email |
| `federated_project_id` | Project for the pool (if different) |
| `federated_project_number` | Project number for WIF principals |

## Outputs

| Name | Description |
|------|-------------|
| `workload_identity_pool_name` | Full resource name of the pool |
| `workload_identity_pool_id` | Pool ID |
| `workload_identity_pool_provider_name` | Full resource name of the OIDC provider |
| `service_account_emails` | Map of SA ID to email |
