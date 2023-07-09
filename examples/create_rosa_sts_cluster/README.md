# ROSA STS cluster with operator IAM roles creation example

This example shows how to create an STS _ROSA_ cluster and operator IAM roles . _ROSA_ stands for Red Hat Openshift Service on AWS
and is a cluster that is created in the AWS cloud infrastructure.
In order to create an STS cluster the user also need to create a specific IAM roles called "operator IAM roles" and oidc provider. 

To run it you should create the resources in the following order: 
* Create a cluster 
* Create the operator IAM roles and oidc provider

To remove the cluster, you shouldn't remove the operator IAM roles till the cluster was successfully removed. 