package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type KubeletConfigArgs struct {
	URL          string `json:"url,omitempty"`
	Cluster      string `json:"cluster,omitempty"`
	PodPidsLimit int    `json:"pod_pids_limit,omitempty"`
}

func (args *KubeletConfigArgs) appendURL() {
	args.URL = CON.GateWayURL
}

type KubeletConfigOutput struct {
	Cluster      string `json:"cluster,omitempty"`
	PodPidsLimit int    `json:"pod_pids_limit,omitempty"`
}

type KubeletConfigService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newKubeletConfigService(clusterType CON.ClusterType, tfExec TerraformExec) (*KubeletConfigService, error) {
	kc := &KubeletConfigService{
		ManifestDir: GetKubeletConfigManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := kc.Init()
	return kc, err
}

func (kc *KubeletConfigService) Init() error {
	return kc.tfExec.RunTerraformInit(kc.ManifestDir)

}

func (kc *KubeletConfigService) Apply(createArgs *KubeletConfigArgs) (output string, err error) {
	createArgs.appendURL()

	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = kc.tfExec.RunTerraformApply(kc.ManifestDir, tfVars)
	return
}
func (kc *KubeletConfigService) Plan(planArgs *KubeletConfigArgs) (output string, err error) {
	planArgs.appendURL()

	var tfVars *TFVars
	tfVars, err = NewTFArgs(planArgs)
	if err != nil {
		return
	}
	output, err = kc.tfExec.RunTerraformPlan(kc.ManifestDir, tfVars)
	return
}

func (kc *KubeletConfigService) Output() (*KubeletConfigOutput, error) {
	out, err := kc.tfExec.RunTerraformOutput(kc.ManifestDir)
	if err != nil {
		return nil, err
	}
	var accOutput = &KubeletConfigOutput{
		PodPidsLimit: h.DigInt(out["pod_pids_limit"], "value"),
	}
	return accOutput, nil
}

func (kc *KubeletConfigService) Destroy(deleteTFVars bool) (string, error) {
	return kc.tfExec.RunTerraformDestroy(kc.ManifestDir, deleteTFVars)
}
