---
page_title: "Hosted Control Plane (HCP) clusters"
subcategory: ""
description: |-
  Instructions on how to interact with HCP clusters.
---

# Create a new cluster

Hosted Control Plane clusters differ from Classic managed OpenShift clusters, the main difference relates a decoupled management plane that enables consolidated control and management of core control plane components. Having the control plane hosted and managed in a ROSA service AWS account rather than in the customer’s individual account offers advantages to aid your business. Reduces complications and cost from scaling the control plane and facilitates workload scheduling to worker nodes. Meaning developers can spend more time developing and testing applications, instead of configuring or waiting for cluster infrastructure to be ready.

Red Hat (RH) provides the access to the management of HCP clusters as well as it's resources through Terraform. Which allows infrastructure to be managed as code, making it much more reliable and reproducible.

The main resource for HCP clusters in Terraform Red Hat Clusters Service (RHCS) Provider is the `rhcs_cluster_rosa_hcp`, this will provision a HCP cluster alongside a machine pool for customer workload.

```terraform
locals {
  sts_roles = {
    role_arn         = "arn:aws:iam::account-id-123:role/account-prefix-HCP-ROSA-Installer-Role",
    support_role_arn = "arn:aws:iam::account-id-123:role/account-prefix-HCP-ROSA-Support-Role",
    instance_iam_roles = {
      worker_role_arn = "arn:aws:iam::account-id-123:role/account-prefix-HCP-ROSA-Worker-Role"
    },
    operator_role_prefix = "operator-prefix",
    oidc_config_id       = "oidc-config-id-123"
  }
}


resource "rhcs_cluster_rosa_hcp" "rosa_sts_cluster" {
  name                   = "my-cluster"
  cloud_region           = "us-east-2"
  aws_account_id         = "123456789012"
  aws_billing_account_id = "123456789012"
  aws_subnet_ids         = ["subnet-1", "subnet-2"]
  availability_zones     = ["us-west-2a", "us-west-2b"]
  replicas               = 2
  version                = "4.15.9"
  properties = {
    rosa_creator_arn = "aws_caller_identity-current-arn"
  }
  sts                                 = local.sts_roles
  wait_for_create_complete            = true
  wait_for_std_compute_nodes_complete = true
}
```

It requires some other cloud resources, such as a VPC so that the subnets and availability zones can be forwarded to the cluster network infrastructure. As well as STS credentials for each of the account and operator roles. And last but not least an OIDC configuration/provider to authenticate the operator roles access to the AWS account it belongs to.
These resources require some interaction with AWS Provider as well, for a basic VPC it will be needed to

