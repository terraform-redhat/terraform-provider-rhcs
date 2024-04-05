output "names" {
  value = rhcs_tuning_config.tcs[*].name
}

output "specs" {
  value = rhcs_tuning_config.tcs[*].spec
}