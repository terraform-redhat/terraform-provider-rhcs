package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type AccountRolesArgs struct {
	AccountRolePrefix   string `json:"account_role_prefix,omitempty"`
	OCMENV              string `json:"rhcs_environment,omitempty"`
	OpenshiftVersion    string `json:"openshift_version,omitempty"`
	URL                 string `json:"url,omitempty"`
	ChannelGroup        string `json:"channel_group,omitempty"`
	UnifiedAccRolesPath string `json:"path,omitempty"`
	SharedVpcRoleArn    string `json:"shared_vpc_role_arn,omitempty"`
}

func (args *AccountRolesArgs) appendURLAndENV() {
	args.URL = CON.GateWayURL
	args.OCMENV = CON.OCMENV
}

type AccountRolesOutput struct {
	AccountRolePrefix string        `json:"account_role_prefix,omitempty"`
	MajorVersion      string        `json:"major_version,omitempty"`
	ChannelGroup      string        `json:"channel_group,omitempty"`
	RHCSGatewayUrl    string        `json:"rhcs_gateway_url,omitempty"`
	RHCSVersions      []interface{} `json:"rhcs_versions,omitempty"`
	InstallerRoleArn  string        `json:"installer_role_arn,omitempty"`
	AWSAccountId      string        `json:"aws_account_id,omitempty"`
}

type AccountRoleService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newAccountRoleService(clusterType CON.ClusterType, tfExec TerraformExec) (*AccountRoleService, error) {
	ars := &AccountRoleService{
		ManifestDir: GetAccountRoleManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := ars.Init()
	return ars, err
}

func (ars *AccountRoleService) Init() error {
	return ars.tfExec.RunTerraformInit(ars.ManifestDir)
}

func (ars *AccountRoleService) Apply(createArgs *AccountRolesArgs, recordtfvars bool) (output string, err error) {
	createArgs.appendURLAndENV()

	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = ars.tfExec.RunTerraformApply(ars.ManifestDir, tfVars)
	return
}

func (ars *AccountRoleService) Output() (*AccountRolesOutput, error) {
	out, err := ars.tfExec.RunTerraformOutput(ars.ManifestDir)
	if err != nil {
		return nil, err
	}
	var accOutput = &AccountRolesOutput{
		AccountRolePrefix: h.DigString(out["account_role_prefix"], "value"),
		MajorVersion:      h.DigString(out["major_version"], "value"),
		ChannelGroup:      h.DigString(out["channel_group"], "value"),
		RHCSGatewayUrl:    h.DigString(out["rhcs_gateway_url"], "value"),
		RHCSVersions:      h.DigArray(out["rhcs_versions"], "value"),
		InstallerRoleArn:  h.DigString(out["installer_role_arn"], "value"),
		AWSAccountId:      h.DigString(out["aws_account_id"], "value"),
	}
	return accOutput, nil
}

func (ars *AccountRoleService) Destroy(deleteTFVars bool) (output string, err error) {
	return ars.tfExec.RunTerraformDestroy(ars.ManifestDir, deleteTFVars)
}
