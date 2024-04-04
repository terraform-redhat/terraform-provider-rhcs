package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type ClusterAutoscalerArgs struct {
	Cluster                     string          `json:"cluster_id,omitempty"`
	OCMENV                      string          `json:"ocm_environment,omitempty"`
	BalanceSimilarNodeGroups    bool            `json:"balance_similar_node_groups,omitempty"`
	SkipNodesWithLocalStorage   bool            `json:"skip_nodes_with_local_storage,omitempty"`
	LogVerbosity                int             `json:"log_verbosity,omitempty"`
	MaxPodGracePeriod           int             `json:"max_pod_grace_period,omitempty"`
	PodPriorityThreshold        int             `json:"pod_priority_threshold,omitempty"`
	IgnoreDaemonsetsUtilization bool            `json:"ignore_daemonsets_utilization,omitempty"`
	MaxNodeProvisionTime        string          `json:"max_node_provision_time,omitempty"`
	BalancingIgnoredLabels      []string        `json:"balancing_ignored_labels,omitempty"`
	ResourceLimits              *ResourceLimits `json:"resource_limits,omitempty"`
	ScaleDown                   *ScaleDown      `json:"scale_down,omitempty"`
}
type ResourceLimits struct {
	Cores         *ResourceRange `json:"cores,omitempty"`
	MaxNodesTotal int            `json:"max_nodes_total,omitempty"`
	Memory        *ResourceRange `json:"memory,omitempty"`
}
type ScaleDown struct {
	DelayAfterAdd        string `json:"delay_after_add,omitempty"`
	DelayAfterDelete     string `json:"delay_after_delete,omitempty"`
	DelayAfterFailure    string `json:"delay_after_failure,omitempty"`
	UnneededTime         string `json:"unneeded_time,omitempty"`
	UtilizationThreshold string `json:"utilization_threshold,omitempty"`
	Enabled              bool   `json:"enabled,omitempty"`
}
type ResourceRange struct {
	Max int `json:"max,omitempty"`
	Min int `json:"min,omitempty"`
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

type ClusterAutoscalerService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newClusterAutoscalerService(clusterType CON.ClusterType, tfExec TerraformExec) (*ClusterAutoscalerService, error) {
	ca := &ClusterAutoscalerService{
		ManifestDir: GetClusterAutoscalerDir(clusterType),
		tfExec:      tfExec,
	}

	err := ca.Init()
	return ca, err
}

func (ca *ClusterAutoscalerService) Init(manifestDirs ...string) error {
	return ca.tfExec.RunTerraformInit(ca.ManifestDir)
}

func (ca *ClusterAutoscalerService) Apply(createArgs *ClusterAutoscalerArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = ca.tfExec.RunTerraformApply(ca.ManifestDir, tfVars)
	return
}

func (ca *ClusterAutoscalerService) Plan(planArgs *ClusterAutoscalerArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(planArgs)
	if err != nil {
		return
	}
	output, err = ca.tfExec.RunTerraformPlan(ca.ManifestDir, tfVars)
	return
}

func (ca *ClusterAutoscalerService) Output() (*ClusterAutoscalerOutput, error) {
	out, err := ca.tfExec.RunTerraformOutput(ca.ManifestDir)
	if err != nil {
		return nil, err
	}
	output := &ClusterAutoscalerOutput{
		Cluster:                     h.DigString(out["cluster_id"], "value"),
		LogVerbosity:                h.DigInt(out["log_verbosity"], "value"),
		BalanceSimilarNodeGroups:    h.DigBool(out["balance_similar_node_groups"], "value"),
		SkipNodesWithLocalStorage:   h.DigBool(out["skip_nodes_with_local_storage"], "value"),
		MaxPodGracePeriod:           h.DigInt(out["max_pod_grace_period"], "value"),
		PodPriorityThreshold:        h.DigInt(out["pod_priority_threshold"], "value"),
		IgnoreDaemonsetsUtilization: h.DigBool(out["ignore_daemonsets_utilization"], "value"),
		MaxNodeProvisionTime:        h.DigString(out["max_node_provision_time"], "value"),
		BalancingIgnoredLabels:      h.DigArrayToString(out["balancing_ignored_labels"], "value"),
		MaxNodesTotal:               h.DigInt(out["max_nodes_total"], "value"),
		MinCores:                    h.DigInt(out["min_cores"], "value"),
		MaxCores:                    h.DigInt(out["max_cores"], "value"),
		MinMemory:                   h.DigInt(out["min_memory"], "value"),
		MaxMemory:                   h.DigInt(out["max_memory"], "value"),
		DelayAfterAdd:               h.DigString(out["delay_after_add"], "value"),
		DelayAfterDelete:            h.DigString(out["delay_after_delete"], "value"),
		DelayAfterFailure:           h.DigString(out["delay_after_failure"], "value"),
		UnneededTime:                h.DigString(out["unneeded_time"], "value"),
		UtilizationThreshold:        h.DigString(out["utilization_threshold"], "value"),
		Enabled:                     h.DigBool(out["enabled"], "value"),
	}
	return output, nil
}

func (ca *ClusterAutoscalerService) Destroy(deleteTFVars bool) (string, error) {
	return ca.tfExec.RunTerraformDestroy(ca.ManifestDir, deleteTFVars)
}
