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
	AZIDs     []string `json:"az_ids,omitempty"`
	AWSRegion string   `json:"aws_region,omitempty"`
	MultiAZ   bool     `json:"multi_az,omitempty"`
	VPCCIDR   string   `json:"vpc_cidr,omitempty"`
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

func (vpc *VPCService) Output() (privateSubnets []string, publicSubnets []string, zones []string, err error) {
	vpcDir := CON.AWSVPCDir
	if vpc.ManifestDir != "" {
		vpcDir = vpc.ManifestDir
	}
	out, err := runTerraformOutput(context.TODO(), vpcDir)
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
	// if err != nil {
	// 	return err
	// }

	// getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id")
	// getClusterIdCmd.Dir = targetDir
	// _, err = getClusterIdCmd.Output()

	return err
}

func NewVPCService(manifestDir ...string) *VPCService {
	vpc := &VPCService{}
	vpc.Init(manifestDir...)
	return vpc
}

// ************ AWS resources ***************************
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
