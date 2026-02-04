package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type OIDCProviderOperatorRolesArgs struct {
	AccountRolePrefix   *string            `hcl:"account_role_prefix"`
	OperatorRolePrefix  *string            `hcl:"operator_role_prefix"`
	OIDCConfig          *string            `hcl:"oidc_config"`
	OIDCPrefix          *string            `hcl:"oidc_prefix"`
	UnifiedAccRolesPath *string            `hcl:"path"`
	Tags                *map[string]string `hcl:"tags"`
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

func NewOIDCProviderOperatorRolesService(tfWorkspace string, clusterType constants.ClusterType) (OIDCProviderOperatorRolesService, error) {
	svc := &oidcProviderOperatorRolesService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetAWSOIDCProviderOperatorRolesManifestDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *oidcProviderOperatorRolesService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *oidcProviderOperatorRolesService) Plan(args *OIDCProviderOperatorRolesArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *oidcProviderOperatorRolesService) Apply(args *OIDCProviderOperatorRolesArgs) (string, error) {
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
