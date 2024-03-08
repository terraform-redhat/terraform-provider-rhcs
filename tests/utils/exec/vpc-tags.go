package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type VPCTagArgs struct {
	AWSRegion string   `json:"aws_region,omitempty"`
	IDs       []string `json:"ids,omitempty"`
	TagKey    string   `json:"key,omitempty"`
	TagValue  string   `json:"value,omitempty"`
}

type VPCTagService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newVPCTagService(clusterType CON.ClusterType, tfExec TerraformExec) (*VPCTagService, error) {
	vpctag := &VPCTagService{
		ManifestDir: GetVpcTagManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := vpctag.Init()
	return vpctag, err
}

func (vpctag *VPCTagService) Init() error {
	return vpctag.tfExec.RunTerraformInit(vpctag.ManifestDir)

}

func (vpctag *VPCTagService) Apply(createArgs *VPCTagArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = vpctag.tfExec.RunTerraformApply(vpctag.ManifestDir, tfVars)
	return
}

func (vpctag *VPCTagService) Destroy(deleteTFVars bool) (string, error) {
	return vpctag.tfExec.RunTerraformDestroy(vpctag.ManifestDir, deleteTFVars)
}
