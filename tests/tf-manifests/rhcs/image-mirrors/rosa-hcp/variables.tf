variable "cluster" {
  type        = string
  description = "The cluster ID where the image mirror will be created"
}

variable "type" {
  type        = string
  description = "The type of image mirror (only 'digest' is currently supported)"
  default     = "digest"
}

variable "source_registry" {
  type        = string
  description = "The source registry that will be mirrored"
}

variable "mirrors" {
  type        = list(string)
  description = "List of mirror registries that will serve content for the source"
}