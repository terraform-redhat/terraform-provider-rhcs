# GitLab identity provider example

Configuring GitLab authentication allows users to log in to OpenShift Container Platform with their GitLab credentials.

If you use GitLab version 7.7.0 to 11.0, you connect using the [OAuth integration](http://doc.gitlab.com/ce/integration/oauth_provider.html***REMOVED***. If you use GitLab version 11.1 or later, you can use [OpenID Connect](https://docs.gitlab.com/ce/integration/openid_connect_provider.html***REMOVED*** (OIDC***REMOVED*** to connect instead of OAuth.
This example shows how to create a Google identity provider.

## Prerequisites

### An OpenShift cluster

This example assumes you have created a cluster via OCM, It can be via this terraform provider or by a different client.
We will need the cluster name and some credentials.

### Setting up your application in GitLab

You will need a client ID/Secret of a [registered GitLab OAuth application](https://docs.gitlab.com/ce/api/oauth2.html***REMOVED***. 
The application must be configured with a callback URL of `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>/oauth2callback/<idp-provider-name>`
For example:
`https://oauth-openshift.apps.openshift-cluster.example.com/oauth2callback/Gitlab`

> **Note**: `<idp-provider-name>` is case-sensitive. Name is defined [here](./main.tf#L37***REMOVED***

## Execution

### Edit the terraform.tfvars

Take your time with editing the variables file. 
There are comments that will explain what each variable is.

### Let's Go!

Simply run `terraform apply`


## OpenShift documentation

 - [GitLab Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-gitlab-identity-provider.html***REMOVED***
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html***REMOVED***
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider***REMOVED***

