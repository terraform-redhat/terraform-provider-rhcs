package exec

import (
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type ImportArgs struct {
	URL          string `json:"url,omitempty"`
	ResourceKind string `json:"resource_kind,omitempty"`
	ResourceName string `json:"resource_name,omitempty"`
	ClusterID    string `json:"cluster_id,omitempty"`
	ObjectName   string `json:"obj_name,omitempty"`
}

func (args *ImportArgs) appendURL() {
	args.URL = CON.GateWayURL
}

type ImportService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newImportService(clusterType CON.ClusterType, tfExec TerraformExec) (*ImportService, error) {
	importTF := &ImportService{
		ManifestDir: GetImportManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := importTF.Init()
	return importTF, err
}

func (is *ImportService) Init() error {
	return is.tfExec.RunTerraformInit(is.ManifestDir)
}

func combineImportStructArgs(importObj *ImportArgs) []string {
	baseArgs := fmt.Sprintf("%s.%s", importObj.ResourceKind, importObj.ResourceName)

	if importObj.ResourceKind != "rhcs_cluster_rosa_classic" {
		return []string{baseArgs, fmt.Sprintf("%s,%s", importObj.ClusterID, importObj.ObjectName)}
	}

	return []string{baseArgs, importObj.ClusterID}
}

func getResourceFullName(importArgs *ImportArgs) string {
	return fmt.Sprintf("%s.%s", importArgs.ResourceKind, importArgs.ResourceName)
}

func (is *ImportService) Import(importArgs *ImportArgs) (output string, err error) {
	importArgs.appendURL()

	resourceFullName := getResourceFullName(importArgs)
	var args []string
	if importArgs.ResourceKind != "rhcs_cluster_rosa_classic" {
		args = []string{resourceFullName, fmt.Sprintf("%s,%s", importArgs.ClusterID, importArgs.ObjectName)}
	} else {
		args = []string{resourceFullName, importArgs.ClusterID}
	}

	output, err = is.tfExec.RunTerraformImport(is.ManifestDir, args...)
	return
}

func (is *ImportService) ShowState(importArgs *ImportArgs) (output string, err error) {
	args := fmt.Sprintf("%s.%s", importArgs.ResourceKind, importArgs.ResourceName)
	output, err = is.tfExec.RunTerraformState(is.ManifestDir, "show", args)
	return
}

func (is *ImportService) RemoveState(importArgs *ImportArgs) (output string, err error) {
	importArgs.appendURL()

	resourceFullName := getResourceFullName(importArgs)
	output, err = is.tfExec.RunTerraformState(is.ManifestDir, "rm", resourceFullName)
	return
}

func (is *ImportService) Destroy(deleteTFVars bool) (output string, err error) {
	return is.tfExec.RunTerraformDestroy(is.ManifestDir, deleteTFVars)
}
