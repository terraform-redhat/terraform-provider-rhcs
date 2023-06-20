# Google identity provider example

Using Google as an identity provider allows any Google user to authenticate to your server.
You can limit authentication to members of a specific hosted domain with the `hosted_domain` configuration attribute.

This example shows how to create a Google identity provider.

## Prerequisites

### An OpenShift cluster

This example assumes you have created a cluster via OCM, It can be via this terraform provider or by a different client.
We will need the cluster name and some credentials.

### Setting up your application in Google

You will need a client ID/Secret of a [registered Google project](https://console.developers.google.com/).
The project must be configured with a redirect URI of `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>/oauth2callback/<idp-provider-name>`.
For example:
`https://oauth-openshift.apps.openshift-cluster.example.com/oauth2callback/Google`

> **Note**: `<idp-provider-name>` is case-sensitive. Name is defined [here](./main.tf#L37)

## Execution

### Edit the terraform.tfvars

Take your time with editing the variables file. 
There are comments that will explain what each variable is.

### Let's Go!

Simply run `terraform apply`


## OpenShift documentation

 - [Google Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-google-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider)
