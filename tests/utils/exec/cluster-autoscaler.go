package exec

import (
	"context"
	"fmt"

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
type ClusterAutoscalerService struct {
	CreationArgs *ClusterAutoscalerArgs
	ManifestDir  string
	Context      context.Context
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

func (ca *ClusterAutoscalerService) Init(manifestDirs ...string) error {
	ca.ManifestDir = CON.ClusterAutoscalerDir
	if len(manifestDirs) != 0 {
		ca.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	ca.Context = ctx
	err := runTerraformInit(ctx, ca.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (ca *ClusterAutoscalerService) Apply(createArgs *ClusterAutoscalerArgs, recordtfargs bool, extraArgs ...string) (string, error) {
	ca.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	output, err := runTerraformApply(ca.Context, ca.ManifestDir, args...)
	if err == nil && recordtfargs {
		recordTFvarsFile(ca.ManifestDir, tfvars)
	}
	return output, err
}

func (ca *ClusterAutoscalerService) Plan(createArgs *ClusterAutoscalerArgs, extraArgs ...string) (string, error) {
	ca.CreationArgs = createArgs
	args, _ := combineStructArgs(createArgs, extraArgs...)
	output, err := runTerraformPlan(ca.Context, ca.ManifestDir, args...)
	return output, err
}

func (ca *ClusterAutoscalerService) Output() (ClusterAutoscalerOutput, error) {
	caDir := CON.ClusterAutoscalerDir
	if ca.ManifestDir != "" {
		caDir = ca.ManifestDir
	}
	var output ClusterAutoscalerOutput
	out, err := runTerraformOutput(context.TODO(), caDir)
	if err != nil {
		return output, err
	}
	output = ClusterAutoscalerOutput{
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

func (ca *ClusterAutoscalerService) Destroy(createArgs ...*ClusterAutoscalerArgs) (output string, err error) {
	if ca.CreationArgs == nil && len(createArgs) == 0 {
		return "", fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := ca.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)

	return runTerraformDestroy(ca.Context, ca.ManifestDir, args...)
}

func NewClusterAutoscalerService(manifestDir ...string) *ClusterAutoscalerService {
	ca := &ClusterAutoscalerService{}
	ca.Init(manifestDir...)
	return ca
}
