output "secret_arn" {
  value = aws_secretsmanager_secret_version.store_in_secret.arn
}

