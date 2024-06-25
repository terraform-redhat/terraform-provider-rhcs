package manifests

import "path"

func GrantTFstateFile(manifestDir string) string {
	return path.Join(manifestDir, "terraform.tfstate")
}
