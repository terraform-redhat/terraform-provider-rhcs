variable "cluster" {
  type        = string
  description = "Cluster ID"
}

variable "id" {
  type        = string
  description = "External auth provider ID"
}

variable "issuer_url" {
  type        = string
  description = "Token issuer URL"
}

variable "issuer_audiences" {
  type        = list(string)
  description = "List of audiences for the token issuer"
}

variable "issuer_ca" {
  type        = string
  default     = null
  description = "Certificate Authority (CA) certificate content"
}

variable "console_client_id" {
  type        = string
  default     = null
  description = "Console client ID"
}

variable "console_client_secret" {
  type        = string
  default     = null
  description = "Console client secret"
  sensitive   = true
}

variable "claim_mapping_username_key" {
  type        = string
  default     = null
  description = "Token claim to extract username from"
}

variable "claim_mapping_groups_key" {
  type        = string
  default     = null
  description = "Token claim to extract groups from"
}

variable "claim_validation_rules" {
  type = list(object({
    claim          = string
    required_value = string
  }))
  default     = []
  description = "Token claim validation rules"
}