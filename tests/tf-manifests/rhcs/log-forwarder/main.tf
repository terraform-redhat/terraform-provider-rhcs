terraform {
  required_providers {
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

resource "rhcs_log_forwarder" "log_forwarder" {
  cluster = var.cluster

  # S3 Configuration (each log forwarder can only use one of s3 or cloudwatch)
  s3 = var.s3_bucket_name != null ? {
    bucket_name   = var.s3_bucket_name
    bucket_prefix = var.s3_bucket_prefix
  } : null

  # CloudWatch Configuration (each log forwarder can only use one of s3 or cloudwatch)
  cloudwatch = var.cloudwatch_log_group_name != null ? {
    log_group_name            = var.cloudwatch_log_group_name
    log_distribution_role_arn = var.cloudwatch_log_distribution_role_arn
  } : null

  applications = var.applications
  groups       = var.groups
}

output "log_forwarder_id" {
  value       = rhcs_log_forwarder.log_forwarder.id
  description = "The log forwarder ID"
}
