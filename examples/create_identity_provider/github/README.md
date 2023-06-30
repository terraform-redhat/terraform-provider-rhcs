# GitHub identity provider

Configuring [GitHub authentication](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/authorizing-oauth-apps) allows users to log in to OpenShift Container Platform with their GitHub credentials.

To prevent anyone with any GitHub user ID from logging in to your OpenShift Container Platform cluster, you can restrict access to only those in specific GitHub organizations.
## Prerequisites

1. You created your [account roles using Terraform](../../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../../examples/create_rosa_cluster/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).
1. **Optional**: You have configured your Terraform.tfvars file.

## Setting up your application in GitHub

To use GitHub or GitHub Enterprise as an identity provider, you must register an application to use.

1. Register an application on GitHub:
    - For GitHub, click [**Settings**](https://github.com/settings/profile) → [**Developer settings**](https://github.com/settings/apps) → [**OAuth Apps**](https://github.com/settings/developers) → [**Register a new OAuth application**](https://github.com/settings/applications/new).
    - For GitHub Enterprise, go to your GitHub Enterprise home page and then click **Settings → Developer settings → Register a new application**.
2. Enter an application name, for example `My OpenShift Install`.
3. Enter a homepage URL, such as `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>`.
4. **Optional**: Enter an application description.    
5. Enter the authorization callback URL, where the end of the URL contains the identity provider `name`:

    `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>/oauth2callback/<idp-provider-name>`
    
    For example:
    `https://oauth-openshift.apps.openshift-cluster.example.com/oauth2callback/Github`.
	
	> **Note**: `<idp-provider-name>` is case-sensitive. Name is defined [here](./main.tf#L37).

6. Click **Register application**. GitHub provides a client ID and a client secret. You need these values to complete the identity provider configuration.

## Applying the Terraform plan

1. You need to either create `terraform.tfvars` file in this directory or add the following items to your existing `*.tfvars` file. You may also export these variables as environmental variables with the following commands:
      1.  This value is the generated GitHub client secret to validate your account. It can be found in the settings of your GitHub account.
          ```
          export TF_VAR_github_client_secret=<github_client_secret>
          ```
      1.  This value is your GitHub client ID. It can be found in the settings of your GitHub account.   
          ```
          export TF_VAR_github_client_id=<client_id>
          ```
      1.  This value is your GitHub organization. 
          ```
          export TF_VAR_github_orgs='["<github_org>"]'
          ```
      1.  This variable is your full [OpenShift Cluster Manager offline token](https://console.redhat.com/openshift/token) that you generated in the prerequisites.  
          ```
          export TF_VAR_token=<ocm_offline_token> 
          ```
      1.  This value should always point to `https://api.openshift.com`.
          ```
          export TF_VAR_url=<ocm_url>
          ```
      1.  The ID of the cluster for which you are creating the identity provider. This ID can be found in the `rosa` command-line interface (CLI) with the command `rosa list cluster`. 
          ```
          export TF_VAR_cluster_id=<cluster_id>
          ```
1. In your local copy of the `github` folder, run the following command:
   ````
   terraform init
   ````
   Running this command accesses all the necessary provider information to apply your Terraform plan.
1. **Optional**: Run the `plan` command to ensure that your Terraform files build correctly without errors. This is not required to apply your Terraform plans.
   ````
   terraform plan -out github.tfplan
   ````
1. Run the apply command to create your GitHub identity provider. 

   > **Note**: If you did not run the `plan` command, you can simply just `apply` without specifying a file.

    ````
    terraform apply <"github.tfplan">
    ````
1. The Terraform applies the plan and creates your identity provider using GitHub. You will see a prompt to confirm you want to create these resources. Enter `yes`, then the process will complete with your resources.

## Resource clean up

After you are done with the resources you created, you should not delete them manually, but instead, use the `destroy` command. Run the following to delete all of your created resources:
  
```
terraform destroy
```

After the command is complete, your resources are deleted.

> **NOTE**: If you manually delete a resource, you create unresolvable issues within your environment.

## Additional resources

 - [GitHub Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-github-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider)

