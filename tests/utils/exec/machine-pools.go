package exec

import (
	"fmt"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type MachinePoolArgs struct {
	Cluster                  string               `json:"cluster,omitempty"`
	OCMENV                   string               `json:"ocm_environment,omitempty"`
	Name                     *string              `json:"name,omitempty"`
	URL                      string               `json:"url,omitempty"`
	MachineType              *string              `json:"machine_type,omitempty"`
	Replicas                 *int                 `json:"replicas,omitempty"`
	AutoscalingEnabled       *bool                `json:"autoscaling_enabled,omitempty"`
	UseSpotInstances         *bool                `json:"use_spot_instances,omitempty"`
	MaxReplicas              *int                 `json:"max_replicas,omitempty"`
	MinReplicas              *int                 `json:"min_replicas,omitempty"`
	MaxSpotPrice             *float64             `json:"max_spot_price,omitempty"`
	Labels                   *map[string]string   `json:"labels,omitempty"`
	Taints                   *[]map[string]string `json:"taints,omitempty"`
	ID                       *string              `json:"id,omitempty"`
	AvailabilityZone         *string              `json:"availability_zone,omitempty"`
	SubnetID                 *string              `json:"subnet_id,omitempty"`
	MultiAZ                  *bool                `json:"multi_availability_zone,omitempty"`
	DiskSize                 *int                 `json:"disk_size,omitempty"`
	AdditionalSecurityGroups *[]string            `json:"additional_security_groups,omitempty"`
}

func (args *MachinePoolArgs) appendURL() {
	args.URL = CON.GateWayURL
}

type MachinePoolOutput struct {
	ID                 string            `json:"machine_pool_id,omitempty"`
	Name               string            `json:"name,omitempty"`
	ClusterID          string            `json:"cluster_id,omitempty"`
	Replicas           int               `json:"replicas,omitempty"`
	MachineType        string            `json:"machine_type,omitempty"`
	AutoscalingEnabled bool              `json:"autoscaling_enabled,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
}

type MachinePoolService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newMachinePoolService(clusterType CON.ClusterType, tfExec TerraformExec) (*MachinePoolService, error) {
	mp := &MachinePoolService{
		ManifestDir: GetMachinePoolsManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := mp.Init()
	return mp, err
}

func (mp *MachinePoolService) Init() error {
	return mp.tfExec.RunTerraformInit(mp.ManifestDir)

}

func (mp *MachinePoolService) Apply(createArgs *MachinePoolArgs) (output string, err error) {
	createArgs.appendURL()

	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = mp.tfExec.RunTerraformApply(mp.ManifestDir, tfVars)
	return
}

func (mp *MachinePoolService) Plan(planArgs *MachinePoolArgs) (output string, err error) {
	planArgs.appendURL()

	var tfVars *TFVars
	tfVars, err = NewTFArgs(planArgs)
	if err != nil {
		return
	}
	output, err = mp.tfExec.RunTerraformPlan(mp.ManifestDir, tfVars)
	return
}

func (mp *MachinePoolService) Output() (*MachinePoolOutput, error) {
	out, err := mp.tfExec.RunTerraformOutput(mp.ManifestDir)
	if err != nil {
		return nil, err
	}
	output := &MachinePoolOutput{
		Replicas:           h.DigInt(out["replicas"], "value"),
		MachineType:        h.DigString(out["machine_type"], "value"),
		Name:               h.DigString(out["name"], "value"),
		AutoscalingEnabled: h.DigBool(out["autoscaling_enabled"]),
	}
	return output, nil
}

func (mp *MachinePoolService) Destroy(deleteTFVars bool) (output string, err error) {
	return mp.tfExec.RunTerraformDestroy(mp.ManifestDir, deleteTFVars)
}

func (mp *MachinePoolService) ShowState(mpResourceName string) (output string, err error) {
	args := fmt.Sprintf("%s.%s", constants.MachinePoolResourceKind, mpResourceName)
	output, err = mp.tfExec.RunTerraformState(mp.ManifestDir, "show", args)
	return
}

func BuildDefaultMachinePoolArgsFromClusterState(clusterResource interface{}) (MachinePoolArgs, error) {
	var machinePoolArgs MachinePoolArgs
	if h.DigString(clusterResource, "type") != "rhcs_cluster_rosa_classic" {
		return machinePoolArgs, fmt.Errorf("expected a cluster resource of type rhcs_cluster_rosa_classic, got %s", h.DigString(clusterResource, "type"))
	}
	if h.DigBool(h.DigArray(clusterResource, "instances")[0], "attributes", "autoscaling_enabled") {
		machinePoolArgs.AutoscalingEnabled = h.BoolPointer(true)
		maxReplicas := h.DigInt(h.DigArray(clusterResource, "instances")[0], "attributes", "max_replicas")
		minReplicas := h.DigInt(h.DigArray(clusterResource, "instances")[0], "attributes", "min_replicas")
		machinePoolArgs.MaxReplicas = &maxReplicas
		machinePoolArgs.MinReplicas = &minReplicas
	} else {
		replicas := h.DigInt(h.DigArray(clusterResource, "instances")[0], "attributes", "replicas")
		machinePoolArgs.Replicas = &replicas
	}
	machinePoolArgs.Name = h.StringPointer("worker")
	machinePoolArgs.MachineType = h.StringPointer(h.DigString(h.DigArray(clusterResource, "instances")[0], "attributes", "compute_machine_type"))
	labelsInterface := h.DigObject(h.DigArray(clusterResource, "instances")[0], "attributes", "default_mp_labels")
	labels := make(map[string]string)
	for key, value := range labelsInterface.(map[string]interface{}) {
		labels[key] = fmt.Sprint(value)
	}
	machinePoolArgs.Labels = &labels
	return machinePoolArgs, nil
}

func BuildDefaultMachinePoolArgsFromDefaultMachinePoolState(defaultMachinePoolResource interface{}) (MachinePoolArgs, error) {
	var machinePoolArgs MachinePoolArgs
	if h.DigString(defaultMachinePoolResource, "type") != "rhcs_machine_pool" && h.DigString(defaultMachinePoolResource, "name") != "worker" {
		return machinePoolArgs, fmt.Errorf("expected a default machinepool resource of type rhcs_machine_pool and named worker, got %s named %s", h.DigString(defaultMachinePoolResource, "type"), h.DigString(defaultMachinePoolResource, "name"))
	}
	if h.DigBool(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "autoscaling_enabled") {
		machinePoolArgs.AutoscalingEnabled = h.BoolPointer(true)
		maxReplicas := h.DigInt(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "max_replicas")
		minReplicas := h.DigInt(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "min_replicas")
		machinePoolArgs.MaxReplicas = &maxReplicas
		machinePoolArgs.MinReplicas = &minReplicas
	} else {
		replicas := h.DigInt(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "replicas")
		machinePoolArgs.Replicas = &replicas
	}
	machinePoolArgs.Name = h.StringPointer("worker")
	machinePoolArgs.MachineType = h.StringPointer(h.DigString(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "machine_type"))
	labelsInterface := h.DigObject(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "labels")
	labels := make(map[string]string)
	for key, value := range labelsInterface.(map[string]interface{}) {
		labels[key] = fmt.Sprint(value)
	}
	machinePoolArgs.Labels = &labels
	if h.DigObject(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "taints") != nil {
		taintsInterface := h.DigArray(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "taints")
		taints := make([]map[string]string, len(taintsInterface))
		for i, taintInterface := range taintsInterface {
			taintMap := taintInterface.(map[string]interface{})
			taint := make(map[string]string)
			for key, value := range taintMap {
				taint[key] = fmt.Sprint(value)
			}
			taints[i] = taint
		}
		machinePoolArgs.Taints = &taints
	}
	return machinePoolArgs, nil
}

/*
	    This func will return MachinePoolArgs with mapping the values from the backend
		    Name    done
			MachineType      				done
			Replicas         				done
			AutoscalingEnabled  			done
			UseSpotInstances     			done
			MaxReplicas       				done
			MinReplicas      				done
			MaxSpotPrice           			done
			Labels                			done
			Taints                			done
			ID                  		It's same with machinepool name and not required.
			AvailabilityZone     	    Todo
			SubnetID                	Todo
			MultiAZ                     not a part of CMS machinepool endpoint reponse till now
			DiskSize                 	done
			AdditionalSecurityGroups *[]string           done
*/
func BuildMachinePoolArgsFromCSResponse(machinePool *cmv1.MachinePool) (machinePoolArgs *MachinePoolArgs) {
	machinePoolArgs = &MachinePoolArgs{}

	if id, ok := machinePool.GetID(); ok {
		machinePoolArgs.Name = h.StringPointer(id)
	}
	if replicas, ok := machinePool.GetReplicas(); ok {
		machinePoolArgs.Replicas = h.IntPointer(replicas)
	}
	if instanceType, ok := machinePool.GetInstanceType(); ok {
		machinePoolArgs.MachineType = h.StringPointer(instanceType)
	}
	if autoscalingEnabled, ok := machinePool.GetAutoscaling(); ok {
		machinePoolArgs.MaxReplicas = h.IntPointer(autoscalingEnabled.MaxReplicas())
		machinePoolArgs.MinReplicas = h.IntPointer(autoscalingEnabled.MinReplicas())
	}

	if maxSpotPrice, ok := machinePool.AWS().GetSpotMarketOptions(); ok {
		machinePoolArgs.MaxSpotPrice = h.Float64Pointer(maxSpotPrice.MaxPrice())
		machinePoolArgs.UseSpotInstances = h.BoolPointer(true)
	}
	if labels, ok := machinePool.GetLabels(); ok {
		machinePoolArgs.Labels = &labels
	}
	if taints, ok := machinePool.GetTaints(); ok {
		taintsMap := make([]map[string]string, len(taints))
		for i, taint := range taints {
			taintMap := make(map[string]string)
			if key, ok := taint.GetKey(); ok {
				taintMap["key"] = key
			}
			if value, ok := taint.GetValue(); ok {
				taintMap["value"] = value
			}
			if effect, ok := taint.GetEffect(); ok {
				taintMap["effect"] = string(effect)
			}
			taintsMap[i] = taintMap
		}
		machinePoolArgs.Taints = &taintsMap
	}
	if diskSize, ok := machinePool.GetRootVolume(); ok {
		machinePoolArgs.DiskSize = h.IntPointer(diskSize.AWS().Size())
	}
	if additionalSecurityGroups, ok := machinePool.AWS().GetAdditionalSecurityGroupIds(); ok {
		machinePoolArgs.AdditionalSecurityGroups = &additionalSecurityGroups
	}
	return machinePoolArgs
}
