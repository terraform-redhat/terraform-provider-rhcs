package exec

import (
	"context"
	"fmt"

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
	CreationArgs *SecurityGroupArgs
	ManifestDir  string
	Context      context.Context
}

func (sgs *SecurityGroupService) Init(manifestDirs ...string) error {
	sgs.ManifestDir = CON.AWSSecurityGroupDir
	if len(manifestDirs) != 0 {
		sgs.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	sgs.Context = ctx
	err := runTerraformInit(ctx, sgs.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (sgs *SecurityGroupService) Apply(createArgs *SecurityGroupArgs, recordtfvars bool, extraArgs ...string) error {
	sgs.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(sgs.Context, sgs.ManifestDir, args...)
	if err != nil {
		return err
	}
	if recordtfvars {
		recordTFvarsFile(sgs.ManifestDir, tfvars)
	}

	return nil
}

func (sgs *SecurityGroupService) Output() (*SecurityGroupsOutput, error) {
	sgsDir := CON.AWSSecurityGroupDir
	if sgs.ManifestDir != "" {
		sgsDir = sgs.ManifestDir
	}
	out, err := runTerraformOutput(context.TODO(), sgsDir)
	sgOut := &SecurityGroupsOutput{
		SGIDs: h.DigArrayToString(out["sg_ids"], "value"),
	}

	return sgOut, err
}

func (sgs *SecurityGroupService) Destroy(createArgs ...*SecurityGroupArgs) error {
	if sgs.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := sgs.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(sgs.Context, sgs.ManifestDir, args...)

	return err
}

func NewSecurityGroupService(manifestDir ...string) *SecurityGroupService {
	sgs := &SecurityGroupService{}
	sgs.Init(manifestDir...)
	return sgs
}
