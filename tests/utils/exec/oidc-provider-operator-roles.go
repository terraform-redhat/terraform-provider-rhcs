package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type OIDCProviderOperatorRolesArgs struct {
	AccountRolePrefix  string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix string `json:"operator_role_prefix,omitempty"`
	Token              string `json:"token,omitempty"`
	URL                string `json:"url,omitempty"`
	OIDCConfig         string `json:"oidc_config,omitempty"`
	AWSRegion          string `json:"aws_region,omitempty"`
}

type OIDCProviderOperatorRolesOutput struct {
	OIDCConfigID       string `json:"oidc_config_id,omitempty"`
	AccountRolePrefix  string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix string `json:"operator_role_prefix,omitempty"`
}

type OIDCProviderOperatorRolesService struct {
	CreationArgs *OIDCProviderOperatorRolesArgs
	ManifestDir  string
	Context      context.Context
}

func (oidcOP *OIDCProviderOperatorRolesService) Init(manifestDirs ...string) error {
	oidcOP.ManifestDir = CON.OIDCProviderOperatorRolesManifestDir
	if len(manifestDirs) != 0 {
		oidcOP.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	oidcOP.Context = ctx
	err := runTerraformInit(ctx, oidcOP.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (oidcOP *OIDCProviderOperatorRolesService) Create(createArgs *OIDCProviderOperatorRolesArgs, extraArgs ...string) (
	*OIDCProviderOperatorRolesOutput, error) {
	oidcOP.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(oidcOP.Context, oidcOP.ManifestDir, args)
	if err != nil {
		return nil, err
	}
	output, err := oidcOP.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (oidcOP *OIDCProviderOperatorRolesService) Output() (*OIDCProviderOperatorRolesOutput, error) {
	out, err := runTerraformOutput(oidcOP.Context, oidcOP.ManifestDir)
	if err != nil {
		return nil, err
	}
	var oidcOPOutput = &OIDCProviderOperatorRolesOutput{
		AccountRolePrefix:  h.DigString(out["account_roles_prefix"], "value"),
		OIDCConfigID:       h.DigString(out["oidc_config_id"], "value"),
		OperatorRolePrefix: h.DigString(out["operator_role_prefix"], "value"),
	}
	return oidcOPOutput, nil
}

func (oidcOP *OIDCProviderOperatorRolesService) Destroy(createArgs ...*OIDCProviderOperatorRolesArgs) error {
	if oidcOP.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := oidcOP.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs)
	err := runTerraformDestroyWithArgs(oidcOP.Context, oidcOP.ManifestDir, args)
	return err
}

func NewOIDCProviderOperatorRolesService(manifestDir ...string) (*OIDCProviderOperatorRolesService, error) {
	oidcOP := &OIDCProviderOperatorRolesService{}
	err := oidcOP.Init(manifestDir...)
	return oidcOP, err
}
