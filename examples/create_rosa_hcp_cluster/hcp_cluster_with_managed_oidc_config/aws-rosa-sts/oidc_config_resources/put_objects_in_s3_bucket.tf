#store in S3
resource "aws_s3_object" "discrover_doc_object" {
  bucket       = aws_s3_bucket.s3_bucket.bucket
  key          = ".well-known/openid-configuration"
  content      = var.discovery_doc
  content_type = "application/json"

  tags = merge(var.tags, {
    red-hat-managed = true
  })
}

resource "aws_s3_object" "jwks_object" {
  bucket       = aws_s3_bucket.s3_bucket.bucket
  key          = "keys.json"
  content      = var.jwks
  content_type = "application/json"

  tags = merge(var.tags, {
    red-hat-managed = true
  })
}
