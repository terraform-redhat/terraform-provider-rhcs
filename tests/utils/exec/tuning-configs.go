package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type TuningConfigArgs struct {
	Cluster           *string `json:"cluster,omitempty"`
	Name              *string `json:"name,omitempty"`
	URL               string  `json:"url,omitempty"`
	Count             *int    `json:"tc_count,omitempty"`
	Spec              *string `json:"spec,omitempty"`
	SpecVMDirtyRatios *[]int  `json:"spec_vm_dirty_ratios,omitempty"`
	SpecPriorities    *[]int  `json:"spec_priorities,omitempty"`
}

type TuningConfigService struct {
	CreationArgs *TuningConfigArgs
	ManifestDir  string
	Context      context.Context
}

type TuningConfigOutput struct {
	Names []string `json:"names,omitempty"`
}

func (tcs *TuningConfigService) Init(manifestDirs ...string) error {
	tcs.ManifestDir = CON.TuningConfigDir
	if len(manifestDirs) != 0 {
		tcs.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	tcs.Context = ctx
	err := runTerraformInit(ctx, tcs.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (tcs *TuningConfigService) Apply(createArgs *TuningConfigArgs, recordtfargs bool, extraArgs ...string) (string, error) {
	createArgs.URL = CON.GateWayURL
	tcs.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	output, err := runTerraformApply(tcs.Context, tcs.ManifestDir, args...)
	if err == nil && recordtfargs {
		recordTFvarsFile(tcs.ManifestDir, tfvars)
	}
	return output, err
}

func (tcs *TuningConfigService) Output() (output *TuningConfigOutput, err error) {
	mpDir := tcs.ManifestDir
	if tcs.ManifestDir == "" {
		mpDir = CON.ClassicMachinePoolDir
	}
	out, err := runTerraformOutput(context.TODO(), mpDir)
	if err != nil {
		return output, err
	}
	output = &TuningConfigOutput{
		Names: h.DigArrayToString(out["names"], "value"),
	}
	return output, nil
}

func (tcs *TuningConfigService) Destroy(createArgs ...*TuningConfigArgs) (output string, err error) {
	if tcs.CreationArgs == nil && len(createArgs) == 0 {
		return "", fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := tcs.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	destroyArgs.URL = CON.GateWayURL
	args, _ := combineStructArgs(destroyArgs)

	return runTerraformDestroy(tcs.Context, tcs.ManifestDir, args...)
}

func NewTuningConfigService(manifestDir ...string) *TuningConfigService {
	mp := &TuningConfigService{}
	mp.Init(manifestDir...)
	return mp
}
