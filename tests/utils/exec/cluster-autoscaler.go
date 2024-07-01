package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type ClusterAutoscalerArgs struct {
	Cluster                     *string         `hcl:"cluster_id"`
	BalanceSimilarNodeGroups    *bool           `hcl:"balance_similar_node_groups"`
	SkipNodesWithLocalStorage   *bool           `hcl:"skip_nodes_with_local_storage"`
	LogVerbosity                *int            `hcl:"log_verbosity"`
	MaxPodGracePeriod           *int            `hcl:"max_pod_grace_period"`
	PodPriorityThreshold        *int            `hcl:"pod_priority_threshold"`
	IgnoreDaemonsetsUtilization *bool           `hcl:"ignore_daemonsets_utilization"`
	MaxNodeProvisionTime        *string         `hcl:"max_node_provision_time"`
	BalancingIgnoredLabels      *[]string       `hcl:"balancing_ignored_labels"`
	ResourceLimits              *ResourceLimits `hcl:"resource_limits"`
	ScaleDown                   *ScaleDown      `hcl:"scale_down"`
}

type ResourceLimits struct {
	Cores         *ResourceRange `cty:"cores"`
	MaxNodesTotal *int           `cty:"max_nodes_total"`
	Memory        *ResourceRange `cty:"memory"`
}

type ScaleDown struct {
	DelayAfterAdd        *string `cty:"delay_after_add"`
	DelayAfterDelete     *string `cty:"delay_after_delete"`
	DelayAfterFailure    *string `cty:"delay_after_failure"`
	UnneededTime         *string `cty:"unneeded_time"`
	UtilizationThreshold *string `cty:"utilization_threshold"`
	Enabled              *bool   `cty:"enabled"`
}

type ResourceRange struct {
	Max *int `cty:"max"`
	Min *int `cty:"min"`
}

type ClusterAutoscalerOutput struct {
	Cluster                     string   `json:"cluster_id,omitempty"`
	BalanceSimilarNodeGroups    bool     `json:"balance_similar_node_groups,omitempty"`
	SkipNodesWithLocalStorage   bool     `json:"skip_nodes_with_local_storage,omitempty"`
	LogVerbosity                int      `json:"log_verbosity,omitempty"`
	MaxPodGracePeriod           int      `json:"max_pod_grace_period,omitempty"`
	PodPriorityThreshold        int      `json:"pod_priority_threshold,omitempty"`
	IgnoreDaemonsetsUtilization bool     `json:"ignore_daemonsets_utilization,omitempty"`
	MaxNodeProvisionTime        string   `json:"max_node_provision_time,omitempty"`
	BalancingIgnoredLabels      []string `json:"balancing_ignored_labels,omitempty"`
	MaxNodesTotal               int      `json:"max_nodes_total,omitempty"`
	MinCores                    int      `json:"min_cores,omitempty"`
	MaxCores                    int      `json:"max_cores,omitempty"`
	MinMemory                   int      `json:"min_memory,omitempty"`
	MaxMemory                   int      `json:"max_memory,omitempty"`
	DelayAfterAdd               string   `json:"delay_after_add,omitempty"`
	DelayAfterDelete            string   `json:"delay_after_delete,omitempty"`
	DelayAfterFailure           string   `json:"delay_after_failure,omitempty"`
	UnneededTime                string   `json:"unneeded_time,omitempty"`
	UtilizationThreshold        string   `json:"utilization_threshold,omitempty"`
	Enabled                     bool     `json:"enabled,omitempty"`
}

type ClusterAutoscalerService interface {
	Init() error
	Plan(args *ClusterAutoscalerArgs) (string, error)
	Apply(args *ClusterAutoscalerArgs) (string, error)
	Output() (*ClusterAutoscalerOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*ClusterAutoscalerArgs, error)
	DeleteTFVars() error
}

type clusterAutoscalerService struct {
	tfExecutor TerraformExecutor
}

func NewClusterAutoscalerService(tfWorkspace string, clusterType constants.ClusterType) (ClusterAutoscalerService, error) {
	svc := &clusterAutoscalerService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetClusterAutoscalerManifestsDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *clusterAutoscalerService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *clusterAutoscalerService) Plan(args *ClusterAutoscalerArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *clusterAutoscalerService) Apply(args *ClusterAutoscalerArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *clusterAutoscalerService) Output() (*ClusterAutoscalerOutput, error) {
	var output ClusterAutoscalerOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *clusterAutoscalerService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *clusterAutoscalerService) ReadTFVars() (*ClusterAutoscalerArgs, error) {
	args := &ClusterAutoscalerArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *clusterAutoscalerService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
