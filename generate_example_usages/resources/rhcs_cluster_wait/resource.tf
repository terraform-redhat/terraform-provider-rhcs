resource "rhcs_cluster_wait" "waiter" {
  cluster = "cluster-id-123"
  # timeout in minutes
  timeout = 60
}
