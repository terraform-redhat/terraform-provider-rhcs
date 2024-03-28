package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type SharedVpcPolicyAndHostedZoneArgs struct {
	SharedVpcAWSSharedCredentialsFiles []string `json:"shared_vpc_aws_shared_credentials_files,omitempty"`
	Region                             string   `json:"region,omitempty"`
	ClusterName                        string   `json:"cluster_name,omitempty"`
	DnsDomainId                        string   `json:"dns_domain_id,omitempty"`
	IngressOperatorRoleArn             string   `json:"ingress_operator_role_arn,omitempty"`
	InstallerRoleArn                   string   `json:"installer_role_arn,omitempty"`
	ClusterAWSAccount                  string   `json:"cluster_aws_account,omitempty"`
	VpcId                              string   `json:"vpc_id,omitempty"`
	Subnets                            []string `json:"subnets,omitempty"`
}

type SharedVpcPolicyAndHostedZoneOutput struct {
	SharedRole   string   `json:"shared_role,omitempty"`
	HostedZoneId string   `json:"hosted_zone_id,omitempty"`
	AZs          []string `json:"azs,omitempty"`
}

type SharedVpcPolicyAndHostedZoneService struct {
	CreationArgs *SharedVpcPolicyAndHostedZoneArgs
	ManifestDir  string
	Context      context.Context
}

func (s *SharedVpcPolicyAndHostedZoneService) Init(manifestDirs ...string) error {
	s.ManifestDir = CON.SharedVpcPolicyAndHostedZoneDir
	if len(manifestDirs) != 0 {
		s.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	s.Context = ctx
	err := runTerraformInit(ctx, s.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (s *SharedVpcPolicyAndHostedZoneService) Apply(createArgs *SharedVpcPolicyAndHostedZoneArgs, recordtfvars bool, extraArgs ...string) error {
	s.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(s.Context, s.ManifestDir, args...)
	if err != nil {
		return err
	}
	if recordtfvars {
		recordTFvarsFile(s.ManifestDir, tfvars)
	}

	return nil
}

func (s *SharedVpcPolicyAndHostedZoneService) Output() (SharedVpcPolicyAndHostedZoneOutput, error) {
	d := CON.SharedVpcPolicyAndHostedZoneDir
	if s.ManifestDir != "" {
		d = s.ManifestDir
	}
	var o SharedVpcPolicyAndHostedZoneOutput
	out, err := runTerraformOutput(context.TODO(), d)
	if err != nil {
		return o, err
	}
	o = SharedVpcPolicyAndHostedZoneOutput{
		SharedRole:   h.DigString(out["shared_role"], "value"),
		HostedZoneId: h.DigString(out["hosted_zone_id"], "value"),
		AZs:          h.DigArrayToString(out["azs"], "value"),
	}

	return o, nil
}

func (s *SharedVpcPolicyAndHostedZoneService) Destroy(createArgs ...*SharedVpcPolicyAndHostedZoneArgs) error {
	if s.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := s.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(s.Context, s.ManifestDir, args...)

	return err
}

func NewSharedVpcPolicyAndHostedZoneService(manifestDir ...string) (*SharedVpcPolicyAndHostedZoneService, error) {
	s := &SharedVpcPolicyAndHostedZoneService{}
	err := s.Init(manifestDir...)
	return s, err
}
