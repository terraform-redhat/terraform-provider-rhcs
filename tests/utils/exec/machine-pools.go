package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

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

func CreateMachinePool(ctx context.Context, args ...string) (string, error) {
	runTerraformInit(ctx, CON.MachinePoolDir)

	runTerraformApplyWithArgs(ctx, CON.MachinePoolDir, args)

	getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id")
	getClusterIdCmd.Dir = CON.MachinePoolDir
	output, err := getClusterIdCmd.Output()
	if err != nil {
		return "", err
	}

	splitOutput := strings.Split(string(output), "\"")
	if len(splitOutput) <= 1 {
		return "", fmt.Errorf("got no cluster id from the output")
	}

	return splitOutput[1], nil
}

func CreateTFMachinePool(ctx context.Context,
	varArgs map[string]interface{}, abArgs ...string) (string, error) {
	err := runTerraformInit(ctx, CON.MachinePoolDir)
	if err != nil {
		return "", err
	}

	args := combineArgs(varArgs, abArgs...)
	_, err = runTerraformApplyWithArgs(ctx, CON.MachinePoolDir, args)
	if err != nil {
		return "", err
	}

	getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id")
	getClusterIdCmd.Dir = CON.MachinePoolDir
	output, err := getClusterIdCmd.Output()
	if err != nil {
		return "", err
	}

	splitOutput := strings.Split(string(output), "\"")
	if len(splitOutput) <= 1 {
		return "", fmt.Errorf("got no cluster id from the output")
	}

	return splitOutput[1], nil
}

func DestroyTFMachinePool(ctx context.Context,
	varArgs map[string]interface{}, abArgs ...string) error {
	err := runTerraformInit(ctx, CON.MachinePoolDir)
	if err != nil {
		return err
	}

	args := combineArgs(varArgs, abArgs...)
	err = runTerraformDestroyWithArgs(ctx, CON.MachinePoolDir, args)
	return err
}

func CreateMyTFMachinePool(clusterArgs *MachinePoolArgs, arg ...string) (string, error) {
	parambytes, _ := json.Marshal(clusterArgs)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	return CreateTFMachinePool(context.TODO(), args, arg...)
}

func DestroyMyTFMachinePool(mpArgs *MachinePoolArgs, arg ...string) error {
	parambytes, _ := json.Marshal(mpArgs)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	return DestroyTFMachinePool(context.TODO(), args, arg...)
}
