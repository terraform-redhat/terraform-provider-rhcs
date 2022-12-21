# ROSA STS cluster creation example

This example shows how to create an operator IAM roles and oidc provider for an existing ROSA STS cluster.
The variables are depended on the output of ROSA STS cluster creation.

To run it:

* Provide OCM Authentication Token

    OCM authentication token that you can get [here](https://console.redhat.com/openshift/token***REMOVED***.

    ```
    export TF_VAR_token=...
    ```

* Provide OCM environment by setting a value to url

    ```
    export TF_VAR_url=...
    ```

* Decide STS operator_role_prefix

    ```
    export TF_VAR_operator_role_prefix=...
    ```

* Provide value to cluster_id
    
    ```
    export TF_VAR_cluster_id=...
    ```

* Provide oidc_endpoint_url

    ```
    export TF_VAR_oidc_endpoint_url=...
    ```

* Provide value to oidc_thumbprint
    
    ```
    export TF_VAR_oidc_thumbprint=...
    ```

* Decide STS account_role_prefix. if not set use the default account IAM roles
    
    ```
    export TF_VAR_account_role_prefix=...
    ```

and then run the `terraform apply` command.
