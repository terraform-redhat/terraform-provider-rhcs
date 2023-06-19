# GitHub identity provider example

Configuring [GitHub authentication](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/authorizing-oauth-apps) allows users to log in to OpenShift Container Platform with their GitHub credentials.
To prevent anyone with any GitHub user ID from logging in to your OpenShift Container Platform cluster, you can restrict access to only those in specific GitHub organizations.

## Prerequisites

### An OpenShift cluster

This example assumes you have created a cluster via OCM, It can be via this terraform provider or by a different client.
We will need the cluster name and some credentials.

### Setting up your application in GitHub

To use GitHub or GitHub Enterprise as an identity provider, you must register an application to use.

Procedure

1. Register an application on GitHub:
    
    - For GitHub, click [**Settings**](https://github.com/settings/profile) → [**Developer settings**](https://github.com/settings/apps) → [**OAuth Apps**](https://github.com/settings/developers) → [**Register a new OAuth application**](https://github.com/settings/applications/new).
        
    - For GitHub Enterprise, go to your GitHub Enterprise home page and then click **Settings → Developer settings → Register a new application**.
        
    
2. Enter an application name, for example `My OpenShift Install`.
    
3. Enter a homepage URL, such as `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>`.
    
4. Optional: Enter an application description.
    
5. Enter the authorization callback URL, where the end of the URL contains the identity provider `name`:
    
    `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>/oauth2callback/<idp-provider-name>`
    
    For example:
    `https://oauth-openshift.apps.openshift-cluster.example.com/oauth2callback/Github`
	
	> **Note**: `<idp-provider-name>` is case-sensitive. Name is defined [here](./main.tf#L37)

6. Click **Register application**. GitHub provides a client ID and a client secret. You need these values to complete the identity provider configuration.

## Execution

### Edit the terraform.tfvars

Take your time with editing the variables file. 
There are comments that will explain what each variable is.

### Let's Go!

Simply run `terraform apply`


## OpenShift documentation

 - [GitHub Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-github-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider)

