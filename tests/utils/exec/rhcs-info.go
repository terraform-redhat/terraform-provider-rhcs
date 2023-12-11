package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type RhcsInfoArgs struct {
	ID           string `json:"id,omitempty"`
	Token        string `json:"token,omitempty"`
	ResourceName string `json:"resource_name,omitempty"`
	ResourceKind string `json:"resource_kind,omitempty"`
}

type RhcsInfoService struct {
	CreationArgs *RhcsInfoArgs
	ManifestDir  string
	Context      context.Context
}

func (rhcsInfo *RhcsInfoService) Init(manifestDirs ...string) error {
	rhcsInfo.ManifestDir = CON.DNSDir
	if len(manifestDirs) != 0 {
		rhcsInfo.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	rhcsInfo.Context = ctx
	err := runTerraformInit(ctx, rhcsInfo.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (rhcsInfo *RhcsInfoService) Create(createArgs *RhcsInfoArgs, extraArgs ...string) error {
	rhcsInfo.CreationArgs = createArgs
	args, _ := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(rhcsInfo.Context, rhcsInfo.ManifestDir, args)
	if err != nil {
		return err
	}
	return nil
}

func (rhcsInfo *RhcsInfoService) Destroy(createArgs ...*RhcsInfoArgs) error {
	if rhcsInfo.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := rhcsInfo.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroyWithArgs(rhcsInfo.Context, rhcsInfo.ManifestDir, args)

	return err
}

func (rhcsInfoService *RhcsInfoService) ShowState(rhcsInfoArgs *RhcsInfoArgs) (string, error) {
	args := fmt.Sprintf("data.%s.%s", rhcsInfoArgs.ResourceKind, rhcsInfoArgs.ResourceName)
	output, err := runTerraformState(rhcsInfoService.ManifestDir, "show", args)
	return output, err
}

func NewRhcsInfoService(manifestDir ...string) *RhcsInfoService {
	rhcsInfo := &RhcsInfoService{}
	rhcsInfo.Init(manifestDir...)
	return rhcsInfo
}
