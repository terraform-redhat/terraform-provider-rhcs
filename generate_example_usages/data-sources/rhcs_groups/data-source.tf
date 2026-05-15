# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

data "rhcs_groups" "groups" {
    cluster = "cluster-id-123"
}