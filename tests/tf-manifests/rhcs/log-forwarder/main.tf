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

# Use existing cluster
data "rhcs_cluster_rosa_hcp" "cluster" {
  id = var.cluster_id
}

# Create log forwarder on the cluster
resource "rhcs_log_forwarder" "log_forwarder" {
  cluster = data.rhcs_cluster_rosa_hcp.cluster.id
  
  s3 = {
    bucket_name   = var.s3_bucket_name
    bucket_prefix = var.s3_bucket_prefix
  }
  
  applications = var.applications
  groups       = var.groups
}

# Query all log forwarders to verify
data "rhcs_log_forwarders" "all" {
  cluster = data.rhcs_cluster_rosa_hcp.cluster.id
  depends_on = [rhcs_log_forwarder.log_forwarder]
}

output "log_forwarder_id" {
  value = rhcs_log_forwarder.log_forwarder.id
}

output "all_log_forwarders" {
  value = data.rhcs_log_forwarders.all.items
}
