package exec

import (
	"context"
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type AccountRolesArgs struct {
	AccountRolePrefix   string `json:"account_role_prefix,omitempty"`
	OCMENV              string `json:"rhcs_environment,omitempty"`
	OpenshiftVersion    string `json:"openshift_version,omitempty"`
	ChannelGroup        string `json:"channel_group,omitempty"`
	UnifiedAccRolesPath string `json:"path,omitempty"`
	SharedVpcRoleArn    string `json:"shared_vpc_role_arn,omitempty"`
}

type AccountRolesOutput struct {
	AccountRolePrefix string        `json:"account_role_prefix,omitempty"`
	MajorVersion      string        `json:"major_version,omitempty"`
	ChannelGroup      string        `json:"channel_group,omitempty"`
	RHCSVersions      []interface{} `json:"rhcs_versions,omitempty"`
	InstallerRoleArn  string        `json:"installer_role_arn,omitempty"`
	AWSAccountId      string        `json:"aws_account_id,omitempty"`
}

type AccountRoleService struct {
	CreationArgs *AccountRolesArgs
	ManifestDir  string
	Context      context.Context
}

func (acc *AccountRoleService) Init(manifestDirs ...string) error {
	acc.ManifestDir = constants.AccountRolesClassicDir
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

func (acc *AccountRoleService) Apply(createArgs *AccountRolesArgs, recordtfvars bool, extraArgs ...string) (*AccountRolesOutput, error) {
	createArgs.OCMENV = constants.RHCS.OCMEnv
	acc.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(acc.Context, acc.ManifestDir, args...)
	if err != nil {
		return nil, err
	}
	if recordtfvars {
		recordTFvarsFile(acc.ManifestDir, tfvars)
	}
	output, err := acc.Output()
	return output, err
}

func (acc *AccountRoleService) Output() (*AccountRolesOutput, error) {
	out, err := runTerraformOutput(acc.Context, acc.ManifestDir)
	if err != nil {
		return nil, err
	}
	var accOutput = &AccountRolesOutput{
		AccountRolePrefix: helper.DigString(out["account_role_prefix"], "value"),
		MajorVersion:      helper.DigString(out["major_version"], "value"),
		ChannelGroup:      helper.DigString(out["channel_group"], "value"),
		RHCSVersions:      helper.DigArray(out["rhcs_versions"], "value"),
		InstallerRoleArn:  helper.DigString(out["installer_role_arn"], "value"),
		AWSAccountId:      helper.DigString(out["aws_account_id"], "value"),
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
	destroyArgs.OCMENV = constants.RHCS.OCMEnv
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(acc.Context, acc.ManifestDir, args...)
	return err
}

func NewAccountRoleService(manifestDir ...string) (*AccountRoleService, error) {
	acc := &AccountRoleService{}
	err := acc.Init(manifestDir...)
	return acc, err
}
