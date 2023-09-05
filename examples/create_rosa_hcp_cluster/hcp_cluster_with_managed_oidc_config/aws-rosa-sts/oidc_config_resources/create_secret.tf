resource "aws_secretsmanager_secret" "secret" {
  name        = var.private_key_secret_name
  description = format("Secret for %s", var.private_key_secret_name)

  tags = merge(var.tags, {
    red-hat-managed = true
  })
}

resource "aws_secretsmanager_secret_version" "store_in_secret" {
  secret_id     = aws_secretsmanager_secret.secret.id
  secret_string = var.private_key
}