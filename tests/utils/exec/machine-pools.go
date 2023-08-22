package exec

***REMOVED***
	"context"
	"encoding/json"
***REMOVED***
	"os/exec"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
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
	ID                 string              `json:"id,omitempty"`
}
type MachinePoolService struct {
	CreationArgs *MachinePoolArgs
	ManifestDir  string
	Context      context.Context
}

type MachinePoolOutput struct {
	ID                 string `json:"machine_pool_id,omitempty"`
	Name               string `json:"name,omitempty"`
	ClusterID          string `json:"cluster_id,omitempty"`
	Replicas           int    `json:"replicas,omitempty"`
	MachineType        string `json:"machine_type,omitempty"`
	AutoscalingEnabled bool   `json:"autoscaling_enabled,omitempty"`
}

func (mp *MachinePoolService***REMOVED*** Init(manifestDirs ...string***REMOVED*** error {
	mp.ManifestDir = CON.AWSVPCDir
	if len(manifestDirs***REMOVED*** != 0 {
		mp.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO(***REMOVED***
	mp.Context = ctx
	err := runTerraformInit(ctx, mp.ManifestDir***REMOVED***
	if err != nil {
		return err
	}
	return nil

}

func (mp *MachinePoolService***REMOVED*** Create(createArgs *MachinePoolArgs, extraArgs ...string***REMOVED*** error {
	mp.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	_, err := runTerraformApplyWithArgs(mp.Context, mp.ManifestDir, args***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

func (mp *MachinePoolService***REMOVED*** Output(***REMOVED*** (MachinePoolOutput, error***REMOVED*** {
	mpDir := CON.MachinePoolDir
	if mp.ManifestDir != "" {
		mpDir = mp.ManifestDir
	}
	var output MachinePoolOutput
	out, err := runTerraformOutput(context.TODO(***REMOVED***, mpDir***REMOVED***
	if err != nil {
		return output, err
	}
	if err != nil {
		return output, err
	}
	replicas := h.DigInt(out["replicas"], "value"***REMOVED***
	machine_type := h.DigString(out["machine_type"], "value"***REMOVED***
	name := h.DigString(out["name"], "value"***REMOVED***
	autoscaling_enabled := h.DigBool(out["autoscaling_enabled"]***REMOVED***
	output = MachinePoolOutput{
		Replicas:           replicas,
		MachineType:        machine_type,
		Name:               name,
		AutoscalingEnabled: autoscaling_enabled,
	}
	return output, nil
}

func (mp *MachinePoolService***REMOVED*** Destroy(createArgs ...*MachinePoolArgs***REMOVED*** error {
	if mp.CreationArgs == nil && len(createArgs***REMOVED*** == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter"***REMOVED***
	}
	destroyArgs := mp.CreationArgs
	if len(createArgs***REMOVED*** != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs***REMOVED***
	err := runTerraformDestroyWithArgs(mp.Context, mp.ManifestDir, args***REMOVED***
	// if err != nil {
	// 	return err
	// }

	// getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	// getClusterIdCmd.Dir = targetDir
	// _, err = getClusterIdCmd.Output(***REMOVED***

	return err
}

func NewMachinePoolService(manifestDir ...string***REMOVED*** *MachinePoolService {
	mp := &MachinePoolService{}
	mp.Init(manifestDir...***REMOVED***
	return mp
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
