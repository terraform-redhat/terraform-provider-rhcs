# List all machine types
data "rhcs_machine_types" "all" {}

# Filter to a specific set of machine types (e.g. to validate instance types)
data "rhcs_machine_types" "selected" {
  search = "id in ('m5.xlarge','m5.2xlarge')"
}
