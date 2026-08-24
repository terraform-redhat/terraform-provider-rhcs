# List versions ordered by raw_id descending
data "rhcs_versions" "all" {
  order = "raw_id desc"
}

# Search for specific versions
data "rhcs_versions" "four_seventeen" {
  search = "raw_id like '4.17%'"
  order  = "raw_id desc"
}
