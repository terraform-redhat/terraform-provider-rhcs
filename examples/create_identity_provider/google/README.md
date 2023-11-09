# Google identity provider

Using Google as an identity provider allows any Google user to authenticate to your server.
You can limit authentication to members of a specific hosted domain with the `hosted_domain` configuration attribute.

## Prerequisites

1. You created your [account roles using Terraform](../../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../../examples/create_rosa_cluster/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).
1. **Optional**: You have configured your Terraform.tfvars file.

## Setting up your application in Google

You will need a client ID/secret of a [registered Google project](https://console.developers.google.com/).
The project must be configured with a redirect URI of `https://oauth-openshift.apps.<cluster-name>.<cluster-domain>/oauth2callback/<idp-provider-name>`.
For example:
`https://oauth-openshift.apps.openshift-cluster.example.com/oauth2callback/Google`.

> **Note**: `<idp-provider-name>` is case-sensitive. Name is defined [here](./main.tf#L37).

## Applying the Terraform plan

1. You need to either create a `terraform.tfvars` file in this directory or add the following items to your existing `*.tfvars` file. You may also export these variables as environmental variables with the following commands:
      1.  This value is the generated Google client secret to validate your account. It can be found in the settings of your Google account.
          ```
          export TF_VAR_google_client_secret=<google_client_secret>
          ```
      1.  This value is your Google client ID. It can be found in the settings of your Google account.   
          ```
          export TF_VAR_google_client_id=<client_id>
          ```
      1.  This value is your Google hosted domain that was generated in the previous step.  
          ```
          export TF_VAR_google_hosted_domain='["<google_hosted_domain>"]'
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
1. In your local copy of the `google` folder, run the following command:
   ````
   terraform init
   ````
   Running this command accesses all the necessary provider information to apply your Terraform plan.
1. **Optional**: Run the `plan` command to ensure that your Terraform files build correctly without errors. This is not required to apply your Terraform plans.
   ````
   terraform plan -out google.tfplan
   ````
1. Run the apply command to create your Google identity provider. 

   > **Note**: If you did not run the `plan` command, you can simply `apply` without specifying a file.

    ````
    terraform apply <"google.tfplan">
    ````
1. The Terraform applies the plan and creates your identity provider using Google. You will see a prompt to confirm you want to create these resources. Enter `yes`, then the process will complete with your resources.

## Resource clean up

After you are done with the resources you created, you should not delete them manually, but instead, use the `destroy` command. Run the following to delete all of your created resources:
  
```
terraform destroy
```

After the command is complete, your resources are deleted.

> **NOTE**: If you manually delete a resource, you create unresolvable issues within your environment.

## Additional resources

 - [Google Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-google-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider)
