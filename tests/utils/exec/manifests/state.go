// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package manifests

import "path"

func GrantTFstateFile(manifestDir string) string {
	return path.Join(manifestDir, "terraform.tfstate")
}
