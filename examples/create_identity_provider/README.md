# Identity Provider for ROSA clusters

The following pages show you how to configure various identity providers to use on your clusters. These identity providers offer ways for your users to log in to your cluster.

## Prerequisites

1. You created your [account roles using Terraform](../../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md***REMOVED***.
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md***REMOVED*** or [an unmanaged OIDC configuration](../../examples/create_rosa_cluster/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md***REMOVED***.
1. **Optional**: You have configured [your Terraform.tfvars file](../../docs/terraform-vars.md***REMOVED***.

## Types of identity providers

The following identity providers are currently supported:

1. [GitHub](../../examples/create_identity_provider/github/README.md***REMOVED***
1. [Gitlab](../../examples/create_identity_provider/gitlab/README.md***REMOVED***
1. [Google](../../examples/create_identity_provider/google/README.md***REMOVED***
1. [HTPassword](../../examples/create_identity_provider/htpasswd/README.md***REMOVED***