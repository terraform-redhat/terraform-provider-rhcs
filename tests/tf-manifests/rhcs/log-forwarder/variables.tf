variable "cluster_id" {
  type        = string
  description = "The ID of the cluster to create the log forwarder on"
}

variable "s3_bucket_name" {
  type        = string
  description = "The name of the S3 bucket for log forwarding"
}

variable "s3_bucket_prefix" {
  type        = string
  description = "The prefix to use for objects stored in the S3 bucket"
  default     = ""
}

variable "applications" {
  type        = list(string)
  description = "List of applications to forward logs for"
  default     = []
}

variable "groups" {
  type = list(object({
    id      = string
    version = optional(string)
  }))
  description = "List of log forwarder groups"
  default     = []
}
