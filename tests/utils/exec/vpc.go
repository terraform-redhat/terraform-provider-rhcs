package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type VPCArgs struct {
	AWSRegion                 *string            `hcl:"aws_region"`
	AWSSharedCredentialsFiles *[]string          `hcl:"aws_shared_credentials_files"`
	NamePrefix                *string            `hcl:"name_prefix"`
	VPCCIDR                   *string            `hcl:"vpc_cidr"`
	AvailabilityZones         *[]string          `hcl:"availability_zones"`
	AvailabilityZonesCount    *int               `hcl:"availability_zones_count"`
	Tags                      *map[string]string `hcl:"tags"`
}

type VPCOutput struct {
	PrivateSubnets    []string `json:"private_subnets,omitempty"`
	PublicSubnets     []string `json:"public_subnets,omitempty"`
	AvailabilityZones []string `json:"availability_zones,omitempty"`
	VPCID             string   `json:"vpc_id,omitempty"`
	VPCCIDR           string   `json:"vpc_cidr,omitempty"`
}
type VPCService interface {
	Init() error
	Plan(args *VPCArgs) (string, error)
	Apply(args *VPCArgs) (string, error)
	Output() (*VPCOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*VPCArgs, error)
	DeleteTFVars() error
}

type vpcService struct {
	tfExecutor TerraformExecutor
}

func NewVPCService(tfWorkspace string, clusterType constants.ClusterType) (VPCService, error) {
	svc := &vpcService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetAWSVPCManifestDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *vpcService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *vpcService) Plan(args *VPCArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *vpcService) Apply(args *VPCArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *vpcService) Output() (*VPCOutput, error) {
	var output VPCOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *vpcService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *vpcService) ReadTFVars() (*VPCArgs, error) {
	args := &VPCArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *vpcService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
