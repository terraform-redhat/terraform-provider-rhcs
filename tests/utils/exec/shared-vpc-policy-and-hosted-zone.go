package exec

import (
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
	tfExec      TerraformExec
	ManifestDir string
}

func newSharedVPCPolicyAndHostedZoneService(clusterType CON.ClusterType, tfExec TerraformExec) (*SharedVpcPolicyAndHostedZoneService, error) {
	sharedVPCService := &SharedVpcPolicyAndHostedZoneService{
		ManifestDir: GetSharedVpcPolicyAndHostedZoneDir(clusterType),
		tfExec:      tfExec,
	}
	err := sharedVPCService.Init()
	return sharedVPCService, err
}

func (svs *SharedVpcPolicyAndHostedZoneService) Init() error {
	return svs.tfExec.RunTerraformInit(svs.ManifestDir)

}

func (svs *SharedVpcPolicyAndHostedZoneService) Apply(createArgs *SharedVpcPolicyAndHostedZoneArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = svs.tfExec.RunTerraformApply(svs.ManifestDir, tfVars)
	return
}

func (svs *SharedVpcPolicyAndHostedZoneService) Output() (*SharedVpcPolicyAndHostedZoneOutput, error) {
	out, err := svs.tfExec.RunTerraformOutput(svs.ManifestDir)
	if err != nil {
		return nil, err
	}
	svsOut := &SharedVpcPolicyAndHostedZoneOutput{
		SharedRole:   h.DigString(out["shared_role"], "value"),
		HostedZoneId: h.DigString(out["hosted_zone_id"], "value"),
		AZs:          h.DigArrayToString(out["azs"], "value"),
	}

	return svsOut, err
}

func (svs *SharedVpcPolicyAndHostedZoneService) Destroy(deleteTFVars bool) (string, error) {
	return svs.tfExec.RunTerraformDestroy(svs.ManifestDir, deleteTFVars)
}
