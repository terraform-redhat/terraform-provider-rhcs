package exec

import (
	"context"
	"encoding/json"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type VPCArgs struct {
	Name      string   `json:"name,omitempty"`
	AWSRegion string   `json:"aws_region,omitempty"`
	VPCCIDR   string   `json:"vpc_cidr,omitempty"`
	MultiAZ   bool     `json:"multi_az,omitempty"`
	AZIDs     []string `json:"az_ids,omitempty"`
}

type VPCOutput struct {
	ClusterPublicSubnets  []string `json:"cluster-public-subnet,omitempty"`
	VPCCIDR               string   `json:"vpc-cidr,omitempty"`
	ClusterPrivateSubnets []string `json:"cluster-private-subnet,omitempty"`
	AZs                   []string `json:"azs,omitempty"`
	NodePrivateSubnets    []string `json:"node-private-subnet,omitempty"`
}

type VPCService struct {
	CreationArgs *VPCArgs
	ManifestDir  string
	Context      context.Context
}

func (vpc *VPCService) Init(manifestDirs ...string) error {
	vpc.ManifestDir = CON.AWSVPCDir
	if len(manifestDirs) != 0 {
		vpc.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	vpc.Context = ctx
	err := runTerraformInit(ctx, vpc.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (vpc *VPCService) Create(createArgs *VPCArgs, extraArgs ...string) error {
	vpc.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(vpc.Context, vpc.ManifestDir, args)
	if err != nil {
		return err
	}
	return nil
}

func (vpc *VPCService) Output() (*VPCOutput, error) {
	vpcDir := CON.AWSVPCDir
	if vpc.ManifestDir != "" {
		vpcDir = vpc.ManifestDir
	}
	out, err := runTerraformOutput(context.TODO(), vpcDir)
	vpcOutput := &VPCOutput{
		VPCCIDR:               h.DigString(out["vpc-cidr"], "value"),
		ClusterPrivateSubnets: h.DigArrayToString(out["cluster-private-subnet"], "value"),
		ClusterPublicSubnets:  h.DigArrayToString(out["cluster-public-subnet"], "value"),
		NodePrivateSubnets:    h.DigArrayToString(out["node-private-subnet"], "value"),
		AZs:                   h.DigArrayToString(out["azs"], "value"),
	}

	return vpcOutput, err
}

func (vpc *VPCService) Destroy(createArgs ...*VPCArgs) error {
	if vpc.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := vpc.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs)
	err := runTerraformDestroyWithArgs(vpc.Context, vpc.ManifestDir, args)

	return err
}

func NewVPCService(manifestDir ...string) *VPCService {
	vpc := &VPCService{}
	vpc.Init(manifestDir...)
	return vpc
}

func CreateAWSVPC(vpcArgs *VPCArgs, arg ...string) (
	privateSubnets []string,
	publicSubnets []string,
	zones []string,
	err error) {
	parambytes, _ := json.Marshal(vpcArgs)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	err = runTerraformInit(context.TODO(), CON.AWSVPCDir)
	if err != nil {
		return nil, nil, nil, err
	}

	combinedArgs := combineArgs(args, arg...)
	_, err = runTerraformApplyWithArgs(context.TODO(), CON.AWSVPCDir, combinedArgs)
	if err != nil {
		return nil, nil, nil, err
	}
	return GetVPCOutputs()
}
func GetVPCOutputs() (privateSubnets []string, publicSubnets []string, zones []string, err error) {
	out, err := runTerraformOutput(context.TODO(), CON.AWSVPCDir)
	if err != nil {
		return nil, nil, nil, err
	}
	privateObj := out["cluster-private-subnet"]
	publicObj := out["cluster-public-subnet"]
	zonesObj := out["azs"]
	privateSubnets = h.DigStringArray(privateObj, "value")
	publicSubnets = h.DigStringArray(publicObj, "value")
	zones = h.DigStringArray(zonesObj, "value")
	return
}

func DestroyAWSVPC(vpcArgs *VPCArgs, arg ...string) error {
	parambytes, _ := json.Marshal(vpcArgs)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	combinedArgs := combineArgs(args, arg...)
	err := runTerraformDestroyWithArgs(context.TODO(), CON.AWSVPCDir, combinedArgs)
	return err
}
