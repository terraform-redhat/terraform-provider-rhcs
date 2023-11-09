# htpasswd identity provider

Using htpasswd authentication in OpenShift Container Platform allows you to identify users based on username and password pairs.

> **IMPORTANT**: htpasswd does not support multiple users.

## Prerequisites

1. You created your [account roles using Terraform](../../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../../examples/create_rosa_cluster/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).
1. **Optional**: You have configured your Terraform.tfvars file.

## Applying the Terraform plan

1. You need to either create a `terraform.tfvars` file in this directory or add the following items to your existing `*.tfvars` file. You may also export these variables as environmental variables with the following commands:
      1.  This value sets the user list for logging into your application.
          ```
          export TF_VAR_users=[{\"username\":\"<user-name-to-login>\",\"password\":\"<password-for-user-name>\"},...]
          ```
      1.  This variable should be your full [OpenShift Cluster Manager offline token](https://console.redhat.com/openshift/token) that you generated in the prerequisites.  
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
1. In your local copy of the `htpasswd` folder, run the following command:
   ````
   terraform init
   ````
   Running this command accesses all the necessary provider information to apply your Terraform plan.
1. **Optional**: Run the `plan` command to ensure that your Terraform files build correctly without errors. This is not required to apply your Terraform plans.
   ````
   terraform plan -out htpasswd.tfplan
   ````
1. Run the apply command to create your htpasswd identity provider. 

   > **Note**: If you did not run the `plan` command, you can simply `apply` without specifying a file.

    ````
    terraform apply <"htpasswd.tfplan">
    ````
1. The Terraform applies the plan and creates your identity provider using htpasswd. You will see a prompt to confirm you want to create these resources. Enter `yes`, then the process will complete with your resources.

## Resource clean up

After you are done with the resources you created, you should not delete them manually, but instead, use the `destroy` command. Run the following to delete all of your created resources:
  
```
terraform destroy
```

After the command is complete, your resources are deleted.

> **NOTE**: If you manually delete a resource, you create unresolvable issues within your environment.

## Additional resources

 - [htpasswd identity provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-htpasswd-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
