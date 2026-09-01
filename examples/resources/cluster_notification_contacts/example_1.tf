resource "rhcs_cluster_notification_contacts" "contacts" {
  cluster_id = "cluster-id-123"
  contacts   = ["ocm-username-1", "ocm-username-2"]
}