```
variable "vpc_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "Cidr block of the desired VPC. This value should not be updated, please create a new resource instead"
}

variable "name_prefix" {
  type        = string
  description = "User-defined prefix for all generated AWS resources of this VPC. This value should not be updated, please create a new resource instead"
}

variable "availability_zones_count" {
  type        = number
  description = "The count of availability zones to utilize within the specified AWS Region, where pairs of public and private subnets will be generated. This value should not be updated, please create a new resource instead"
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "AWS tags to be applied to generated AWS resources of this VPC."
}

locals {
  tags = var.tags == null ? {} : var.tags
}

resource "aws_vpc" "vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = merge(
    {
      "Name" = "${var.name_prefix}-vpc"
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.vpc.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

resource "aws_subnet" "public_subnet" {
  count = var.availability_zones_count

  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(var.vpc_cidr, var.availability_zones_count * 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]
  tags = merge(
    {
      "Name" = join("-", [var.name_prefix, "subnet", "public${count.index + 1}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
      "kubernetes.io/role/elb" = ""
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_subnet" "private_subnet" {
  count = var.availability_zones_count

  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(var.vpc_cidr, var.availability_zones_count * 2, count.index + var.availability_zones_count)
  availability_zone = data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]
  tags = merge(
    {
      "Name" = join("-", [var.name_prefix, "subnet", "private${count.index + 1}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
      "kubernetes.io/role/internal-elb" = ""
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Internet gateway
#
resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = "${var.name_prefix}-igw"
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Elastic IPs for NAT gateways
#
resource "aws_eip" "eip" {
  count = var.availability_zones_count

  domain = "vpc"
  tags = merge(
    {
      "Name" = join("-", [var.name_prefix, "eip", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# NAT gateways
#
resource "aws_nat_gateway" "public_nat_gateway" {
  count = var.availability_zones_count

  allocation_id = aws_eip.eip[count.index].id
  subnet_id     = aws_subnet.public_subnet[count.index].id

  tags = merge(
    {
      "Name" = join("-", [var.name_prefix, "nat", "public${count.index}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Route tables
#
resource "aws_route_table" "public_route_table" {
  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = "${var.name_prefix}-public"
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_route_table" "private_route_table" {
  count = var.availability_zones_count

  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = join("-", [var.name_prefix, "rtb", "private${count.index}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Routes
#
# Send all IPv4 traffic to the internet gateway
resource "aws_route" "ipv4_egress_route" {
  route_table_id         = aws_route_table.public_route_table.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gateway.id
  depends_on             = [aws_route_table.public_route_table]
}

# Send all IPv6 traffic to the internet gateway
resource "aws_route" "ipv6_egress_route" {
  route_table_id              = aws_route_table.public_route_table.id
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.internet_gateway.id
  depends_on                  = [aws_route_table.public_route_table]
}

# Send private traffic to NAT
resource "aws_route" "private_nat" {
  count = var.availability_zones_count

  route_table_id         = aws_route_table.private_route_table[count.index].id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.public_nat_gateway[count.index].id
  depends_on             = [aws_route_table.private_route_table, aws_nat_gateway.public_nat_gateway]
}


# Private route for vpc endpoint
resource "aws_vpc_endpoint_route_table_association" "private_vpc_endpoint_route_table_association" {
  count = var.availability_zones_count

  route_table_id  = aws_route_table.private_route_table[count.index].id
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

#
# Route table associations
#
resource "aws_route_table_association" "public_route_table_association" {
  count = var.availability_zones_count

  subnet_id      = aws_subnet.public_subnet[count.index].id
  route_table_id = aws_route_table.public_route_table.id
}

resource "aws_route_table_association" "private_route_table_association" {
  count = var.availability_zones_count

  subnet_id      = aws_subnet.private_subnet[count.index].id
  route_table_id = aws_route_table.private_route_table[count.index].id
}

# This resource is used in order to add dependencies on all resources 
# Any resource uses this VPC ID, must wait to all resources creation completion
resource "time_sleep" "vpc_resources_wait" {
  create_duration = "20s"
  destroy_duration = "20s"
  triggers = {
    vpc_id                                           = aws_vpc.vpc.id
    cidr_block                                       = aws_vpc.vpc.cidr_block
    ipv4_egress_route_id                             = aws_route.ipv4_egress_route.id
    ipv6_egress_route_id                             = aws_route.ipv6_egress_route.id
    private_nat_ids                                  = jsonencode([for value in aws_route.private_nat : value.id])
    private_vpc_endpoint_route_table_association_ids = jsonencode([for value in aws_vpc_endpoint_route_table_association.private_vpc_endpoint_route_table_association : value.id])
    public_route_table_association_ids               = jsonencode([for value in aws_route_table_association.public_route_table_association : value.id])
    private_route_table_association_ids              = jsonencode([for value in aws_route_table_association.private_route_table_association : value.id])
  }
}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"

  # New configuration to exclude Local Zones
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
```

And then for account and operator roles, those require some extra thought. Each of the roles have a permission policy, which, as the name implies, have a set of permissions that allow the cluster to interact with different AWS services. In HCP the permission policies are all AWS managed, which ensures both Red Hat and AWS reviews the permissions and conditions to ensure more security to your clusters.

To create the account roles it is needed to:

