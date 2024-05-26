package exec

import (
	"context"
	"fmt"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type MachinePoolArgs struct {
	Cluster                  *string              `json:"cluster,omitempty"`
	Name                     *string              `json:"name,omitempty"`
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
	Tags                     *map[string]string   `json:"tags,omitempty"`

	// HCP supported
	TuningConfigs              *[]string `json:"tuning_configs,omitempty"`
	UpgradeAcknowledgementsFor *string   `json:"upgrade_acknowledgements_for,omitempty"`
	OpenshiftVersion           *string   `json:"openshift_version,omitempty"`
	AutoRepair                 *bool     `json:"auto_repair,omitempty"`
}

type MachinePoolService struct {
	CreationArgs *MachinePoolArgs
	ManifestDir  string
	Context      context.Context
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

func (mp *MachinePoolService) Init(manifestDirs ...string) error {
	mp.ManifestDir = constants.ClassicMachinePoolDir
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

func (mp *MachinePoolService) MagicImport(createArgs *MachinePoolArgs, extraArgs ...string) error {
	mp.CreationArgs = createArgs
	args, _ := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(mp.Context, mp.ManifestDir, args...)
	if err != nil {
		return err
	}
	return nil
}

func (mp *MachinePoolService) Apply(createArgs *MachinePoolArgs, recordtfargs bool, extraArgs ...string) (string, error) {
	mp.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	output, err := runTerraformApply(mp.Context, mp.ManifestDir, args...)
	if err == nil && recordtfargs {
		recordTFvarsFile(mp.ManifestDir, tfvars)
	}
	return output, err
}

func (mp *MachinePoolService) Plan(createArgs *MachinePoolArgs, extraArgs ...string) (string, error) {
	mp.CreationArgs = createArgs
	args, _ := combineStructArgs(createArgs, extraArgs...)
	output, err := runTerraformPlan(mp.Context, mp.ManifestDir, args...)
	return output, err
}

func (mp *MachinePoolService) Output() (MachinePoolOutput, error) {
	mpDir := mp.ManifestDir
	if mp.ManifestDir == "" {
		mpDir = constants.ClassicMachinePoolDir
	}
	var output MachinePoolOutput
	out, err := runTerraformOutput(context.TODO(), mpDir)
	if err != nil {
		return output, err
	}
	if err != nil {
		return output, err
	}
	replicas := helper.DigInt(out["replicas"], "value")
	machine_type := helper.DigString(out["machine_type"], "value")
	name := helper.DigString(out["name"], "value")
	autoscaling_enabled := helper.DigBool(out["autoscaling_enabled"])
	output = MachinePoolOutput{
		Replicas:           replicas,
		MachineType:        machine_type,
		Name:               name,
		AutoscalingEnabled: autoscaling_enabled,
	}
	return output, nil
}

func (mp *MachinePoolService) Destroy(createArgs ...*MachinePoolArgs) (output string, err error) {
	if mp.CreationArgs == nil && len(createArgs) == 0 {
		return "", fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := mp.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)

	return runTerraformDestroy(mp.Context, mp.ManifestDir, args...)
}

func NewMachinePoolService(manifestDir ...string) *MachinePoolService {
	mp := &MachinePoolService{}
	mp.Init(manifestDir...)
	return mp
}

func BuildDefaultMachinePoolArgsFromClusterState(clusterResource interface{}) (MachinePoolArgs, error) {
	var machinePoolArgs MachinePoolArgs
	if helper.DigString(clusterResource, "type") != "rhcs_cluster_rosa_classic" {
		return machinePoolArgs, fmt.Errorf("expected a cluster resource of type rhcs_cluster_rosa_classic, got %s", helper.DigString(clusterResource, "type"))
	}
	if helper.DigBool(helper.DigArray(clusterResource, "instances")[0], "attributes", "autoscaling_enabled") {
		machinePoolArgs.AutoscalingEnabled = helper.BoolPointer(true)
		maxReplicas := helper.DigInt(helper.DigArray(clusterResource, "instances")[0], "attributes", "max_replicas")
		minReplicas := helper.DigInt(helper.DigArray(clusterResource, "instances")[0], "attributes", "min_replicas")
		machinePoolArgs.MaxReplicas = &maxReplicas
		machinePoolArgs.MinReplicas = &minReplicas
	} else {
		replicas := helper.DigInt(helper.DigArray(clusterResource, "instances")[0], "attributes", "replicas")
		machinePoolArgs.Replicas = &replicas
	}
	machinePoolArgs.Name = helper.StringPointer("worker")
	machinePoolArgs.MachineType = helper.StringPointer(helper.DigString(helper.DigArray(clusterResource, "instances")[0], "attributes", "compute_machine_type"))
	labelsInterface := helper.DigObject(helper.DigArray(clusterResource, "instances")[0], "attributes", "default_mp_labels")
	labels := make(map[string]string)
	for key, value := range labelsInterface.(map[string]interface{}) {
		labels[key] = fmt.Sprint(value)
	}
	machinePoolArgs.Labels = &labels
	return machinePoolArgs, nil
}

func BuildDefaultMachinePoolArgsFromDefaultMachinePoolState(defaultMachinePoolResource interface{}) (MachinePoolArgs, error) {
	var machinePoolArgs MachinePoolArgs
	if helper.DigString(defaultMachinePoolResource, "type") != "rhcs_machine_pool" && helper.DigString(defaultMachinePoolResource, "name") != "worker" {
		return machinePoolArgs, fmt.Errorf("expected a default machinepool resource of type rhcs_machine_pool and named worker, got %s named %s", helper.DigString(defaultMachinePoolResource, "type"), helper.DigString(defaultMachinePoolResource, "name"))
	}
	if helper.DigBool(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "autoscaling_enabled") {
		machinePoolArgs.AutoscalingEnabled = helper.BoolPointer(true)
		maxReplicas := helper.DigInt(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "max_replicas")
		minReplicas := helper.DigInt(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "min_replicas")
		machinePoolArgs.MaxReplicas = &maxReplicas
		machinePoolArgs.MinReplicas = &minReplicas
	} else {
		replicas := helper.DigInt(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "replicas")
		machinePoolArgs.Replicas = &replicas
	}
	machinePoolArgs.Name = helper.StringPointer("worker")
	machinePoolArgs.MachineType = helper.StringPointer(helper.DigString(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "machine_type"))
	labelsInterface := helper.DigObject(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "labels")
	labels := make(map[string]string)
	for key, value := range labelsInterface.(map[string]interface{}) {
		labels[key] = fmt.Sprint(value)
	}
	machinePoolArgs.Labels = &labels
	if helper.DigObject(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "taints") != nil {
		taintsInterface := helper.DigArray(helper.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "taints")
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
func BuildMachinePoolArgsFromCSResponse(machinePool *cmv1.MachinePool) MachinePoolArgs {
	var machinePoolArgs MachinePoolArgs

	if id, ok := machinePool.GetID(); ok {
		machinePoolArgs.Name = helper.StringPointer(id)
	}
	if replicas, ok := machinePool.GetReplicas(); ok {
		machinePoolArgs.Replicas = helper.IntPointer(replicas)
	}
	if instanceType, ok := machinePool.GetInstanceType(); ok {
		machinePoolArgs.MachineType = helper.StringPointer(instanceType)
	}
	if autoscalingEnabled, ok := machinePool.GetAutoscaling(); ok {
		machinePoolArgs.MaxReplicas = helper.IntPointer(autoscalingEnabled.MaxReplicas())
		machinePoolArgs.MinReplicas = helper.IntPointer(autoscalingEnabled.MinReplicas())
	}

	if maxSpotPrice, ok := machinePool.AWS().GetSpotMarketOptions(); ok {
		machinePoolArgs.MaxSpotPrice = helper.Float64Pointer(maxSpotPrice.MaxPrice())
		machinePoolArgs.UseSpotInstances = helper.BoolPointer(true)
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
		machinePoolArgs.DiskSize = helper.IntPointer(diskSize.AWS().Size())
	}
	if additionalSecurityGroups, ok := machinePool.AWS().GetAdditionalSecurityGroupIds(); ok {
		machinePoolArgs.AdditionalSecurityGroups = &additionalSecurityGroups
	}
	return machinePoolArgs
}
