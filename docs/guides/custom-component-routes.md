---
page_title: "Custom Component Routes for ROSA HCP Clusters"
subcategory: ""
description: |-
  Instructions on how to configure custom hostnames and TLS certificates for Console and Downloads routes on ROSA HCP clusters.
---

# Configuring Custom Component Routes

Custom component routes enable you to configure custom hostnames and TLS certificates for the OpenShift Console and Downloads routes on ROSA HCP clusters. This is useful for organizations with strict enterprise domain requirements, compliance policies, or those migrating from ROSA Classic clusters that already use custom component routes.

## Overview

Component routes allow you to override the default hostnames for cluster management interfaces:

- **Console**: The OpenShift web console (default: `console-openshift-console.apps.<cluster-domain>`)
- **Downloads**: The CLI and tools download page (default: `downloads-openshift-console.apps.<cluster-domain>`)

**Note**: customized OAuth component routes are not supported

## Prerequisites

Before configuring custom component routes, ensure you have:

1. A ROSA HCP cluster created via Terraform using the `rhcs_cluster_rosa_hcp` resource. This is a day-2 task only.
2. TLS certificates and keys for your custom hostnames
3. TLS secrets created in the `openshift-config` namespace on your hosted cluster
4. Minimum ROSA CLI version: 1.2.65
5. Minimum RHCS provider version: 1.7.8
6. DNS records pointing your custom hostnames to the cluster's ingress load balancer

### Creating TLS Secrets

Before configuring component routes in Terraform, create the TLS secrets on your cluster:

```bash
# Create secret for console route
oc create secret tls console-tls-secret \
  --cert=/path/to/console.crt \
  --key=/path/to/console.key \
  -n openshift-config

# Create secret for downloads route
oc create secret tls downloads-tls-secret \
  --cert=/path/to/downloads.crt \
  --key=/path/to/downloads.key \
  -n openshift-config
```

## Configuration Examples

### Setting Component Routes for Console and Downloads

Configure both console and downloads routes with custom hostnames:

```terraform
resource "rhcs_hcp_default_ingress" "default_ingress" {
  cluster          = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  listening_method = "external"

  component_routes = {
    console = {
      hostname       = "console.company.com"
      tls_secret_ref = "console-tls-secret"
    }
    downloads = {
      hostname       = "downloads.company.com"
      tls_secret_ref = "downloads-tls-secret"
    }
  }
}
```

### Setting Only Console Route

Configure just the console route, leaving downloads at the default:

```terraform
resource "rhcs_hcp_default_ingress" "default_ingress" {
  cluster          = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  listening_method = "external"

  component_routes = {
    console = {
      hostname       = "console.company.com"
      tls_secret_ref = "console-tls-secret"
    }
  }
}
```

### Clearing Component Routes

To remove custom component routes and revert to defaults, remove the `component_routes` block entirely:

```terraform
resource "rhcs_hcp_default_ingress" "default_ingress" {
  cluster          = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  listening_method = "external"
  # component_routes block removed - routes will revert to defaults
}
```

Run `terraform apply` to apply the change.

### Clearing Individual Routes

To clear one route while keeping another, simply remove that route from the map:

```terraform
resource "rhcs_hcp_default_ingress" "default_ingress" {
  cluster          = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  listening_method = "external"

  component_routes = {
    # Only console route specified - downloads will revert to default
    console = {
      hostname       = "console.company.com"
      tls_secret_ref = "console-tls-secret"
    }
  }
}
```

## Verifying the Configuration

After applying your Terraform configuration, verify that the component routes are configured correctly:

### Check Cluster Ingress Configuration

```bash
oc get ingresses.config.openshift.io cluster -o jsonpath='{.spec.componentRoutes}' | jq
```

Expected output:

```json
[
  {
    "hostname": "console.company.com",
    "name": "console",
    "namespace": "openshift-console",
    "servingCertKeyPairSecret": {
      "name": "console-tls-secret"
    }
  },
  {
    "hostname": "downloads.company.com",
    "name": "downloads",
    "namespace": "openshift-console",
    "servingCertKeyPairSecret": {
      "name": "downloads-tls-secret"
    }
  }
]
```

### Check Custom Routes

Verify that the custom routes are created:

```bash
oc get routes -n openshift-console
```

You should see routes like `console-custom` and `downloads-custom` with your custom hostnames.

### Verify Redirection

The original routes will redirect (HTTP 301) to the custom hostnames:

```bash
curl -I https://console-openshift-console.apps.<cluster-domain>
# Should return 301 redirect to https://console.company.com
```

### Test Access

Access the console via your custom hostname:

```bash
curl -k https://console.company.com
# Should return the OpenShift console page
```

## Configuration Reference

### component_routes

- **Type**: Map of objects (optional)
- **Valid keys**: `console`, `downloads`
- **Description**: Component route configuration for console and downloads routes

Each component route object supports:

- `hostname` (Required) - Custom hostname for the route (e.g., `console.company.com`)
- `tls_secret_ref` (Required) - Name of the TLS secret in the `openshift-config` namespace (e.g., `console-tls-secret`)

**Requirements**:

- Both `hostname` and `tls_secret_ref` must be provided together or both must be empty
- The referenced TLS secret must exist in the cluster (`openshift-config` namespace) before applying the configuration

## Additional Notes

- **OAuth not supported**: Unlike ROSA Classic clusters, OAuth component routes cannot be configured on HCP clusters because the OAuth server runs on the management cluster
- **DNS configuration**: Ensure your custom hostnames resolve to the cluster's ingress load balancer before applying the configuration
- **Certificate validation**: The TLS certificates must be valid for the custom hostnames and trusted by clients accessing the console

## Using the ROSA CLI

Component routes can also be configured via the ROSA CLI:

### Flag Mode

```bash
rosa edit ingress -c <cluster-name> \
  --component-routes 'console: hostname=console.company.com;tlsSecretRef=console-tls-secret, downloads: hostname=downloads.company.com;tlsSecretRef=downloads-tls-secret'
```

### Interactive Mode

```bash
rosa edit ingress -c <cluster-name> -i
```

The interactive mode will prompt for console and downloads routes (OAuth is not prompted for HCP clusters).

## Troubleshooting

### Validation Errors

If you receive a 400 error from the backend:

```
Can't update 'console' component route hostname without also supplying a new TLS secret reference
```

This means you provided only `hostname` or only `tls_secret_ref`. Both fields must be provided together.

### Secret Not Found

If the cluster cannot find the TLS secret:

1. Verify the secret exists: `oc get secret -n openshift-config <secret-name>`
2. Verify the secret is of type `kubernetes.io/tls`
3. Ensure the secret contains both `tls.crt` and `tls.key` data

## Related Documentation

- [rhcs_hcp_default_ingress Resource](https://registry.terraform.io/providers/terraform-redhat/rhcs/latest/docs/resources/hcp_default_ingress)
- [rhcs_cluster_rosa_hcp Resource](https://registry.terraform.io/providers/terraform-redhat/rhcs/latest/docs/resources/cluster_rosa_hcp)
- [ROSA Documentation](https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/)
