---
page_title: "Use tfvars file"
subcategory: ""
description: |-
  Instructions on how to add your variables using tfvars file.
---

# .tfvars file

Terraform allows you to define variable files called `*.tfvars` to create a reusable file for all the variables for a project. The following is an example that covers all of the required variables to run a majority of the Terraform examples in this repository. 

If a particular module does not use a variable, you will see the following error. This error can be ignored.

```
│ Warning: Value for undeclared variable
│ 
│ The root module does not declare a variable named "operator_role_prefix" but a value was found in file
│ "terraform.tfvars". If you meant to use this value, add a "variable" block to the configuration.
│ 
│ To silence these warnings, use TF_VAR_... environment variables to provide certain "global" settings to all configurations in your organization. To
│ reduce the verbosity of these warnings, use the -compact-warnings option.
╵

```

You may still define environmental variables. The exported values will override any of the variables in both the `variables.tf` file and the `terraform.tfvars` file.

### Example Terraform .tfvars

The following example should serve as a basis for your own `*.tfvars` files. You can create multiple versions of this file, and then, apply and destroy using this file with the `-var-file=` flag. 

This example only includes the variables needed for creating your account-wide roles and creating cluster with a managed OIDC configuration. You can also add the needed variables for creating your identity provider and machine pools to this file.

> **NOTE**: The `token` value in this example requires you to generate an offline OpenShift Cluster Manager token. You can do that in the [Red Hat Hybrid Cloud Console](https://console.redhat.com/openshift/token).

```
account_role_prefix = "<user-prefix>"
availability_zones   = ["<az-within-cloud-region>"]
cloud_region = "<aws-cloud-region>"
cluster_name = "<name-of-cluster>"
operator_role_prefix = "<user-prefix>
token = "<ocm-offline-token>"
url = "<url-of-environment"
```

### Additional resources

* See the Terraform documentation for more information on [Terraform variables](https://developer.hashicorp.com/terraform/language/values/variables).
