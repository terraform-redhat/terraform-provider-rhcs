terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  url = var.url
}

resource "rhcs_tuning_config" "tcs" {
  count = var.tc_count

  cluster = var.cluster
  name = "${var.name_prefix}-${count.index}"
  spec = jsonencode(
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