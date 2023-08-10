package exec

***REMOVED***
	"context"
	"encoding/json"
***REMOVED***
	"os/exec"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***

type MachinePoolArgs struct {
	Cluster            string              `json:"cluster,omitempty"`
	OCMENV             string              `json:"ocm_environment,omitempty"`
	Name               string              `json:"name,omitempty"`
	Token              string              `json:"token,omitempty"`
	URL                string              `json:"url,omitempty"`
	MachineType        string              `json:"machine_type,omitempty"`
	Replicas           int                 `json:"replicas,omitempty"`
	AutoscalingEnabled bool                `json:"autoscaling_enabled,omitempty"`
	UseSpotInstances   bool                `json:"use_spot_instances,omitempty"`
	MaxReplicas        int                 `json:"max_replicas,omitempty"`
	MinReplicas        int                 `json:"min_replicas,omitempty"`
	MaxSpotPrice       float64             `json:"max_spot_price,omitempty"`
	Labels             map[string]string   `json:"labels,omitempty"`
	Taints             []map[string]string `json:"taints,omitempty"`
}

// ****************************** Machinepool CMD ***************************

func CreateMachinePool(ctx context.Context, args ...string***REMOVED*** (string, error***REMOVED*** {
	runTerraformInit(ctx, CON.MachinePoolDir***REMOVED***

	runTerraformApplyWithArgs(ctx, CON.MachinePoolDir, args***REMOVED***

	getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	getClusterIdCmd.Dir = CON.MachinePoolDir
	output, err := getClusterIdCmd.Output(***REMOVED***
	if err != nil {
		return "", err
	}

	splitOutput := strings.Split(string(output***REMOVED***, "\""***REMOVED***
	if len(splitOutput***REMOVED*** <= 1 {
		return "", fmt.Errorf("got no cluster id from the output"***REMOVED***
	}

	return splitOutput[1], nil
}

func CreateTFMachinePool(ctx context.Context,
	varArgs map[string]interface{}, abArgs ...string***REMOVED*** (string, error***REMOVED*** {
	err := runTerraformInit(ctx, CON.MachinePoolDir***REMOVED***
	if err != nil {
		return "", err
	}

	args := combineArgs(varArgs, abArgs...***REMOVED***
	_, err = runTerraformApplyWithArgs(ctx, CON.MachinePoolDir, args***REMOVED***
	if err != nil {
		return "", err
	}

	getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	getClusterIdCmd.Dir = CON.MachinePoolDir
	output, err := getClusterIdCmd.Output(***REMOVED***
	if err != nil {
		return "", err
	}

	splitOutput := strings.Split(string(output***REMOVED***, "\""***REMOVED***
	if len(splitOutput***REMOVED*** <= 1 {
		return "", fmt.Errorf("got no cluster id from the output"***REMOVED***
	}

	return splitOutput[1], nil
}

func DestroyTFMachinePool(ctx context.Context,
	varArgs map[string]interface{}, abArgs ...string***REMOVED*** error {
	err := runTerraformInit(ctx, CON.MachinePoolDir***REMOVED***
	if err != nil {
		return err
	}

	args := combineArgs(varArgs, abArgs...***REMOVED***
	err = runTerraformDestroyWithArgs(ctx, CON.MachinePoolDir, args***REMOVED***
	return err
}

func CreateMyTFMachinePool(clusterArgs *MachinePoolArgs, arg ...string***REMOVED*** (string, error***REMOVED*** {
	parambytes, _ := json.Marshal(clusterArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	return CreateTFMachinePool(context.TODO(***REMOVED***, args, arg...***REMOVED***
}

func DestroyMyTFMachinePool(mpArgs *MachinePoolArgs, arg ...string***REMOVED*** error {
	parambytes, _ := json.Marshal(mpArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	return DestroyTFMachinePool(context.TODO(***REMOVED***, args, arg...***REMOVED***
}
