# Creating a ROSA HCP cluster with a managed OIDC configuration ID

This Terraform example creates a ROSA HCP cluster that uses a managed OIDC configuration. For more information on OIDC configurations, see the [OpenID Connect Overview](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/introduction_to_rosa/rosa-oidc-overview) in the Red Hat Customer Portal. It also creates cluster admin user for the HCP cluster. And, AWS account and operator roles, and the policies for the roles.

## Prerequisites

* You have an AWS account. Your locally configured AWS credentials are used to access to AWS API for validation purposes.
* You have installed the latest version of the ROSA CLI (`rosa`).
* You have installed the latest version of the Openshift CLI (`oc`).
* You have an offline OpenShift Cluster Manager token. This token can be generated in the [Red Hat Hybrid Cloud Console](https://console.redhat.com/openshift/token).
* You have installed Terraform. See the [Terraform page](https://developer.hashicorp.com/terraform/downloads) for the latest version.
* You have created an AWS VPC with public and private subnets.

## ROSA HCP cluster creation

1. To run the `terraform apply` you need to set up some variables. This guide uses environmental variables. For more on Terraform variables, see [Managing Variables](https://developer.hashicorp.com/terraform/enterprise/workspaces/variables/managing-variables) in the Terraform documentation.

   > **NOTE**: If you exported these variables in your current command-line session when running the account-roles Terraform example, you do not need to export them again.

   Run the following commands to export your variables. Provide your values in lieu of the brackets. Note that any values declared in the `variables.tf` are the default values if you do not export a superseding value.
        
    1. Your account role prefix prepends to all of your created roles. This value should match the value you set when creating your account roles. This value cannot end with a hyphen (-).
        ```
        export TF_VAR_account_role_prefix=<account_role_prefix>
        ```
    1.  Your Operator-role prefix prepends all of your Operator roles and policies. You should use the same value as your account-role prefix for consistency, though it is not required.
        ```
        export TF_VAR_operator_role_prefix=<operator_role_prefix>
        ```
    1. The availability zone should be a zone within your AWS region, which you chose with the `TF_VAR_cloud_region` variable. This value must be enclosed by double quotes, a bracket pair, and single quotes.
        ```    
        export TF_VAR_availability_zones='["<aws_az_region>"]' 
        ```
    1. The aws subnet ids should be AWS subnets within your AWS VPC, which you chose with the `TF_VAR_aws_subnet_ids` variable. This value must be enclosed by double quotes, a bracket pair, and single quotes.
        ```
        export TF_VAR_aws_subnet_ids='["<aws_public_subnetid>","<aws_private_subnetid>"]'
    1. This value sets the AWS region where your cluster launches.
        ```
        export TF_VAR_cloud_region=<aws_region_name>
        ```
    1.  This value is your cluster name. The cluster name cannot exceed 15 characters or the `terraform apply` fails.  
        ```
        export TF_VAR_cluster_name=<cluster_name>
        ```
    1.  This variable should be your full [OpenShift Cluster Manager offline token](https://console.redhat.com/openshift/token) that you generated in the prerequisites.  
        ```
        export TF_VAR_token=<ocm_offline_token>
        ```
    1.  This value should always point to `https://api.openshift.com`.  
        ```
        export TF_VAR_url=<ocm_url>
        ```
    1.  **Optional**: You can set the desired cluster admin username with this variable. The default is cluster-admin.
        ```    
        export TF_VAR_username=<cluster_admin_username>
        ```
    1.  **Optional**: You can set the desired OpenShift version with this variable. The default is 4.13.8.
        ```
        export TF_VAR_openshift_version=<choose_openshift_version>
    1.  **Optional**: If you want to set any specific AWS tags for your cluster, you can use this variable to declare those tags.   
         ```
         export TF_VAR_tags=<aws_resource_tags>
         ```
   ````
   terraform init
   ````
   Running this command accesses all the necessary provider information to apply your Terraform plan.
1. **Optional**: Run the `plan` command to ensure that your Terraform files build correctly without errors. This is not required to apply your Terraform plans.
   ````
   terraform plan -out rosa-hcp-cluster.tfplan
   ````
1. Run the apply command to create your ROSA HCP cluster. 

   > **Note**: If you did not run the `plan` command, you can simply just `apply` without specifying a file.

    ````
    terraform apply <"rosa-hcp-cluster.tfplan">
    ````
1. The Terraform applies the plan and creates your HCP cluster. You will see a prompt to confirm you want to create these resources. Enter `yes`, then the process will complete with your resources.

## Verification

1. In the `rosa` command-line interface (CLI), run the following command to verify that the cluster was created:
    ````
    rosa list cluster
    ````
1. You see your cluster when the command finishes. 
    ````
    ID                                NAME            STATE  TOPOLOGY
    24cg9rtij6pshqaqrgephn2rrfte1pd6  user-rosa-tf-2  ready  Hosted CP

1. In the `rosa` CLI, run the following command to verify that the account roles are created:
    ````
    rosa list account-roles
    ````
1. You see your roles when the command finishes.
    ````
    I: Fetching account roles
    ROLE NAME                           ROLE TYPE      ROLE ARN                                                    OPENSHIFT VERSION  AWS Managed
    ManagedOpenShift-ControlPlane-Role  Control plane  arn:aws:iam::XXXXX:role/ManagedOpenShift-ControlPlane-Role  4.13.8             No
    ManagedOpenShift-Installer-Role     Installer      arn:aws:iam::XXXXX:role/ManagedOpenShift-Installer-Role     4.13.8             No
    ManagedOpenShift-Support-Role       Support        arn:aws:iam::XXXXX:role/ManagedOpenShift-Support-Role       4.13.8             No
    ManagedOpenShift-Worker-Role        Worker         arn:aws:iam::XXXXX:role/ManagedOpenShift-Worker-Role        4.13.8             No

1. In the `oc` command-line interface (CLI), run the following command to verify that you can login to the cluster:
    ````
    oc login <openshift_api_url> --username cluster-admin --password <get_random_password_from_terraform.tfstate>
    ````
1. You see your login session when the command finishes.
    ````
    Login successful.

    You have access to 75 projects, the list has been suppressed. You can list all projects with 'oc projects'

    Using project "default".

## Resource clean up

After you are done with the resources you created, you should not delete them manually, but instead, use the `destroy` command. Run the following to delete all of your created resources:
  
```
terraform destroy
```

After the command is complete, your resources are deleted.

> **NOTE**: If you manually delete a resource, you create unresolvable issues within your environment.
