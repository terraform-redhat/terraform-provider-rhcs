resource "rhcs_tuning_config" "hcp_tuning_config" {
  cluster = "cluster-id-123"
  name    = "my-config"
  spec = jsonencode({
    "profile" : [
      {
        "data" : "[main]\nsummary=Custom OpenShift profile\ninclude=openshift-node\n\n[sysctl]\nvm.dirty_ratio=\"65\"\n",
        "name" : "tuned-72521-1-profile"
      }
    ],
    "recommend" : [
      {
        "priority" : 20,
        "profile" : "tuned-72521-1-profile"
      }
    ]
  })
}
