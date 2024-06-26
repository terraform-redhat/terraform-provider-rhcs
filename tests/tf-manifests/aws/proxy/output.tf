output "https_proxies" {
  value = [for proxy in aws_instance.proxies : "https://${proxy.private_ip}:8080"]
}

output "http_proxies" {
  value = [for proxy in aws_instance.proxies : "http://${proxy.private_ip}:8080"]
}

output "no_proxies" {
  value = [for proxy in aws_instance.proxies : "quay.io"]
}

output "additional_trust_bundles" {
  value = [for proxy in aws_instance.proxies : data.local_file.additional_trust_bundle.content]
}
