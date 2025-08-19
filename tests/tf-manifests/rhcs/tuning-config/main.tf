terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0-0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

resource "rhcs_tuning_config" "tcs" {
  count = var.tc_count

  cluster = var.cluster
  name    = var.tc_count == 1 ? var.name : "${var.name}-${count.index}"
  spec    = var.specs[count.index].spec_type == "file" ? file(var.specs[count.index].spec_value) : var.specs[count.index].spec_value
}
