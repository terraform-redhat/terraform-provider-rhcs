variable vpc_cidr{
    type=string
    default = "10.0.0.0/16"
    validation {
      condition     = contains(["10.0.0.0/16", "11.0.0.0/16", "12.0.0.0/16"], var.vpc_cidr***REMOVED***
      error_message = "VPC CIDR limited to: 10.0.0.0/16 11.0.0.0/16 12.0.0.0/16"
    }
    
}
variable multi_az{
    type = bool
    default = false
}
variable "aws_region" {
  type        = string
  description = "The region to create the ROSA cluster in"
}
variable "test_zones"{
  type= list(string***REMOVED***
  
}

variable "az_ids" {
  type        = object({
    eu-west-1 = list(string***REMOVED***
    us-east-1 = list(string***REMOVED***
    us-east-2 = list(string***REMOVED***
    us-west-2 = list(string***REMOVED***
  }***REMOVED***
  description = "A list of region-mapped AZ IDs that a subnet should get deployed into"
  default     = {
    eu-central-1 = ["eu-central-1a", "eu-central-1b","eu-central-1c"]
    eu-west-1 = ["eu-west-1a", "eu-west-1b","eu-west-1c"]
    us-east-1 = ["us-east-1a", "us-east-1b","us-east-1c"]
    us-east-2 = ["us-east-2a", "us-east-2b","us-east-2c"]
    us-west-2 = ["us-west-2a", "us-west-2b","us-west-2c"]
  }
}

variable "cluster_name" {
    type        = string
    description = "The name of the ROSA cluster to create"
    default     = "tf-ocm"
  
}