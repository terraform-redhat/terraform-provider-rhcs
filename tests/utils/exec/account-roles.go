package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type AccountRolesArgs struct {
	AccountRolePrefix string `json:"account_role_prefix,omitempty"`
	OCMENV            string `json:"rhcs_environment,omitempty"`
	OpenshiftVersion  string `json:"openshift_version,omitempty"`
	Token             string `json:"token,omitempty"`
	URL               string `json:"url,omitempty"`
	ChannelGroup      string `json:"channel_group,omitempty"`
}

type AccountRolesOutput struct {
	AccountRolePrefix string        `json:"account_role_prefix,omitempty"`
	MajorVersion      string        `json:"major_version,omitempty"`
	ChannelGroup      string        `json:"channel_group,omitempty"`
	RHCSGatewayUrl    string        `json:"rhcs_gateway_url,omitempty"`
	RHCSVersions      []interface{} `json:"rhcs_versions,omitempty"`
}

type AccountRoleService struct {
	CreationArgs *AccountRolesArgs
	ManifestDir  string
	Context      context.Context
}

func (acc *AccountRoleService) Init(manifestDirs ...string) error {
	acc.ManifestDir = CON.AccountRolesDir
	if len(manifestDirs) != 0 {
		acc.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	acc.Context = ctx
	err := runTerraformInit(ctx, acc.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (acc *AccountRoleService) Create(createArgs *AccountRolesArgs, extraArgs ...string) (*AccountRolesOutput, error) {
	acc.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(acc.Context, acc.ManifestDir, args)
	if err != nil {
		return nil, err
	}
	output, err := acc.Output()
	return output, nil
}

func (acc *AccountRoleService) Output() (*AccountRolesOutput, error) {
	out, err := runTerraformOutput(acc.Context, acc.ManifestDir)
	if err != nil {
		return nil, err
	}
	var accOutput = &AccountRolesOutput{
		AccountRolePrefix: h.DigString(out["account_roles_prefix"], "value"),
		MajorVersion:      h.DigString(out["major_version"], "value"),
		ChannelGroup:      h.DigString(out["channel_group"], "value"),
		RHCSGatewayUrl:    h.DigString(out["rhcs_gateway_url"], "value"),
		RHCSVersions:      h.DigArray(out["rhcs_versions"], "value"),
	}
	return accOutput, nil
}

func (acc *AccountRoleService) Destroy(createArgs ...*AccountRolesArgs) error {
	if acc.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := acc.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs)
	err := runTerraformDestroyWithArgs(acc.Context, acc.ManifestDir, args)
	return err
}

func NewAccountRoleService(manifestDir ...string) (*AccountRoleService, error) {
	acc := &AccountRoleService{}
	err := acc.Init(manifestDir...)
	return acc, err
}
