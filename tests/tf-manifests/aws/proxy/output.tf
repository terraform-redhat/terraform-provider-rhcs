output "https_proxy" {
  value = "https://${aws_instance.tf-proxy.private_ip}:8080"
}

output "http_proxy" {
  value = "http://${aws_instance.tf-proxy.private_ip}:8080"
}

output "no_proxy" {
  value = "quay.io"
}

output "additional_trust_bundle" {
  value = data.local_file.additional_trust_bundle.content
}
