package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type OIDCProviderOperatorRolesArgs struct {
	AccountRolePrefix   string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix  string `json:"operator_role_prefix,omitempty"`
	URL                 string `json:"url,omitempty"`
	OIDCConfig          string `json:"oidc_config,omitempty"`
	AWSRegion           string `json:"aws_region,omitempty"`
	OCMENV              string `json:"rhcs_environment,omitempty"`
	UnifiedAccRolesPath string `json:"path,omitempty"`
}

func (args *OIDCProviderOperatorRolesArgs) appendURL() {
	args.URL = CON.GateWayURL
	args.OCMENV = CON.OCMENV
}

type OIDCProviderOperatorRolesOutput struct {
	OIDCConfigID           string `json:"oidc_config_id,omitempty"`
	AccountRolePrefix      string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix     string `json:"operator_role_prefix,omitempty"`
	IngressOperatorRoleArn string `json:"ingress_operator_role_arn,omitempty"`
}

type OIDCProviderOperatorRolesService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newOIDCProviderOperatorRoleService(clusterType CON.ClusterType, tfExec TerraformExec) (*OIDCProviderOperatorRolesService, error) {
	oidcOP := &OIDCProviderOperatorRolesService{
		ManifestDir: GetOIDCProviderOperatorRolesManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := oidcOP.Init()
	return oidcOP, err
}

func (oidcOP *OIDCProviderOperatorRolesService) Init() error {
	return oidcOP.tfExec.RunTerraformInit(oidcOP.ManifestDir)
}

func (oidcOP *OIDCProviderOperatorRolesService) Apply(createArgs *OIDCProviderOperatorRolesArgs) (
	output string, err error) {
	createArgs.appendURL()

	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = oidcOP.tfExec.RunTerraformApply(oidcOP.ManifestDir, tfVars)
	return
}

func (oidcOP *OIDCProviderOperatorRolesService) Output() (*OIDCProviderOperatorRolesOutput, error) {
	out, err := oidcOP.tfExec.RunTerraformOutput(oidcOP.ManifestDir)
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

func (oidcOP *OIDCProviderOperatorRolesService) Destroy(deleteTFVars bool) (string, error) {
	return oidcOP.tfExec.RunTerraformDestroy(oidcOP.ManifestDir, deleteTFVars)
}
