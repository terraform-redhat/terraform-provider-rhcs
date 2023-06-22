# LDAP identity provider example

Configure the ldap identity provider to validate user names and passwords against an LDAPv3 server, using simple bind authentication.

## Prerequisites

### An OpenShift cluster

This example assumes you have created a cluster via OCM, It can be via this terraform provider or by a different client.
We will need the cluster name and some credentials.

### LDAP server and configured

Please read the [official docs](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-ldap-identity-provider.html#identity-provider-overview_configuring-ldap-identity-provider).


## Execution

### Setting variables
Take a look at `variables.tf` to see what variables are needed.

You have 2 options to pass variables.
 1. Create a `terraform.tfvars` file
 2. Set environment variables with a format of `TF_VAR_<variable_name>` for example: `export TF_VAR_token="..."`

### Let's Go!

Simply run `terraform apply`


## OpenShift documentation

 - [LDAP Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-ldap-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider)

