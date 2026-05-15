# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

resource "rhcs_hcp_default_ingress" "default_ingress" {
  cluster          = "cluster-id-123"
  listening_method = "external"
}
