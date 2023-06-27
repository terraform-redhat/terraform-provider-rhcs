# OIDC provider creation example

This example shows how to create an OIDC config, an operator IAM roles and an OIDC provider.

In order to create unmanaged OIDC config you'll need to create those resources: 
1. OIDC config input - using the resource called `rhcs_rosa_oidc_config_input`
2. AWS resources - using the module that celled `oidc_config_input_resources` in the main.tf file
3. OIDC config = using the resource `rhcs_rosa_oidc_config`

After you created the OIDC config you can create the OIDC provider and operator roles.

To run it:

* Provide OCM Authentication Token

  OCM authentication token that you can get [here](https://console.redhat.com/openshift/token).

    ```
    export TF_VAR_token=...
    ```

* Provide OCM environment by setting a value to url

    ```
    export TF_VAR_url=...
    ```
* Indicate weather this is managed or unmanaged provider config

    ```
    export TF_VAR_managed=[true/false]
    ```
* For unmanaged provider config, provide Installer Role ARN by setting a value 

    ```
    export TF_VAR_installer_role_arn=...
    ```

* Decide STS operator_role_prefix

    ```
    export TF_VAR_operator_role_prefix=...
    ```

* Decide STS account_role_prefix. if not set use the default account IAM roles

    ```
    export TF_VAR_account_role_prefix=...
    ```
* You can decide which cloud region to use, this is optional, default us-east-2
    ```
    export TF_VAR_cloud_region=...
    ```

* Provide List of AWS resource tags to apply (optional):
    ```
    export TF_VAR_tags=...
    ```

and then run the `terraform apply` command.
