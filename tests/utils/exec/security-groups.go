package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type SecurityGroupArgs struct {
	NamePrefix  *string `hcl:"name_prefix"`
	SGNumber    *int    `hcl:"sg_number"`
	VPCID       *string `hcl:"vpc_id"`
	Description *string `hcl:"description"`
	AWSRegion   *string `hcl:"aws_region"`
}

type SecurityGroupsOutput struct {
	SGIDs []string `json:"sg_ids,omitempty"`
}

type SecurityGroupService interface {
	Init() error
	Plan(args *SecurityGroupArgs) (string, error)
	Apply(args *SecurityGroupArgs) (string, error)
	Output() (*SecurityGroupsOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*SecurityGroupArgs, error)
	DeleteTFVars() error
}

type securityGroupService struct {
	tfExecutor TerraformExecutor
}

func NewSecurityGroupService(tfWorkspace string, clusterType constants.ClusterType) (SecurityGroupService, error) {
	svc := &securityGroupService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetAWSSecurityGroupManifestDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *securityGroupService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *securityGroupService) Plan(args *SecurityGroupArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *securityGroupService) Apply(args *SecurityGroupArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *securityGroupService) Output() (*SecurityGroupsOutput, error) {
	var output SecurityGroupsOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *securityGroupService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *securityGroupService) ReadTFVars() (*SecurityGroupArgs, error) {
	args := &SecurityGroupArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *securityGroupService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
