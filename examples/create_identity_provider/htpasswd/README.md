# htpasswd identity provider example

Using htpasswd authentication in OpenShift Container Platform allows you to identify users based on username/password pairs.
## Prerequisites

### An OpenShift cluster

This example assumes you have created a cluster via OCM, It can be via this terraform provider or by a different client.
We will need the cluster name and some credentials.

## Execution

### Edit the terraform.tfvars

Take your time with editing the variables file. 
There are comments that will explain what each variable is.

### Let's Go!

Simply run `terraform apply`


## OpenShift documentation

 - [htpasswd Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-htpasswd-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
