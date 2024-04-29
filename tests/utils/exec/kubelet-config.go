package exec

import (
	"context"
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type KubeletConfigArgs struct {
	Cluster      string `json:"cluster,omitempty"`
	PodPidsLimit int    `json:"pod_pids_limit,omitempty"`
}

type KubeletConfigOutput struct {
	Cluster      string `json:"cluster,omitempty"`
	PodPidsLimit int    `json:"pod_pids_limit,omitempty"`
}

type KubeletConfigService struct {
	CreationArgs *KubeletConfigArgs
	ManifestDir  string
	Context      context.Context
}

func (kc *KubeletConfigService) Init(manifestDirs ...string) error {
	kc.ManifestDir = constants.KubeletConfigDir
	if len(manifestDirs) != 0 {
		kc.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	kc.Context = ctx
	err := runTerraformInit(ctx, kc.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (kc *KubeletConfigService) Apply(createArgs *KubeletConfigArgs, recordtfvars bool, extraArgs ...string) (*KubeletConfigOutput, error) {
	kc.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(kc.Context, kc.ManifestDir, args...)
	if err != nil {
		return nil, err
	}
	if recordtfvars {
		recordTFvarsFile(kc.ManifestDir, tfvars)
	}
	output, err := kc.Output()
	return output, err
}
func (kc *KubeletConfigService) Plan(createArgs *KubeletConfigArgs, extraArgs ...string) (string, error) {
	kc.CreationArgs = createArgs
	args, _ := combineStructArgs(createArgs, extraArgs...)
	output, err := runTerraformPlan(kc.Context, kc.ManifestDir, args...)

	return output, err
}
func (kc *KubeletConfigService) Output() (*KubeletConfigOutput, error) {
	out, err := runTerraformOutput(kc.Context, kc.ManifestDir)
	if err != nil {
		return nil, err
	}
	var accOutput = &KubeletConfigOutput{
		PodPidsLimit: helper.DigInt(out["pod_pids_limit"], "value"),
	}
	return accOutput, nil
}

func (kc *KubeletConfigService) Destroy(createArgs ...*KubeletConfigArgs) (string, error) {
	if kc.CreationArgs == nil && len(createArgs) == 0 {
		return "", fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := kc.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	output, err := runTerraformDestroy(kc.Context, kc.ManifestDir, args...)
	return output, err
}

func NewKubeletConfigService(manifestDir ...string) (*KubeletConfigService, error) {
	kc := &KubeletConfigService{}
	err := kc.Init(manifestDir...)
	return kc, err
}
