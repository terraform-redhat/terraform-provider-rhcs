variable "cluster_id" {
  type        = string
  description = "The ID of the ROSA HCP cluster where the image mirror will be created"
}

variable "source_registry" {
  type        = string
  description = "The source registry that will be mirrored (e.g., 'docker.io/library/nginx')"
  default     = "docker.io/library/nginx"
}

variable "mirrors" {
  type        = list(string)
  description = "List of mirror registries that will serve content for the source"
  default     = ["quay.io/my-org/nginx"]
}

variable "type" {
  type        = string
  description = "The type of image mirror (only 'digest' is currently supported)"
  default     = "digest"
}