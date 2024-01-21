output "sg_ids" {
  description = "created SG"
  value       = module.web_server_sg.*.security_group_id
}