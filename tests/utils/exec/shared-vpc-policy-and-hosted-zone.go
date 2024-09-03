package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type SharedVpcPolicyAndHostedZoneArgs struct {
	SharedVpcAWSSharedCredentialsFiles *[]string `hcl:"shared_vpc_aws_shared_credentials_files"`
	Region                             *string   `hcl:"region"`
	ClusterName                        *string   `hcl:"cluster_name"`
	DomainPrefix                       *string   `hcl:"domain_prefix"`
	DnsDomainId                        *string   `hcl:"dns_domain_id"`
	IngressOperatorRoleArn             *string   `hcl:"ingress_operator_role_arn"`
	InstallerRoleArn                   *string   `hcl:"installer_role_arn"`
	ClusterAWSAccount                  *string   `hcl:"cluster_aws_account"`
	VpcId                              *string   `hcl:"vpc_id"`
	Subnets                            *[]string `hcl:"subnets"`
}

type SharedVpcPolicyAndHostedZoneOutput struct {
	SharedRole    string   `json:"shared_role,omitempty"`
	HostedZoneId  string   `json:"hosted_zone_id,omitempty"`
	SharedSubnets []string `json:"shared_subnets,omitempty"`
}

type SharedVpcPolicyAndHostedZoneService interface {
	Init() error
	Plan(args *SharedVpcPolicyAndHostedZoneArgs) (string, error)
	Apply(args *SharedVpcPolicyAndHostedZoneArgs) (string, error)
	Output() (*SharedVpcPolicyAndHostedZoneOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*SharedVpcPolicyAndHostedZoneArgs, error)
	DeleteTFVars() error
}

type sharedVpcPolicyAndHostedZoneService struct {
	tfExecutor TerraformExecutor
}

func NewSharedVpcPolicyAndHostedZoneService(manifestsDirs ...string) (SharedVpcPolicyAndHostedZoneService, error) {
	manifestsDir := constants.SharedVpcPolicyAndHostedZoneDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &sharedVpcPolicyAndHostedZoneService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *sharedVpcPolicyAndHostedZoneService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *sharedVpcPolicyAndHostedZoneService) Plan(args *SharedVpcPolicyAndHostedZoneArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *sharedVpcPolicyAndHostedZoneService) Apply(args *SharedVpcPolicyAndHostedZoneArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *sharedVpcPolicyAndHostedZoneService) Output() (*SharedVpcPolicyAndHostedZoneOutput, error) {
	var output SharedVpcPolicyAndHostedZoneOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *sharedVpcPolicyAndHostedZoneService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *sharedVpcPolicyAndHostedZoneService) ReadTFVars() (*SharedVpcPolicyAndHostedZoneArgs, error) {
	args := &SharedVpcPolicyAndHostedZoneArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *sharedVpcPolicyAndHostedZoneService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