```
variable "account_role_prefix" {
  type    = string
  description = "Prefix to be used when creating the account roles"
  default = "tf-acc"
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies. Must begin and end with '/'."
  type        = string
  default     = "/"
}

variable "permissions_boundary" {
  description = "The ARN of the policy that is used to set the permissions boundary for the IAM roles in STS clusters."
  type        = string
  default     = ""
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}

locals {
  path = coalesce(var.path, "/")
  account_roles_properties = [
    {
      role_name            = "HCP-ROSA-Installer"
      role_type            = "installer"
      policy_details       = "arn:aws:iam::aws:policy/service-role/ROSAInstallerPolicy"
      principal_type       = "AWS"
      principal_identifier = "arn:${data.aws_partition.current.partition}:iam::${data.rhcs_info.current.ocm_aws_account_id}:role/RH-Managed-OpenShift-Installer"
    },
    {
      role_name            = "HCP-ROSA-Support"
      role_type            = "support"
      policy_details       = "arn:aws:iam::aws:policy/service-role/ROSASRESupportPolicy"
      principal_type       = "AWS"
      // This is a SRE RH Support role which is used to assume this support role
      principal_identifier = data.rhcs_hcp_policies.all_policies.account_role_policies["sts_support_rh_sre_role"]
    },
    {
      role_name            = "HCP-ROSA-Worker"
      role_type            = "instance_worker"
      policy_details       = "arn:aws:iam::aws:policy/service-role/ROSAWorkerInstancePolicy"
      principal_type       = "Service"
      principal_identifier = "ec2.amazonaws.com"
    },
  ]
  account_roles_count = length(local.account_roles_properties)
  account_role_prefix_valid = var.account_role_prefix != null ? (
    var.account_role_prefix
    ) : (
    "account-role-${random_string.default_random[0].result}"
  )
}

data "aws_iam_policy_document" "custom_trust_policy" {
  count = local.account_roles_count

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = local.account_roles_properties[count.index].principal_type
      identifiers = [local.account_roles_properties[count.index].principal_identifier]
    }
  }
}

resource "aws_iam_role" "account_role" {
  count                = local.account_roles_count
  name                 = substr("${local.account_role_prefix_valid}-${local.account_roles_properties[count.index].role_name}-Role", 0, 64)
  permissions_boundary = var.permissions_boundary
  path                 = local.path
  assume_role_policy   = data.aws_iam_policy_document.custom_trust_policy[count.index].json

  tags = merge(var.tags, {
    red-hat-managed       = true
    rosa_hcp_policies     = true
    rosa_managed_policies = true
    rosa_role_prefix      = local.account_role_prefix_valid
    rosa_role_type        = local.account_roles_properties[count.index].role_type
  })
}

resource "aws_iam_role_policy_attachment" "account_role_policy_attachment" {
  count      = local.account_roles_count
  role       = aws_iam_role.account_role[count.index].name
  policy_arn = local.account_roles_properties[count.index].policy_details
}

resource "random_string" "default_random" {
  count = var.account_role_prefix != null ? 0 : 1

  length  = 4
  special = false
  upper   = false
}

data "rhcs_hcp_policies" "all_policies" {}

data "aws_partition" "current" {}

data "rhcs_info" "current" {}

resource "time_sleep" "account_iam_resources_wait" {
  destroy_duration = "10s"
  create_duration  = "10s"
  triggers = {
    account_iam_role_name = jsonencode([ for value in aws_iam_role.account_role : value.name])
    account_roles_arn     = jsonencode({ for idx, value in aws_iam_role.account_role : local.account_roles_properties[idx].role_name => value.arn })
    account_policy_arns   = jsonencode([ for value in aws_iam_role_policy_attachment.account_role_policy_attachment : value.policy_arn])
    account_role_prefix   = local.account_role_prefix_valid
    path                  = local.path
  }
}
```

The operator roles, depends first on the OIDC configuration, which can be setup in two flows, a Red Hat managed configuration or a customer managed configuration.
For a RH managed configuration it is as simple as:
```
variable "tags" {
  type        = map(string)
  default     = null
  description = "List of AWS resource tags to apply."
}

resource "rhcs_rosa_oidc_config" "oidc_config" {
  managed            = true
}

resource "aws_iam_openid_connect_provider" "oidc_provider" {
  url = "https://${rhcs_rosa_oidc_config.oidc_config.oidc_endpoint_url}"

  client_id_list = [
    "openshift",
    "sts.amazonaws.com"
  ]

  tags = var.tags

  thumbprint_list = [rhcs_rosa_oidc_config.oidc_config.thumbprint]
}

data "aws_region" "current" {}

resource "time_sleep" "wait_10_seconds" {
  create_duration  = "10s"
  destroy_duration = "10s"
  triggers = {
    oidc_config_id                         = rhcs_rosa_oidc_config.oidc_config.id
    oidc_endpoint_url                      = rhcs_rosa_oidc_config.oidc_config.oidc_endpoint_url
    oidc_provider_url                      = aws_iam_openid_connect_provider.oidc_provider.url
  }
}
```

