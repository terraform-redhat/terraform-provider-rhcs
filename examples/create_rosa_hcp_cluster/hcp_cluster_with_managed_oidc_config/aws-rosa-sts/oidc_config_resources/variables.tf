variable "bucket_name" {
  description = "The S3 bucket name"
  type        = string
}

variable "discovery_doc" {
  description = "The discovery document string file"
  type        = string
}

variable "jwks" {
  description = "Json web key set string file"
  type        = string
}

variable "private_key" {
  description = "RSA private key"
  type        = string
}

variable "private_key_file_name" {
  description = "The private key file name"
  type        = string
}

variable "private_key_secret_name" {
  description = "The secret name that store the private key"
  type        = string
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}