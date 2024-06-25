package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
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

func NewRhcsInfoService(tfWorkspace string, clusterType constants.ClusterType) (RhcsInfoService, error) {
	svc := &rhcsInfoService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetRHCSInfoManifestsDir(clusterType)),
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
