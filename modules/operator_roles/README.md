# operator role module

Create rosa operator roles and identity provider in an declarative way
Terraform AWS ROSA STS Roles

In order to deploy [ROSA](https://docs.openshift.com/rosa/welcome/index.html***REMOVED*** with [STS](https://docs.openshift.com/rosa/rosa_planning/rosa-sts-aws-prereqs.html***REMOVED***, AWS Account needs to have the following roles placed:

* Account Roles (One per AWS account***REMOVED***
* OCM Roles (For OCM UI, One per OCM Org***REMOVED***
* User Role (For OCM UI, One per OCM user account***REMOVED***
* Operator Roles (One Per Cluster***REMOVED***
* OIDC Identity Provider (One Per Cluster***REMOVED***

This terraform module tries to replicate rosa CLI roles creation so that:

* Users have a declartive way to create AWS roles.
* Users can implement security/infrastructure as code practices.
* Batch creation of operator roles.

## Prerequisites

* AWS Admin Account
* OCM Account and OCM CLI
* ROSA CLI

## Get OCM Information

When creating operator IAM roles, the roles require cluster id, operator role prefix, OIDC endpoint url and thumbprint


The information can be retrieved from ocm cli.
```
ocm whoami
{
  "kind": "Account",
  "id": "26kcPSEHi0Y6MkTS7OowxfFYmZo",
  "href": "/api/accounts_mgmt/v1/accounts/26kcPSEHi0Y6MkTS7OowxfFYmZo",
  "created_at": "2022-03-22T17:43:19Z",
  "email": "shading+mobb@redhat.com",
  "first_name": "Shaozhen",
  "last_name": "Ding",
  "organization": {
    "kind": "Organization",
    "id": "1rkxPO7W12geIcRWITwI0I8VIQV",
    "href": "/api/accounts_mgmt/v1/organizations/1rkxPO7W12geIcRWITwI0I8VIQV",
    "created_at": "2021-04-27T14:31:03Z",
    "ebs_account_id": "7113273",
    "external_id": "14540493",
    "name": "Red Hat, Inc.",
    "updated_at": "2022-06-16T14:17:02Z"
  },
  "updated_at": "2022-05-25T02:02:02Z",
  "username": "shading_mobb"
}
```

## Get Clusters Information.

In order to create operator roles for clusters. Users need to provide cluster id, operator role prefix, OIDC Endpoint URL and thumbprint

```
 rosa describe cluster -c shaozhenprivate -o json
{
  "kind": "Cluster",
  "id": "1srtno3qggal8ujsegvtb2njvbmhdu8c",
  "href": "/api/clusters_mgmt/v1/clusters/1srtno3qggal8ujsegvtb2njvbmhdu8c",
  "aws": {
    "sts": {
      "oidc_endpoint_url": "https://rh-oidc.s3.us-east-1.amazonaws.com/1srtno3qggal8ujsegvtb2njvbmhdu8c",
      "operator_iam_roles": [
        {
          "id": "",
          "name": "ebs-cloud-credentials",
          "namespace": "openshift-cluster-csi-drivers",
          "role_arn": "arn:aws:iam::${AWS_ACCOUNT_ID}:role/shaozhenprivate-w4e1-openshift-cluster-csi-drivers-ebs-cloud-cre",
          "service_account": ""
        },
```

In the above example:

* cluster_id =  1srtno3qggal8ujsegvtb2njvbmhdu8c
* operator_role_prefix = shaozhenprivate-w4e1
* account_role_prefix = ManagedOpenShift
* rh_oidc_endpoint_url = rh-oidc.s3.us-east-1.amazonaws.com
* thumberprint - calculated 

## Usage

### Sample Usage

```
module operator_role {
    source = "https://github.com/openshift-online/terraform-provider-ocm/modules/ocm_cluster_rosa_classic/operator_roles"
    ## Create a list of operator roles for clusters. The cluster id and operator_role_prefix can retrieve from OCM.
    cluster_id = "1ssjjr1b2npkg9c70e8kqehfeqmscqeu"
    operator_role_prefix = "terraform-prefix"
    account_role_prefix = "ManagedOpenShift"
    rh_oidc_provider_url = "rh-oidc.s3.us-east-1.amazonaws.com/1srtno3qggal8ujsegvtb2njvbmhdu8c" 
    rh_oidc_provider_thumbprint = "917e732d330f9a12404f73d8bea36948b929dffc"
}
```
