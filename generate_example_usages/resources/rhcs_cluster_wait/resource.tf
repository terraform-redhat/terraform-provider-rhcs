resource "rhcs_cluster_wait" "waiter" {
  cluster = "<cluster-id>"
  # timeout in minutes
  timeout = 60
}
