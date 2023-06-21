# LDAP identity provider example

Configure the ldap identity provider to validate user names and passwords against an LDAPv3 server, using simple bind authentication.

## Prerequisites

### An OpenShift cluster

This example assumes you have created a cluster via OCM, It can be via this terraform provider or by a different client.
We will need the cluster name and some credentials.

### LDAP server and configured

Please read the [official docs](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-ldap-identity-provider.html#identity-provider-overview_configuring-ldap-identity-provider***REMOVED***.


## Execution

### Edit the terraform.tfvars

Take your time with editing the variables file. 
There are comments that will explain what each variable is.

### Let's Go!

Simply run `terraform apply`


## OpenShift documentation

 - [LDAP Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-ldap-identity-provider.html***REMOVED***
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html***REMOVED***
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider***REMOVED***

