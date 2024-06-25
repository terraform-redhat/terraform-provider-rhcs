package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type TuningConfigArgs struct {
	Cluster *string             `hcl:"cluster"`
	Name    *string             `hcl:"name"`
	Count   *int                `hcl:"tc_count"`
	Specs   *[]TuningConfigSpec `hcl:"specs"`
}

type TuningConfigSpec struct {
	Type  *string `cty:"spec_type"`
	Value *string `cty:"spec_value"`
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

func NewTuningConfigService(tfWorkspace string, clusterType constants.ClusterType) (TuningConfigService, error) {
	svc := &tuningConfigService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetTuningConfigManifestsDir(clusterType)),
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

func NewTuningConfigSpecFromString(specValue string) TuningConfigSpec {
	return TuningConfigSpec{
		Type:  helper.StringPointer("string"),
		Value: &specValue,
	}
}

func NewTuningConfigSpecFromFile(specFile string) TuningConfigSpec {
	return TuningConfigSpec{
		Type:  helper.StringPointer("file"),
		Value: &specFile,
	}
}
