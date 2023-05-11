# ROSA STS cluster creation example with a managed OIDC configuration ID

This example shows how to create managed OIDC config an operator IAM roles and OIDC provider before creating a cluster.

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

* Decide STS operator_role_prefix

    ```
    export TF_VAR_operator_role_prefix=...
    ```

* Decide STS account_role_prefix. if not set use the default account IAM roles
    
    ```
    export TF_VAR_account_role_prefix=...
    ```

and then run the `terraform apply` command.
