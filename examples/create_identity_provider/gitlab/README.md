# GitLab identity provider

Configuring GitLab authentication allows users to log in to OpenShift Container Platform with their GitLab credentials.

If you use GitLab version 7.7.0 to 11.0, you can connect using the [OAuth integration](http://doc.gitlab.com/ce/integration/oauth_provider.html). If you use GitLab version 11.1 or later, you can use [OpenID Connect](https://docs.gitlab.com/ce/integration/openid_connect_provider.html) (OIDC) to connect instead of OAuth.

## Prerequisites

1. You created your [account roles using Terraform](../../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../../examples/create_rosa_cluster/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).
1. **Optional**: You have configured your Terraform.tfvars file.

## Setting up your application in GitLab

You will need a client ID/secret of a [registered GitLab OAuth application](https://docs.gitlab.com/ce/api/oauth2.html). 
The application must be configured with a callback URL of `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>/oauth2callback/<idp-provider-name>`
For example:
`https://oauth-openshift.apps.openshift-cluster.example.com/oauth2callback/Gitlab`.

> **Note**: `<idp-provider-name>` is case-sensitive. Name is defined [here](./main.tf#L37).

## Applying the Terraform plan

1. You need to either create a `terraform.tfvars` file in this directory or add the following items to your existing `*.tfvars` file. You may also export these variables as environmental variables with the following commands:
      1.  This value is the generated GitLab client secret to validate your account. It can be found in the settings of your GitLab account.
          ```
          export TF_VAR_gitlab_client_secret=<gitlab_client_secret>
          ```
      1.  This value is your GitLab client ID. It can be found in the settings of your GitLab account.   
          ```
          export TF_VAR_gitlab_client_id=<client_id>
          ```
      1.  This value is your GitLab URL that was generated in the previous step.  
          ```
          export TF_VAR_gitlab_url='["<gitlab_url>"]'
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
1. In your local copy of the `gitlab` folder, run the following command:
   ````
   terraform init
   ````
   Running this command accesses all the necessary provider information to apply your Terraform plan.
1. **Optional**: Run the `plan` command to ensure that your Terraform files build correctly without errors. This is not required to apply your Terraform plans.
   ````
   terraform plan -out gitlab.tfplan
   ````
1. Run the apply command to create your GitLab identity provider. 

   > **Note**: If you did not run the `plan` command, you can simply just `apply` without specifying a file.

    ````
    terraform apply <"gitlab.tfplan">
    ````
1. The Terraform applies the plan and creates your identity provider using GitLab. You will see a prompt to confirm you want to create these resources. Enter `yes`, then the process will complete with your resources.

## Resource clean up

After you are done with the resources you created, you should not delete them manually, but instead, use the `destroy` command. Run the following to delete all of your created resources:
  
```
terraform destroy
```

After the command is complete, your resources are deleted.

> **NOTE**: If you manually delete a resource, you create unresolvable issues within your environment.

## Additional resources

 - [GitLab Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-gitlab-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider)
