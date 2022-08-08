# Rosa cluster creation example

This example shows how to create a _Rosa_ cluster with IAM User. _Rosa_ stands for Red Hat Openshift Service on AWS
and is a cluster that is created in the AWS cloud infrastructure.

In order to create a _Rosa_ cluster it is necessary to have user named
`osdCcsAdmin` with the `AdministratorAccess` role. The example assumes the `osdCcsAdmin`
is already created.

To run it:

* Provide OCM Authentication Token

OCM authentication token that you can get [here](https://console.redhat.com/openshift/token).

```
export TF_VAR_token=...
```

`main.tf` file and then run the `terraform apply` command.

