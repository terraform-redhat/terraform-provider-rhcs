package exec

***REMOVED***
	"context"
	"encoding/json"
***REMOVED***

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
***REMOVED***

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

func (vpc *VPCService***REMOVED*** Init(manifestDirs ...string***REMOVED*** error {
	vpc.ManifestDir = CON.AWSVPCDir
	if len(manifestDirs***REMOVED*** != 0 {
		vpc.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO(***REMOVED***
	vpc.Context = ctx
	err := runTerraformInit(ctx, vpc.ManifestDir***REMOVED***
	if err != nil {
		return err
	}
	return nil

}

func (vpc *VPCService***REMOVED*** Create(createArgs *VPCArgs, extraArgs ...string***REMOVED*** error {
	vpc.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	_, err := runTerraformApplyWithArgs(vpc.Context, vpc.ManifestDir, args***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

func (vpc *VPCService***REMOVED*** Output(***REMOVED*** (privateSubnets []string, publicSubnets []string, zones []string, err error***REMOVED*** {
	vpcDir := CON.AWSVPCDir
	if vpc.ManifestDir != "" {
		vpcDir = vpc.ManifestDir
	}
	out, err := runTerraformOutput(context.TODO(***REMOVED***, vpcDir***REMOVED***
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

func (vpc *VPCService***REMOVED*** Destroy(createArgs ...*VPCArgs***REMOVED*** error {
	if vpc.CreationArgs == nil && len(createArgs***REMOVED*** == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter"***REMOVED***
	}
	destroyArgs := vpc.CreationArgs
	if len(createArgs***REMOVED*** != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs***REMOVED***
	err := runTerraformDestroyWithArgs(vpc.Context, vpc.ManifestDir, args***REMOVED***
	// if err != nil {
	// 	return err
	// }

	// getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	// getClusterIdCmd.Dir = targetDir
	// _, err = getClusterIdCmd.Output(***REMOVED***

	return err
}

func NewVPCService(manifestDir ...string***REMOVED*** *VPCService {
	vpc := &VPCService{}
	vpc.Init(manifestDir...***REMOVED***
	return vpc
}

// ************ AWS resources ***************************
func CreateAWSVPC(vpcArgs *VPCArgs, arg ...string***REMOVED*** (
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

func DestroyAWSVPC(vpcArgs *VPCArgs, arg ...string***REMOVED*** error {
	parambytes, _ := json.Marshal(vpcArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	combinedArgs := combineArgs(args, arg...***REMOVED***
	err := runTerraformDestroyWithArgs(context.TODO(***REMOVED***, CON.AWSVPCDir, combinedArgs***REMOVED***
	return err
}
