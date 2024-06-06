package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type AccountRolesArgs struct {
	AccountRolePrefix   *string `hcl:"account_role_prefix"`
	OCMENV              *string `hcl:"rhcs_environment"`
	OpenshiftVersion    *string `hcl:"openshift_version"`
	ChannelGroup        *string `hcl:"channel_group"`
	UnifiedAccRolesPath *string `hcl:"path"`
	SharedVpcRoleArn    *string `hcl:"shared_vpc_role_arn"`
}

type AccountRolesOutput struct {
	RHCSVersions      []interface{} `json:"rhcs_versions,omitempty"`
	AccountRolePrefix string        `json:"account_role_prefix,omitempty"`
	Path              string        `json:"path,omitempty"`
	MajorVersion      string        `json:"major_version,omitempty"`
	ChannelGroup      string        `json:"channel_group,omitempty"`
	InstallerRoleArn  string        `json:"installer_role_arn,omitempty"`
	AWSAccountId      string        `json:"aws_account_id,omitempty"`
}

type AccountRoleService interface {
	Init() error
	Plan(args *AccountRolesArgs) (string, error)
	Apply(args *AccountRolesArgs) (string, error)
	Output() (*AccountRolesOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*AccountRolesArgs, error)
	DeleteTFVars() error
}

type svcountRoleService struct {
	tfExecutor TerraformExecutor
}

func NewAccountRoleService(manifestsDirs ...string) (AccountRoleService, error) {
	manifestsDir := constants.AccountRolesClassicDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &svcountRoleService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *svcountRoleService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *svcountRoleService) Plan(args *AccountRolesArgs) (string, error) {
	args.OCMENV = &constants.RHCS.OCMEnv
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *svcountRoleService) Apply(args *AccountRolesArgs) (string, error) {
	args.OCMENV = &constants.RHCS.OCMEnv
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *svcountRoleService) Output() (*AccountRolesOutput, error) {
	var output AccountRolesOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *svcountRoleService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *svcountRoleService) ReadTFVars() (*AccountRolesArgs, error) {
	args := &AccountRolesArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *svcountRoleService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
