variable "cluster" {
  type        = string
  description = "The cluster ID to attach the log forwarder to"
}

variable "s3_bucket_name" {
  type        = string
  description = "S3 bucket name for log forwarding"
  default     = null
}

variable "s3_bucket_prefix" {
  type        = string
  description = "S3 bucket prefix for log forwarding"
  default     = null
}

variable "cloudwatch_log_group_name" {
  type        = string
  description = "CloudWatch log group name"
  default     = null
}

variable "cloudwatch_log_distribution_role_arn" {
  type        = string
  description = "CloudWatch log distribution role ARN"
  default     = null
}

variable "applications" {
  type        = list(string)
  description = "List of additional applications to forward logs for"
  default     = []
}

variable "groups" {
  type = list(object({
    id      = string
    version = string
  }))
  description = "List of log forwarder groups"
  default     = []
}
