package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type SecurityGroupArgs struct {
	NamePrefix  string `json:"name_prefix,omitempty"`
	SGNumber    int    `json:"sg_number,omitempty"`
	VPCID       string `json:"vpc_id,omitempty"`
	Description string `json:"description,omitempty"`
	AWSRegion   string `json:"aws_region,omitempty"`
}

type SecurityGroupsOutput struct {
	SGIDs []string `json:"sg_ids,omitempty"`
}

type SecurityGroupService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newSecurityGroupService(clusterType CON.ClusterType, tfExec TerraformExec) (*SecurityGroupService, error) {
	sgs := &SecurityGroupService{
		ManifestDir: GetSecurityGroupManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := sgs.Init()
	return sgs, err
}

func (sgs *SecurityGroupService) Init() error {
	return sgs.tfExec.RunTerraformInit(sgs.ManifestDir)
}

func (sgs *SecurityGroupService) Apply(createArgs *SecurityGroupArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = sgs.tfExec.RunTerraformApply(sgs.ManifestDir, tfVars)
	return
}

func (sgs *SecurityGroupService) Output() (*SecurityGroupsOutput, error) {
	out, err := sgs.tfExec.RunTerraformOutput(sgs.ManifestDir)
	if err != nil {
		return nil, err
	}
	sgOut := &SecurityGroupsOutput{
		SGIDs: h.DigArrayToString(out["sg_ids"], "value"),
	}

	return sgOut, err
}

func (sgs *SecurityGroupService) Destroy(deleteTFVars bool) (string, error) {
	return sgs.tfExec.RunTerraformDestroy(sgs.ManifestDir, deleteTFVars)
}
