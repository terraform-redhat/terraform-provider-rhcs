package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type OCMRoleLinkArgs struct {
	RoleArn *string `hcl:"role_arn"`
}

type OCMRoleLink struct {
	RoleArn        string `json:"role_arn,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
}

type OCMRoleLinkService interface {
	Init() error
	Plan(args *OCMRoleLinkArgs) (string, error)
	Apply(args *OCMRoleLinkArgs) (string, error)
	Output() (*OCMRoleLink, error)
	Destroy() (string, error)

	ReadTFVars() (*OCMRoleLinkArgs, error)
	WriteTFVars(args *OCMRoleLinkArgs) error
	DeleteTFVars() error
}

type linkOCMRoleService struct {
	tfExecutor TerraformExecutor
}

func NewOCMRoleLinkService(tfWorkspace string) (OCMRoleLinkService, error) {
	svc := &linkOCMRoleService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetOCMRoleLinkManifestsDir()),
	}
	err := svc.Init()
	return svc, err
}

func (svc *linkOCMRoleService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *linkOCMRoleService) Plan(args *OCMRoleLinkArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *linkOCMRoleService) Apply(args *OCMRoleLinkArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *linkOCMRoleService) Output() (*OCMRoleLink, error) {
	output := &OCMRoleLink{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (svc *linkOCMRoleService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *linkOCMRoleService) ReadTFVars() (*OCMRoleLinkArgs, error) {
	args := &OCMRoleLinkArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *linkOCMRoleService) WriteTFVars(args *OCMRoleLinkArgs) error {
	err := svc.tfExecutor.WriteTerraformVars(args)
	return err
}

func (svc *linkOCMRoleService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
