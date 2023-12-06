package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type VPCTagArgs struct {
	AWSRegion string   `json:"aws_region,omitempty"`
	IDs       []string `json:"ids,omitempty"`
	TagKey    string   `json:"key,omitempty"`
	TagValue  string   `json:"value,omitempty"`
}

type VPCTagService struct {
	CreationArgs *VPCTagArgs
	ManifestDir  string
	Context      context.Context
}

func (vpctag *VPCTagService) Init(manifestDirs ...string) error {
	vpctag.ManifestDir = CON.AWSVPCTagDir
	if len(manifestDirs) != 0 {
		vpctag.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	vpctag.Context = ctx
	err := runTerraformInit(ctx, vpctag.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (vpctag *VPCTagService) Create(createArgs *VPCTagArgs, extraArgs ...string) error {
	vpctag.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(vpctag.Context, vpctag.ManifestDir, args)
	if err != nil {
		return err
	}
	return nil
}

func (vpctag *VPCTagService) Destroy(createArgs ...*VPCTagArgs) error {
	if vpctag.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := vpctag.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroyWithArgs(vpctag.Context, vpctag.ManifestDir, args)

	return err
}

func NewVPCTagService(manifestDir ...string) *VPCTagService {
	vpctag := &VPCTagService{}
	vpctag.Init(manifestDir...)
	return vpctag
}
