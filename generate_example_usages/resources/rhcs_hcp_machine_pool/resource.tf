resource "rhcs_hcp_machine_pool" "machine_pool" {
  cluster  = "cluster-id-123"
  name     = "my-pool"
  replicas = 1
  autoscaling = {
    enabled = false
  }
  subnet_id = "subnet-id-1"
  aws_node_pool = {
    instance_type = "m5.xlarge"
  }
  auto_repair = true
}
