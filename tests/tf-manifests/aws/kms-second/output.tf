output "arn" {
  description = "The ARN of the key"
  value       = aws_kms_key.cluster_kms_key.arn
}