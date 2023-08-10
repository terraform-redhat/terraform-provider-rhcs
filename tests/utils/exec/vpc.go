package exec

***REMOVED***
	"context"
	"encoding/json"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
***REMOVED***

type VPCVariables struct {
	Name      string   `json:"name,omitempty"`
	AZIDs     []string `json:"az_ids,omitempty"`
	AWSRegion string   `json:"aws_region,omitempty"`
	MultiAZ   bool     `json:"multi_az,omitempty"`
	VPCCIDR   string   `json:"vpc_cidr,omitempty"`
}

// ************ AWS resources ***************************
func CreateAWSVPC(vpcArgs *VPCVariables, arg ...string***REMOVED*** (
	privateSubnets []string,
	publicSubnets []string,
	zones []string,
	err error***REMOVED*** {
	parambytes, _ := json.Marshal(vpcArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	err = runTerraformInit(context.TODO(***REMOVED***, CON.AWSVPCDir***REMOVED***
	if err != nil {
		return nil, nil, nil, err
	}

	combinedArgs := combineArgs(args, arg...***REMOVED***
	_, err = runTerraformApplyWithArgs(context.TODO(***REMOVED***, CON.AWSVPCDir, combinedArgs***REMOVED***
	if err != nil {
		return nil, nil, nil, err
	}
	return GetVPCOutputs(***REMOVED***
}
func GetVPCOutputs(***REMOVED*** (privateSubnets []string, publicSubnets []string, zones []string, err error***REMOVED*** {
	out, err := runTerraformOutput(context.TODO(***REMOVED***, CON.AWSVPCDir***REMOVED***
	if err != nil {
		return nil, nil, nil, err
	}
	privateObj := out["cluster-private-subnet"]
	publicObj := out["cluster-public-subnet"]
	zonesObj := out["azs"]
	privateSubnets = h.DigStringArray(privateObj, "value"***REMOVED***
	publicSubnets = h.DigStringArray(publicObj, "value"***REMOVED***
	zones = h.DigStringArray(zonesObj, "value"***REMOVED***
	return
}

func DestroyAWSVPC(vpcArgs *VPCVariables, arg ...string***REMOVED*** error {
	parambytes, _ := json.Marshal(vpcArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	combinedArgs := combineArgs(args, arg...***REMOVED***
	err := runTerraformDestroyWithArgs(context.TODO(***REMOVED***, CON.AWSVPCDir, combinedArgs***REMOVED***
	return err
}
