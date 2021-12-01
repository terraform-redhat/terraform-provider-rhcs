# _CCS_ cluster creation example

This example shows how to create a _CCS_ cluster. _CCS_ stands for _costumer
cloud subscription_ and is a cluster that is created in the cloud
infrastructure (AWS account or GCP project, for example) provided by the user.

In order to create a _CCS_ cluster it is necessary to have user named
`osdCcsAdmin` with the `AdministratorAccess` role. The example uses the `aws`
provider to create that user and attach the role. To do so it uses the
credentials of the current user, typically taken from the `~/.aws/credentials`
file. See the documentation of the `aws` provider for details.

Once the `osdCcsAdmin` user is created the example uses the `aws` provider
again to generate an access key and secret for that user. That key and secret
are then used to create the cluster.

To run it use adjust the description of the provider and the cluster in the
`main.tf` file and then run the `terraform apply` command.
