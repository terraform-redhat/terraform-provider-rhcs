# ROSA cluster creation example

This example shows how to create a _ROSA_ cluster with IAM User. _ROSA_ stands for Red Hat Openshift Service on AWS
and is a cluster that is created in the AWS cloud infrastructure.

 * Please NOTE: ROSA clusters built with non-STS mode (via `osdCcsAdmin` or similar) is a legacy method that will eventually be deprecated. The recommended method is to build a ROSA cluster with AWS STS.

In order to create a _ROSA_ cluster it is necessary to have user named
`osdCcsAdmin` with the `AdministratorAccess` role. The example assumes the `osdCcsAdmin`
is already created.

To run it:

* Provide OCM Authentication Token

OCM authentication token that you can get [here](https://console.redhat.com/openshift/token).

```
export TF_VAR_token=...
```

`main.tf` file and then run the `terraform apply` command.

