package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type OIDCProviderOperatorRolesArgs struct {
	AccountRolePrefix   *string `hcl:"account_role_prefix"`
	OperatorRolePrefix  *string `hcl:"operator_role_prefix"`
	OIDCConfig          *string `hcl:"oidc_config"`
	AWSRegion           *string `hcl:"aws_region"`
	OCMENV              *string `hcl:"rhcs_environment"`
	UnifiedAccRolesPath *string `hcl:"path"`
}

type OIDCProviderOperatorRolesOutput struct {
	OIDCConfigID           string `json:"oidc_config_id,omitempty"`
	AccountRolePrefix      string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix     string `json:"operator_role_prefix,omitempty"`
	IngressOperatorRoleArn string `json:"ingress_operator_role_arn,omitempty"`
}

type OIDCProviderOperatorRolesService interface {
	Init() error
	Plan(args *OIDCProviderOperatorRolesArgs) (string, error)
	Apply(args *OIDCProviderOperatorRolesArgs) (string, error)
	Output() (*OIDCProviderOperatorRolesOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*OIDCProviderOperatorRolesArgs, error)
	DeleteTFVars() error
}

type oidcProviderOperatorRolesService struct {
	tfExecutor TerraformExecutor
}

func NewOIDCProviderOperatorRolesService(manifestsDirs ...string) (OIDCProviderOperatorRolesService, error) {
	manifestsDir := constants.OIDCProviderOperatorRolesClassicManifestDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &oidcProviderOperatorRolesService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *oidcProviderOperatorRolesService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *oidcProviderOperatorRolesService) Plan(args *OIDCProviderOperatorRolesArgs) (string, error) {
	args.OCMENV = &constants.RHCS.OCMEnv
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *oidcProviderOperatorRolesService) Apply(args *OIDCProviderOperatorRolesArgs) (string, error) {
	args.OCMENV = &constants.RHCS.OCMEnv
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *oidcProviderOperatorRolesService) Output() (*OIDCProviderOperatorRolesOutput, error) {
	var output OIDCProviderOperatorRolesOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *oidcProviderOperatorRolesService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *oidcProviderOperatorRolesService) ReadTFVars() (*OIDCProviderOperatorRolesArgs, error) {
	args := &OIDCProviderOperatorRolesArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *oidcProviderOperatorRolesService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
