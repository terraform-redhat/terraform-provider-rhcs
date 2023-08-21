package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
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
type MachinePoolService struct {
	CreationArgs *MachinePoolArgs
	ManifestDir  string
	Context      context.Context
}

func (mp *MachinePoolService) Init(manifestDirs ...string) error {
	mp.ManifestDir = CON.AWSVPCDir
	if len(manifestDirs) != 0 {
		mp.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	mp.Context = ctx
	err := runTerraformInit(ctx, mp.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (mp *MachinePoolService) Create(createArgs *MachinePoolArgs, extraArgs ...string) error {
	mp.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(mp.Context, mp.ManifestDir, args)
	if err != nil {
		return err
	}
	return nil
}

func (mp *MachinePoolService) Output() (mpName string, err error) {
	mpDir := CON.MachinePoolDir
	if mp.ManifestDir != "" {
		mpDir = mp.ManifestDir
	}
	out, err := runTerraformOutput(context.TODO(), mpDir)
	if err != nil {
		return "", err
	}
	mpNameObj := out["id"]
	mpName = h.DigString(mpNameObj, "value")
	return
}

func (mp *MachinePoolService) Destroy(createArgs ...*MachinePoolArgs) error {
	if mp.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := mp.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs)
	err := runTerraformDestroyWithArgs(mp.Context, mp.ManifestDir, args)
	// if err != nil {
	// 	return err
	// }

	// getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id")
	// getClusterIdCmd.Dir = targetDir
	// _, err = getClusterIdCmd.Output()

	return err
}

func NewMachinePoolService(manifestDir ...string) *MachinePoolService {
	mp := &MachinePoolService{}
	mp.Init(manifestDir...)
	return mp
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
