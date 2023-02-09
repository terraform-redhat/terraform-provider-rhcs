# Gitlab identity provider example

This example shows how to create a Gitlab identity provider. This can be configured to use GitLab.com or 
any other GitLab instance as an identity provider.

To run it:

* Register a new application in Gitlab

  Client ID and secret can be retrieved by registering a new Gitlab application [here](https://docs.gitlab.com/ee/integration/oauth_provider.html).

* Provide Gitlab client id by setting a value to client id   
    ```
    export TF_VAR_gitlab_client_id=...
    ```
* Provide Gitlab client secret by setting a value to client secret   
    ```
    export TF_VAR_gitlab_client_secret=...
    ```

* Provide Gitlab environment by setting a value to url. Must use an `https://` scheme, must not have query parameters and not have a fragment. 
    ```
    export TF_VAR_gitlab_url=...
    ```

and then run the `terraform apply` command.