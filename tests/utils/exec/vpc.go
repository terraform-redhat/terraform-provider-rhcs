package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type VPCArgs struct {
	Name                      string   `json:"name,omitempty"`
	AWSRegion                 string   `json:"aws_region,omitempty"`
	VPCCIDR                   string   `json:"vpc_cidr,omitempty"`
	MultiAZ                   bool     `json:"multi_az,omitempty"`
	AZIDs                     []string `json:"az_ids,omitempty"`
	HCP                       bool     `json:"hcp,omitempty"`
	AWSSharedCredentialsFiles []string `json:"aws_shared_credentials_files,omitempty"`
}

type VPCOutput struct {
	ClusterPublicSubnets  []string `json:"cluster-public-subnet,omitempty"`
	VPCCIDR               string   `json:"vpc-cidr,omitempty"`
	ClusterPrivateSubnets []string `json:"cluster-private-subnet,omitempty"`
	AZs                   []string `json:"azs,omitempty"`
	NodePrivateSubnets    []string `json:"node-private-subnet,omitempty"`
	VPCID                 string   `json:"vpc_id,omitempty"`
}

type VPCService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newVPCService(clusterType CON.ClusterType, tfExec TerraformExec) (*VPCService, error) {
	vpc := &VPCService{
		ManifestDir: GetVpcManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := vpc.Init()
	return vpc, err
}

func (vpc *VPCService) Init() error {
	return vpc.tfExec.RunTerraformInit(vpc.ManifestDir)
}

func (vpc *VPCService) Apply(createArgs *VPCArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = vpc.tfExec.RunTerraformApply(vpc.ManifestDir, tfVars)
	return
}

func (vpc *VPCService) Output() (*VPCOutput, error) {
	out, err := vpc.tfExec.RunTerraformOutput(vpc.ManifestDir)
	if err != nil {
		return nil, err
	}
	vpcOutput := &VPCOutput{
		VPCCIDR:               h.DigString(out["vpc-cidr"], "value"),
		ClusterPrivateSubnets: h.DigArrayToString(out["cluster-private-subnet"], "value"),
		ClusterPublicSubnets:  h.DigArrayToString(out["cluster-public-subnet"], "value"),
		NodePrivateSubnets:    h.DigArrayToString(out["node-private-subnet"], "value"),
		AZs:                   h.DigArrayToString(out["azs"], "value"),
		VPCID:                 h.DigString(out["vpc-id"], "value"),
	}

	return vpcOutput, err
}

func (vpc *VPCService) Destroy(deleteTFVars bool) (string, error) {
	return vpc.tfExec.RunTerraformDestroy(vpc.ManifestDir, deleteTFVars)
}
