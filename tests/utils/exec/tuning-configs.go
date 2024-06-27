package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type TuningConfigArgs struct {
	Cluster           *string `hcl:"cluster"`
	Name              *string `hcl:"name"`
	Count             *int    `hcl:"tc_count"`
	Spec              *string `hcl:"spec"`
	SpecVMDirtyRatios *[]int  `hcl:"spec_vm_dirty_ratios"`
	SpecPriorities    *[]int  `hcl:"spec_priorities"`
}
type TuningConfigOutput struct {
	Names []string `json:"names,omitempty"`
	Specs []string `json:"specs,omitempty"`
}

type TuningConfigService interface {
	Init() error
	Plan(args *TuningConfigArgs) (string, error)
	Apply(args *TuningConfigArgs) (string, error)
	Output() (*TuningConfigOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*TuningConfigArgs, error)
	DeleteTFVars() error
}

type tuningConfigService struct {
	tfExecutor TerraformExecutor
}

func NewTuningConfigService(manifestsDirs ...string) (TuningConfigService, error) {
	manifestsDir := constants.TuningConfigDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &tuningConfigService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *tuningConfigService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *tuningConfigService) Plan(args *TuningConfigArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *tuningConfigService) Apply(args *TuningConfigArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *tuningConfigService) Output() (*TuningConfigOutput, error) {
	var output TuningConfigOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *tuningConfigService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *tuningConfigService) ReadTFVars() (*TuningConfigArgs, error) {
	args := &TuningConfigArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *tuningConfigService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
