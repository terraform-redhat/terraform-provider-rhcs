# LDAP identity provider

Configure the ldap identity provider to validate user names and passwords against an LDAPv3 server, using simple bind authentication.

## Prerequisites

1. You created your [account roles using Terraform](../../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../../examples/create_rosa_cluster/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).
1. **Optional**: You have configured your Terraform.tfvars file.

## LDAP server and configuration

For more information about LDAP server authentication and configuration, see [About LDAP authentication](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-ldap-identity-provider.html#identity-provider-overview_configuring-ldap-identity-provider).


## Applying the Terraform plan

1. You need to either create `terraform.tfvars` file in this directory or add the following items to your existing `*.tfvars` file. You may also export these variables as environmental variables with the following commands:
variable "" {
  type        = string
  description = "An RFC 2255 URL which specifies the LDAP search parameters to use."
}
      1.  This variable is a Boolean expression that allows you to decide if TLS connections to the server are allowed. The default value is **false**.   
          ```
          export TF_VAR_ldap_insecure=<true_or_false>
          ```
      1.  This variable points to an RFC 2255 URL for your LDAP search parameters. 
          ```
          export TF_VAR_ldap_url=<URL_for_parameters>
          ```
      1.  This variable should be your full [OpenShift Cluster Manager offline token](https://console.redhat.com/openshift/token) that you generated in the prerequisites.  
          ```
          export TF_VAR_token=<ocm_offline_token> 
          ```
      1.  This value should point to your OpenShift instance.  
          ```
          export TF_VAR_url=<ocm_url>
          ```
      1.  The ID of the cluster for which you are creating the identity provider. This ID can be found in the `rosa` command-line interface (CLI) with the command `rosa list cluster`. 
          ```
          export TF_VAR_cluster_id=<cluster_id>
          ```
      1.  **Optional**: This variable includes any additional trust certificate authority bundles.
          ```
          export TF_VAR_ldap_ca=<trust-certificate-authority-bundle>
          ```    
1. In your local copy of the `github` folder, run the following command:
   ````
   terraform init
   ````
   Running this command accesses all the necessary provider information to apply your Terraform plan.
1. **Optional**: Run the `plan` command to ensure that your Terraform files build correctly without errors. This is not required to apply your Terraform plan.
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

 - [LDAP Identity Provider](https://docs.openshift.com/container-platform/4.12/authentication/identity_providers/configuring-ldap-identity-provider.html)
 - [Understanding identity provider configuration](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html)
 - [Mapping Methods](https://docs.openshift.com/container-platform/4.12/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider)
