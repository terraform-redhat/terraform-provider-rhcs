terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

locals {
  defaultSpec = jsonencode(
{
    "profile": [
      {
        "data": "[main]\nsummary=Custom OpenShift profile\ninclude=openshift-node\n\n[sysctl]\nvm.dirty_ratio=\"65\"\n",
        "name": "tuned-profile"
      }
    ],
    "recommend": [
      {
        "priority": 10,
        "profile": "tuned-profile"
      }
    ]
 }
)
  spec = var.spec != null ? (var.spec != "" ? jsonencode(var.spec) : var.spec) : local.defaultSpec
}

resource "rhcs_tuning_config" "tcs" {
  count = var.tc_count

  cluster = var.cluster
  name = var.tc_count == 1 ? var.name : "${var.name}-${count.index}"
  spec = var.tc_count == 1 ? local.spec : jsonencode(
{
    "profile": [
      {
        "data": "[main]\nsummary=Custom OpenShift profile\ninclude=openshift-node\n\n[sysctl]\nvm.dirty_ratio=\"${var.spec_vm_dirty_ratios[count.index]}\"\n",
        "name": "tuned-${count.index}-profile"
      }
    ],
    "recommend": [
      {
        "priority": var.spec_priorities[count.index],
        "profile": "tuned-${count.index}-profile"
      }
    ]
 }
)
}