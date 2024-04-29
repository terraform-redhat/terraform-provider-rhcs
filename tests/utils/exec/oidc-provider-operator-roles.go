package exec

import (
	"context"
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type OIDCProviderOperatorRolesArgs struct {
	AccountRolePrefix   string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix  string `json:"operator_role_prefix,omitempty"`
	OIDCConfig          string `json:"oidc_config,omitempty"`
	AWSRegion           string `json:"aws_region,omitempty"`
	OCMENV              string `json:"rhcs_environment,omitempty"`
	UnifiedAccRolesPath string `json:"path,omitempty"`
}

type OIDCProviderOperatorRolesOutput struct {
	OIDCConfigID           string `json:"oidc_config_id,omitempty"`
	AccountRolePrefix      string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix     string `json:"operator_role_prefix,omitempty"`
	IngressOperatorRoleArn string `json:"ingress_operator_role_arn,omitempty"`
}

type OIDCProviderOperatorRolesService struct {
	CreationArgs *OIDCProviderOperatorRolesArgs
	ManifestDir  string
	Context      context.Context
}

func (oidcOP *OIDCProviderOperatorRolesService) Init(manifestDirs ...string) error {
	oidcOP.ManifestDir = CON.OIDCProviderOperatorRolesClassicManifestDir
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

func (oidcOP *OIDCProviderOperatorRolesService) Apply(createArgs *OIDCProviderOperatorRolesArgs, recordtfvars bool, extraArgs ...string) (
	*OIDCProviderOperatorRolesOutput, error) {
	createArgs.OCMENV = constants.RHCS.OCMEnv
	oidcOP.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(oidcOP.Context, oidcOP.ManifestDir, args...)
	if err != nil {
		return nil, err
	}
	output, err := oidcOP.Output()
	if err != nil {
		return nil, err
	}
	if recordtfvars {
		recordTFvarsFile(oidcOP.ManifestDir, tfvars)
	}

	return output, nil
}

func (oidcOP *OIDCProviderOperatorRolesService) Output() (*OIDCProviderOperatorRolesOutput, error) {
	out, err := runTerraformOutput(oidcOP.Context, oidcOP.ManifestDir)
	if err != nil {
		return nil, err
	}
	var oidcOPOutput = &OIDCProviderOperatorRolesOutput{
		AccountRolePrefix:      h.DigString(out["account_roles_prefix"], "value"),
		OIDCConfigID:           h.DigString(out["oidc_config_id"], "value"),
		OperatorRolePrefix:     h.DigString(out["operator_role_prefix"], "value"),
		IngressOperatorRoleArn: h.DigString(out["ingress_operator_role_arn"], "value"),
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
	destroyArgs.OCMENV = constants.RHCS.OCMEnv
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(oidcOP.Context, oidcOP.ManifestDir, args...)
	return err
}

func NewOIDCProviderOperatorRolesService(manifestDir ...string) (*OIDCProviderOperatorRolesService, error) {
	oidcOP := &OIDCProviderOperatorRolesService{}
	err := oidcOP.Init(manifestDir...)
	return oidcOP, err
}
