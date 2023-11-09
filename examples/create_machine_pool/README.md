# Creating a machine pool for an existing Openshift cluster

This Terraform example creates a **machine pool** for an existing ROSA cluster. Creating a new machine pool allows users to add different machine types in the cluster.

## Prerequisites

* You have an AWS account.
* You have installed the latest version of the ROSA command-line interface (CLI) (`rosa`).
* You have an offline OpenShift Cluster Manager token. This token can be generated in the [Red Hat Hybrid Cloud Console](https://console.redhat.com/openshift/token).
* You have installed Terraform. See the [Terraform page](https://developer.hashicorp.com/terraform/downloads) for the latest version.
* You have already created an OpenShift Cluster Manager cluster.
## Creating machine pools

1. To run the `terraform apply` you need to set up some variables. This guide uses environmental variables. For more on Terraform variables, see [Managing Variables](https://developer.hashicorp.com/terraform/enterprise/workspaces/variables/managing-variables) in the Terraform documentation.

   > **NOTE**: If you exported these variables in your current command-line session when running the account-roles Terraform example, you do not need to export them again.

   Run the following commands to export your variables. Provide your values in lieu of the brackets. Note that any values declared in the `variables.tf` are the default values if you do not export a superseding value.

    *  OpenShift Cluster Manager offline token:

         This variable should be your full [OpenShift Cluster Manager offline token](https://console.redhat.com/openshift/token) that you generated in the prerequisites.  
        ```
        export TF_VAR_token=<ocm_offline_token> 
        ```
    *  OpenShift Cluster Manager URL:
    
       This value should always point to `https://api.openshift.com`.  
        ```
        export TF_VAR_url=<ocm_url>
        ```
    * Machine pool cluster ID:

      This ID can be found in the `rosa` CLI with the command `rosa list cluster`.
        ```
        export TF_VAR_cluster_id=<cluster_id>
        ```
    * Machine pool name:
        ```
        export TF_VAR_name=<name>
        ```
    * AWS instance type:
        ```
        export TF_VAR_machine_type=<machine_type>
        ```
    * Amount of the machine created:
        ```
        export TF_VAR_replicas=<replica>
        ```
1. In your local copy of the `create_machine_pool` folder, run the following command:
   ````
   terraform init
   ````
   Running this command accesses all the necessary provider information to apply your Terraform plan.
1. **Optional**: Run the `plan` command to ensure that your Terraform files build correctly without errors. This is not required to apply your Terraform plans.
   ````
   terraform plan -out machine-pool.tfplan
   ````
1. Run the apply command to create the machine pool. 

   > **Note**: If you did not run the `plan` command, you can simply just `apply` without specifying a file.

    ````
    terraform apply <"machine-pool.tfplan">
    ````
1. The Terraform applies the plan and creates the machine pool. You will see a prompt to confirm you want to create these resources. Enter `yes`, then the process will complete with your resources.

## Verification

1. In your command-line interface, run the following command to verify that the machine pools are created:
    ````
    rosa list machinepools -c <cluster_id>
    ````
## Resource clean up

After you are done with the resources you created, you should not delete them manually, but instead, use the `destroy` command. Run the following to delete all of your created resources:
```
terraform destroy
```
After the command is complete, your resources are deleted.

> **NOTE**: If you manually delete a resource, you create unresolvable issues within your environment.
