package exec

import (
	"context"
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
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
	ca.ManifestDir = constants.ClassicClusterAutoscalerDir
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
	caDir := constants.ClassicClusterAutoscalerDir
	if ca.ManifestDir != "" {
		caDir = ca.ManifestDir
	}
	var output ClusterAutoscalerOutput
	out, err := runTerraformOutput(context.TODO(), caDir)
	if err != nil {
		return output, err
	}
	output = ClusterAutoscalerOutput{
		Cluster:                     helper.DigString(out["cluster_id"], "value"),
		LogVerbosity:                helper.DigInt(out["log_verbosity"], "value"),
		BalanceSimilarNodeGroups:    helper.DigBool(out["balance_similar_node_groups"], "value"),
		SkipNodesWithLocalStorage:   helper.DigBool(out["skip_nodes_with_local_storage"], "value"),
		MaxPodGracePeriod:           helper.DigInt(out["max_pod_grace_period"], "value"),
		PodPriorityThreshold:        helper.DigInt(out["pod_priority_threshold"], "value"),
		IgnoreDaemonsetsUtilization: helper.DigBool(out["ignore_daemonsets_utilization"], "value"),
		MaxNodeProvisionTime:        helper.DigString(out["max_node_provision_time"], "value"),
		BalancingIgnoredLabels:      helper.DigArrayToString(out["balancing_ignored_labels"], "value"),
		MaxNodesTotal:               helper.DigInt(out["max_nodes_total"], "value"),
		MinCores:                    helper.DigInt(out["min_cores"], "value"),
		MaxCores:                    helper.DigInt(out["max_cores"], "value"),
		MinMemory:                   helper.DigInt(out["min_memory"], "value"),
		MaxMemory:                   helper.DigInt(out["max_memory"], "value"),
		DelayAfterAdd:               helper.DigString(out["delay_after_add"], "value"),
		DelayAfterDelete:            helper.DigString(out["delay_after_delete"], "value"),
		DelayAfterFailure:           helper.DigString(out["delay_after_failure"], "value"),
		UnneededTime:                helper.DigString(out["unneeded_time"], "value"),
		UtilizationThreshold:        helper.DigString(out["utilization_threshold"], "value"),
		Enabled:                     helper.DigBool(out["enabled"], "value"),
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
