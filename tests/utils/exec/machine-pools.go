package exec

import (
	"fmt"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type MachinePoolArgs struct {
	Cluster                  *string              `hcl:"cluster"`
	Name                     *string              `hcl:"name"`
	MachineType              *string              `hcl:"machine_type"`
	Replicas                 *int                 `hcl:"replicas"`
	AutoscalingEnabled       *bool                `hcl:"autoscaling_enabled"`
	UseSpotInstances         *bool                `hcl:"use_spot_instances"`
	MaxReplicas              *int                 `hcl:"max_replicas"`
	MinReplicas              *int                 `hcl:"min_replicas"`
	MaxSpotPrice             *float64             `hcl:"max_spot_price"`
	Labels                   *map[string]string   `hcl:"labels"`
	Taints                   *[]map[string]string `hcl:"taints"`
	ID                       *string              `hcl:"id"`
	AvailabilityZone         *string              `hcl:"availability_zone"`
	SubnetID                 *string              `hcl:"subnet_id"`
	MultiAZ                  *bool                `hcl:"multi_availability_zone"`
	DiskSize                 *int                 `hcl:"disk_size"`
	AdditionalSecurityGroups *[]string            `hcl:"additional_security_groups"`
	Tags                     *map[string]string   `hcl:"tags"`

	// HCP supported
	TuningConfigs              *[]string `hcl:"tuning_configs"`
	UpgradeAcknowledgementsFor *string   `hcl:"upgrade_acknowledgements_for"`
	OpenshiftVersion           *string   `hcl:"openshift_version"`
	AutoRepair                 *bool     `hcl:"auto_repair"`
}

type MachinePoolOutput struct {
	ID                 string             `json:"machine_pool_id,omitempty"`
	Name               string             `json:"name,omitempty"`
	ClusterID          string             `json:"cluster_id,omitempty"`
	Replicas           int                `json:"replicas,omitempty"`
	MachineType        string             `json:"machine_type,omitempty"`
	AutoscalingEnabled bool               `json:"autoscaling_enabled,omitempty"`
	Labels             map[string]string  `json:"labels,omitempty"`
	Taints             []MachinePoolTaint `json:"taints,omitempty"`
	TuningConfigs      []string           `json:"tuning_configs,omitempty"`
}

type MachinePoolTaint struct {
	Key          string `json:"key,omitempty"`
	Value        string `json:"value,omitempty"`
	ScheduleType string `json:"schedule_type,omitempty"`
}

type MachinePoolService interface {
	Init() error
	Plan(args *MachinePoolArgs) (string, error)
	Apply(args *MachinePoolArgs) (string, error)
	Output() (*MachinePoolOutput, error)
	Destroy() (string, error)
	ShowState(resource string) (string, error)
	RemoveState(resource string) (string, error)
	ReadTFVars() (*MachinePoolArgs, error)
	DeleteTFVars() error
}

type machinePoolService struct {
	tfExecutor TerraformExecutor
}

func NewMachinePoolService(manifestsDirs ...string) (MachinePoolService, error) {
	manifestsDir := constants.ClassicMachinePoolDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &machinePoolService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *machinePoolService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *machinePoolService) Plan(args *MachinePoolArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *machinePoolService) Apply(args *MachinePoolArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *machinePoolService) Output() (*MachinePoolOutput, error) {
	var output MachinePoolOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *machinePoolService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *machinePoolService) ShowState(resource string) (string, error) {
	return svc.tfExecutor.RunTerraformState("show", resource)
}

func (svc *machinePoolService) RemoveState(resource string) (string, error) {
	return svc.tfExecutor.RunTerraformState("rm", resource)
}

func (svc *machinePoolService) ReadTFVars() (*MachinePoolArgs, error) {
	args := &MachinePoolArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *machinePoolService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
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
func BuildMachinePoolArgsFromCSResponse(clusterID string, machinePool *cmv1.MachinePool) *MachinePoolArgs {
	var machinePoolArgs MachinePoolArgs

	machinePoolArgs.Cluster = helper.StringPointer(clusterID)

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
	return &machinePoolArgs
}
