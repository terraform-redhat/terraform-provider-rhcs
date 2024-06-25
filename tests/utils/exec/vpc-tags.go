package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type VPCTagArgs struct {
	AWSRegion *string   `hcl:"aws_region"`
	IDs       *[]string `hcl:"ids"`
	TagKey    *string   `hcl:"key"`
	TagValue  *string   `hcl:"value"`
}

type VPCTagOutput struct {
}

type VPCTagService interface {
	Init() error
	Plan(args *VPCTagArgs) (string, error)
	Apply(args *VPCTagArgs) (string, error)
	Output() (*VPCTagOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*VPCTagArgs, error)
	DeleteTFVars() error
}

type vpcTagService struct {
	tfExecutor TerraformExecutor
}

func NewVPCTagService(tfWorkspace string, clusterType constants.ClusterType) (VPCTagService, error) {
	svc := &vpcTagService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetAWSVPCTagManifestDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *vpcTagService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *vpcTagService) Plan(args *VPCTagArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *vpcTagService) Apply(args *VPCTagArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *vpcTagService) Output() (*VPCTagOutput, error) {
	var output VPCTagOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *vpcTagService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *vpcTagService) ReadTFVars() (*VPCTagArgs, error) {
	args := &VPCTagArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *vpcTagService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