However, the customer configuration requires an S3 bucket, alongside the stored files for the OIDC protocol compliancy, and an AWS secret, which will store a private key that is forwarded to RH for cluster setup.
```
variable "installer_role_arn" {
  type        = string
  default     = null
  description = "The Amazon Resource Name (ARN) associated with the AWS IAM role used by the ROSA installer. Applicable exclusively to unmanaged OIDC; otherwise, leave empty."
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "List of AWS resource tags to apply."
}

resource "rhcs_rosa_oidc_config" "oidc_config" {
  managed            = false
  secret_arn         = aws_secretsmanager_secret.secret[0].arn
  issuer_url         = rhcs_rosa_oidc_config_input.oidc_input[0].issuer_url
  installer_role_arn = var.installer_role_arn
}

resource "aws_iam_openid_connect_provider" "oidc_provider" {
  url = "https://${rhcs_rosa_oidc_config.oidc_config.oidc_endpoint_url}"

  client_id_list = [
    "openshift",
    "sts.amazonaws.com"
  ]

  tags = var.tags

  thumbprint_list = [rhcs_rosa_oidc_config.oidc_config.thumbprint]
}

resource "aws_s3_bucket" "s3_bucket" {
  bucket = rhcs_rosa_oidc_config_input.oidc_input[count.index].bucket_name

  tags = merge(var.tags, {
    red-hat-managed = true
  })
}

resource "aws_s3_bucket_public_access_block" "public_access_block" {
  bucket = aws_s3_bucket.s3_bucket[count.index].id

  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = false
  restrict_public_buckets = false
}

data "aws_iam_policy_document" "allow_access_from_another_account" {

  statement {
    principals {
      identifiers = ["*"]
      type        = "*"
    }
    sid    = "AllowReadPublicAccess"
    effect = "Allow"
    actions = [
      "s3:GetObject",
    ]

    resources = [
      format("arn:aws:s3:::%s/*", rhcs_rosa_oidc_config_input.oidc_input[count.index].bucket_name),
    ]
  }
}

resource "aws_s3_bucket_policy" "allow_access_from_another_account" {
  bucket = aws_s3_bucket.s3_bucket[count.index].id
  policy = data.aws_iam_policy_document.allow_access_from_another_account[count.index].json
}

resource "rhcs_rosa_oidc_config_input" "oidc_input" {

  region = data.aws_region.current.name
}

resource "aws_secretsmanager_secret" "secret" {
  name        = rhcs_rosa_oidc_config_input.oidc_input[count.index].private_key_secret_name
  description = format("Secret for %s", rhcs_rosa_oidc_config_input.oidc_input[count.index].private_key_secret_name)

  tags = merge(var.tags, {
    red-hat-managed = true
  })
}

resource "aws_secretsmanager_secret_version" "store_in_secret" {
  secret_id     = aws_secretsmanager_secret.secret[count.index].id
  secret_string = rhcs_rosa_oidc_config_input.oidc_input[count.index].private_key
}

resource "aws_s3_object" "discover_doc_object" {

  bucket       = aws_s3_bucket.s3_bucket[count.index].id
  key          = ".well-known/openid-configuration"
  content      = rhcs_rosa_oidc_config_input.oidc_input[count.index].discovery_doc
  content_type = "application/json"

  tags = merge(var.tags, {
    red-hat-managed = true
  })
}

resource "aws_s3_object" "jwks_object" {

  bucket       = aws_s3_bucket.s3_bucket[count.index].id
  key          = "keys.json"
  content      = rhcs_rosa_oidc_config_input.oidc_input[count.index].jwks
  content_type = "application/json"

  tags = merge(var.tags, {
    red-hat-managed = true
  })
}

data "aws_region" "current" {}

resource "time_sleep" "wait_10_seconds" {
  create_duration  = "10s"
  destroy_duration = "10s"
  triggers = {
    oidc_config_id                         = rhcs_rosa_oidc_config.oidc_config.id
    oidc_endpoint_url                      = rhcs_rosa_oidc_config.oidc_config.oidc_endpoint_url
    oidc_provider_url                      = aws_iam_openid_connect_provider.oidc_provider.url
    discover_doc_object                    = aws_s3_object.discover_doc_object[0].checksum_sha1
    s3_object                              = aws_s3_object.jwks_object[0].checksum_sha1
    policy_attached_to_bucket              = aws_s3_bucket_policy.allow_access_from_another_account[0].bucket
    public_access_block_attached_to_bucket = aws_s3_bucket_public_access_block.public_access_block[0].bucket
    secret_arn                             = aws_secretsmanager_secret.secret[0].arn
  }
}
```
For the actual operator roles:
```
variable "operator_role_prefix" {
  type = string
  description = "Prefix to be used when creating the operator roles"
  default = "tf-op"
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies. Must begin and end with '/'."
  type        = string
  default     = "/"
}

variable "permissions_boundary" {
  description = "The ARN of the policy that is used to set the permissions boundary for the IAM roles in STS clusters."
  type        = string
  default     = ""
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}

variable "oidc_endpoint_url" {
  description = "oidc provider url"
  type        = string
}

locals {
  operator_roles_properties = [
    {
      operator_name      = "installer-cloud-credentials"
      operator_namespace = "openshift-image-registry"
      role_name          = "openshift-image-registry-installer-cloud-credentials"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSAImageRegistryOperatorPolicy"
      service_accounts   = ["system:serviceaccount:openshift-image-registry:cluster-image-registry-operator", "system:serviceaccount:openshift-image-registry:registry"]
    },
    {
      operator_name      = "cloud-credentials"
      operator_namespace = "openshift-ingress-operator"
      role_name          = "openshift-ingress-operator-cloud-credentials"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSAIngressOperatorPolicy"
      service_accounts   = ["system:serviceaccount:openshift-ingress-operator:ingress-operator"]
    },
    {
      operator_name      = "ebs-cloud-credentials"
      operator_namespace = "openshift-cluster-csi-drivers"
      role_name          = "openshift-cluster-csi-drivers-ebs-cloud-credentials"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSAAmazonEBSCSIDriverOperatorPolicy"
      service_accounts   = ["system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator", "system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa"]
    },
    {
      operator_name      = "cloud-credentials"
      operator_namespace = "openshift-cloud-network-config-controller"
      role_name          = "openshift-cloud-network-config-controller-cloud-credentials"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSACloudNetworkConfigOperatorPolicy"
      service_accounts   = ["system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller"]
    },
    {
      operator_name      = "kube-controller-manager"
      operator_namespace = "kube-system"
      role_name          = "kube-system-kube-controller-manager"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSAKubeControllerPolicy"
      service_accounts   = ["system:serviceaccount:kube-system:kube-controller-manager"]
    },
    {
      operator_name      = "capa-controller-manager"
      operator_namespace = "kube-system"
      role_name          = "kube-system-capa-controller-manager"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSANodePoolManagementPolicy"
      service_accounts   = ["system:serviceaccount:kube-system:capa-controller-manager"]
    },
    {
      operator_name      = "control-plane-operator"
      operator_namespace = "kube-system"
      role_name          = "kube-system-control-plane-operator"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSAControlPlaneOperatorPolicy"
      service_accounts   = ["system:serviceaccount:kube-system:control-plane-operator"]
    },
    {
      operator_name      = "kms-provider"
      operator_namespace = "kube-system"
      role_name          = "kube-system-kms-provider"
      policy_details     = "arn:aws:iam::aws:policy/service-role/ROSAKMSProviderPolicy"
      service_accounts   = ["system:serviceaccount:kube-system:kms-provider"]
    },
  ]
  operator_roles_count = length(local.operator_roles_properties)
  operator_role_prefix = var.operator_role_prefix
  path                 = coalesce(var.path, "/")
}

data "aws_iam_policy_document" "custom_trust_policy" {
  count = local.operator_roles_count

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${var.oidc_endpoint_url}"]
    }
    condition {
      test     = "StringEquals"
      variable = "${var.oidc_endpoint_url}:sub"
      values   = local.operator_roles_properties[count.index].service_accounts
    }
  }
}

resource "aws_iam_role" "operator_role" {
  count                = local.operator_roles_count
  name                 = substr("${local.operator_role_prefix}-${local.operator_roles_properties[count.index].operator_namespace}-${local.operator_roles_properties[count.index].operator_name}", 0, 64)
  permissions_boundary = var.permissions_boundary
  path                 = local.path
  assume_role_policy   = data.aws_iam_policy_document.custom_trust_policy[count.index].json

  tags = merge(var.tags, {
    rosa_managed_policies = true
    rosa_hcp_policies     = true
    red-hat-managed       = true
    operator_namespace    = local.operator_roles_properties[count.index].operator_namespace
    operator_name         = local.operator_roles_properties[count.index].operator_name
  })
}

resource "aws_iam_role_policy_attachment" "operator_role_policy_attachment" {
  count      = local.operator_roles_count
  role       = aws_iam_role.operator_role[count.index].name
  policy_arn = local.operator_roles_properties[count.index].policy_details
}

data "aws_caller_identity" "current" {}

resource "time_sleep" "role_resources_propagation" {
  create_duration = "20s"
  triggers = {
    operator_role_prefix = local.operator_role_prefix
    operator_role_arns   = jsonencode([for value in aws_iam_role.operator_role : value.arn])
    operator_policy_arns = jsonencode([for value in aws_iam_role_policy_attachment.operator_role_policy_attachment : value.policy_arn])
  }
}
```

All of these operations are covered by Terraform Hosted Control Plane Modules, which aims to aid customers in having a supported and verified manner to manage resources which requires a more complex setup. Of course extra customization from the customer side is valid and may require not depending on the modules.