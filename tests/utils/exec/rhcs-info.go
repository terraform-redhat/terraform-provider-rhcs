package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type RhcsInfoArgs struct {
}

type RhcsInfoOutput struct{}

type RhcsInfoService interface {
	Init() error
	Apply(args *RhcsInfoArgs) (string, error)
	ShowState(resource string) (string, error)
}

type rhcsInfoService struct {
	tfExecutor TerraformExecutor
}

func NewRhcsInfoService(manifestsDirs ...string) (RhcsInfoService, error) {
	manifestsDir := constants.RhcsInfoDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &rhcsInfoService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *rhcsInfoService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *rhcsInfoService) Apply(args *RhcsInfoArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *rhcsInfoService) ShowState(resource string) (string, error) {
	return svc.tfExecutor.RunTerraformState("show", resource)
}
