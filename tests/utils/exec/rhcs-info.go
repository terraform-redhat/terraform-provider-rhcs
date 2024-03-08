package exec

import (
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type RhcsInfoArgs struct {
	ID           string `json:"id,omitempty"`
	ResourceName string `json:"resource_name,omitempty"`
	ResourceKind string `json:"resource_kind,omitempty"`
}

type RhcsInfoService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newRhcsInfoService(clusterType CON.ClusterType, tfExec TerraformExec) (*RhcsInfoService, error) {
	rhcsInfo := &RhcsInfoService{
		ManifestDir: GetRhcsInfoManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := rhcsInfo.Init()
	return rhcsInfo, err
}

func (rhcsInfo *RhcsInfoService) Init() error {
	return rhcsInfo.tfExec.RunTerraformInit(rhcsInfo.ManifestDir)
}

func (rhcsInfo *RhcsInfoService) Apply(createArgs *RhcsInfoArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = rhcsInfo.tfExec.RunTerraformApply(rhcsInfo.ManifestDir, tfVars)
	return
}

func (rhcsInfo *RhcsInfoService) Destroy(deleteTFVars bool) (string, error) {
	return rhcsInfo.tfExec.RunTerraformDestroy(rhcsInfo.ManifestDir, deleteTFVars)
}

func (rhcsInfoService *RhcsInfoService) ShowState(rhcsInfoArgs *RhcsInfoArgs) (output string, err error) {
	args := fmt.Sprintf("data.%s.%s", rhcsInfoArgs.ResourceKind, rhcsInfoArgs.ResourceName)
	output, err = rhcsInfoService.tfExec.RunTerraformState(rhcsInfoService.ManifestDir, "show", args)
	return
}
