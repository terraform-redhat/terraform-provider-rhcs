resource "rhcs_machine_pool" "machine_pool" {
  cluster             = "<cluster-id>"
  name                = "mpname"
  machine_type        = "r5.xlarge"
  replicas            = 3
  labels = {one="bar1bari", two ="baz2il"}
}
