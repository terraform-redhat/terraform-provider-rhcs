# operator role module

Create rosa operator roles and identity provider in an declarative way
Terraform AWS ROSA STS Roles

In order to deploy [ROSA](https://docs.openshift.com/rosa/welcome/index.html) with [STS](https://docs.openshift.com/rosa/rosa_planning/rosa-sts-aws-prereqs.html), AWS Account needs to have the following roles placed:

* Account Roles (One per AWS account)
* OCM Roles (For OCM UI, One per OCM Org)
* User Role (For OCM UI, One per OCM user account)
* Operator Roles (One Per Cluster)
* OIDC Identity Provider (One Per Cluster)

This terraform module tries to replicate rosa CLI roles creation so that:

* Users have a declartive way to create AWS roles.
* Users can implement security/infrastructure as code practices.
* Batch creation of operator roles.

## Prerequisites

* AWS Admin Account configured by using AWS CLI in AWS configuration file
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

In order to create operator roles for clusters.
Users need to provide cluster id, OIDC Endpoint URL and thumbprint and operator roles properties list.

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


The operator roles properties variable is the output of the data source `ocm_rosa_operator_roles` and it's a list of 6 maps which looks like:
```
operator_iam_roles = [
  {
    "operator_name" = "cloud-credentials"
    "operator_namespace" = "openshift-ingress-operator"
    "policy_name" = "ManagedOpenShift-openshift-ingress-operator-cloud-credentials"
    "role_arn" = "arn:aws:iam::765374464689:role/terrafom-operator-openshift-ingress-operator-cloud-credentials"
    "role_name" = "terrafom-operator-openshift-ingress-operator-cloud-credentials"
    "service_accounts" = [
      "system:serviceaccount:openshift-ingress-operator:ingress-operator",
    ]
  },
  {
    "operator_name" = "ebs-cloud-credentials"
    "operator_namespace" = "openshift-cluster-csi-drivers"
    "policy_name" = "ManagedOpenShift-openshift-cluster-csi-drivers-ebs-cloud-credent"
    "role_arn" = "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cluster-csi-drivers-ebs-cloud-creden"
    "role_name" = "terrafom-operator-openshift-cluster-csi-drivers-ebs-cloud-creden"
    "service_accounts" = [
      "system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
      "system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
    ]
  },
  {
    "operator_name" = "cloud-credentials"
    "operator_namespace" = "openshift-cloud-network-config-controller"
    "policy_name" = "ManagedOpenShift-openshift-cloud-network-config-controller-cloud"
    "role_arn" = "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-network-config-controller-clou"
    "role_name" = "terrafom-operator-openshift-cloud-network-config-controller-clou"
    "service_accounts" = [
      "system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller",
    ]
  },
  {
    "operator_name" = "aws-cloud-credentials"
    "operator_namespace" = "openshift-machine-api"
    "policy_name" = "ManagedOpenShift-openshift-machine-api-aws-cloud-credentials"
    "role_arn" = "arn:aws:iam::765374464689:role/terrafom-operator-openshift-machine-api-aws-cloud-credentials"
    "role_name" = "terrafom-operator-openshift-machine-api-aws-cloud-credentials"
    "service_accounts" = [
      "system:serviceaccount:openshift-machine-api:machine-api-controllers",
    ]
  },
  {
    "operator_name" = "cloud-credential-operator-iam-ro-creds"
    "operator_namespace" = "openshift-cloud-credential-operator"
    "policy_name" = "ManagedOpenShift-openshift-cloud-credential-operator-cloud-crede"
    "role_arn" = "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-credential-operator-cloud-cred"
    "role_name" = "terrafom-operator-openshift-cloud-credential-operator-cloud-cred"
    "service_accounts" = [
      "system:serviceaccount:openshift-cloud-credential-operator:cloud-credential-operator",
    ]
  },
  {
    "operator_name" = "installer-cloud-credentials"
    "operator_namespace" = "openshift-image-registry"
    "policy_name" = "ManagedOpenShift-openshift-image-registry-installer-cloud-creden"
    "role_arn" = "arn:aws:iam::765374464689:role/terrafom-operator-openshift-image-registry-installer-cloud-crede"
    "role_name" = "terrafom-operator-openshift-image-registry-installer-cloud-crede"
    "service_accounts" = [
      "system:serviceaccount:openshift-image-registry:cluster-image-registry-operator",
      "system:serviceaccount:openshift-image-registry:registry",
    ]
  },
]

```
## Usage

### Sample Usage

```
data "ocm_rosa_operator_roles" "operator_roles" {
  operator_role_prefix = var.operator_role_prefix
  account_role_prefix = var.account_role_prefix
}

module operator_roles {
    source  = "git::https://github.com/terraform-redhat/terraform-provider-ocm.git//modules/operator_roles"

    cluster_id = ocm_cluster_rosa_classic.rosa_sts_cluster.id
    rh_oidc_provider_thumbprint = ocm_cluster_rosa_classic.rosa_sts_cluster.sts.thumbprint
    rh_oidc_provider_url = ocm_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url
    operator_roles_properties = data.ocm_rosa_operator_roles.operator_roles.operator_iam_roles
}
```
