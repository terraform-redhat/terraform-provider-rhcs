package exec

import (
	"context"
	"encoding/json"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type VPCVariables struct {
	Name      string   `json:"name,omitempty"`
	AZIDs     []string `json:"az_ids,omitempty"`
	AWSRegion string   `json:"aws_region,omitempty"`
	MultiAZ   bool     `json:"multi_az,omitempty"`
	VPCCIDR   string   `json:"vpc_cidr,omitempty"`
}

// ************ AWS resources ***************************
func CreateAWSVPC(vpcArgs *VPCVariables, arg ...string) (
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

func DestroyAWSVPC(vpcArgs *VPCVariables, arg ...string) error {
	parambytes, _ := json.Marshal(vpcArgs)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	combinedArgs := combineArgs(args, arg...)
	err := runTerraformDestroyWithArgs(context.TODO(), CON.AWSVPCDir, combinedArgs)
	return err
}
