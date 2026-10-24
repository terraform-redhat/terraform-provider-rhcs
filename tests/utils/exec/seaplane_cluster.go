// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// SeaplaneClusterArgs contains the Terraform variable values for the
// rhcs_cluster_seaplane manifest.
type SeaplaneClusterArgs struct {
	SeaplaneURL                    *string `hcl:"seaplane_url"`
	ReleaseImage                   *string `hcl:"release_image"`
	AWSRegion                      *string `hcl:"aws_region"`
	ClusterName                    *string `hcl:"cluster_name"`
	OperatorRolesPrefix            *string `hcl:"operator_roles_prefix"`
	SubnetID                       *string `hcl:"subnet_id"`
	VPCID                          *string `hcl:"vpc_id"`
	AvailabilityZone               *string `hcl:"availability_zone"`
	ExpirationTimestamp            *string `hcl:"expiration_timestamp"`
	OIDCConfigID                   *string `hcl:"oidc_config_id"`
	SchedulerProfile               *string `hcl:"scheduler_profile"`
	AutoscalerMaxNodeProvisionTime *string `hcl:"autoscaler_max_node_provision_time"`
	AutoscalerMaxNodesTotal        *int64  `hcl:"autoscaler_max_nodes_total"`
	AutoscalerMaxPodGracePeriod    *int64  `hcl:"autoscaler_max_pod_grace_period"`
	AutoscalerPodPriorityThreshold *int64  `hcl:"autoscaler_pod_priority_threshold"`
	ComponentRoutes                *string `hcl:"component_routes"`
	// EnabledCapabilities / DisabledCapabilities are JSON-encoded lists of
	// HyperShift capabilities set at cluster creation (e.g. ImageRegistry).
	// Immutable.
	EnabledCapabilities  *string `hcl:"enabled_capabilities"`
	DisabledCapabilities *string `hcl:"disabled_capabilities"`
	// Tags are customer AWS tags applied to resources for this cluster. Changes
	// are propagated to the HostedCluster and NodePools in place.
	Tags       *map[string]string `hcl:"tags"`
	HTTPProxy  *string            `hcl:"http_proxy"`
	HTTPSProxy *string            `hcl:"https_proxy"`
	NoProxy    *string            `hcl:"no_proxy"`
}

// SeaplaneClusterOutput holds the Terraform output values from the Seaplane
// cluster manifest.
type SeaplaneClusterOutput struct {
	ClusterID                      string `json:"cluster_id,omitempty"`
	ClusterName                    string `json:"cluster_name,omitempty"`
	Phase                          string `json:"cluster_phase,omitempty"`
	APIURL                         string `json:"cluster_api_url,omitempty"`
	SchedulerProfile               string `json:"scheduler_profile,omitempty"`
	AutoscalerMaxNodeProvisionTime string `json:"autoscaler_max_node_provision_time,omitempty"`
	AutoscalerMaxNodesTotal        int64  `json:"autoscaler_max_nodes_total,omitempty"`
	AutoscalerMaxPodGracePeriod    int64  `json:"autoscaler_max_pod_grace_period,omitempty"`
	AutoscalerPodPriorityThreshold int64  `json:"autoscaler_pod_priority_threshold,omitempty"`
	ComponentRoutes                string `json:"component_routes,omitempty"`
	EnabledCapabilities            string `json:"cluster_enabled_capabilities,omitempty"`
	DisabledCapabilities           string `json:"cluster_disabled_capabilities,omitempty"`
}

// SeaplaneClusterService manages the lifecycle of an rhcs_cluster_seaplane
// Terraform resource via a TerraformExecutor.
type SeaplaneClusterService interface {
	Init() error
	Apply(args *SeaplaneClusterArgs) (string, error)
	Output() (*SeaplaneClusterOutput, error)
	Destroy() (string, error)
	WriteTFVars(args *SeaplaneClusterArgs) error
	ReadTFVars() (*SeaplaneClusterArgs, error)
	DeleteTFVars() error
}

type seaplaneClusterService struct {
	tfExecutor TerraformExecutor
}

// NewSeaplaneClusterService creates a SeaplaneClusterService rooted at the
// Seaplane cluster tf-manifests directory. tfWorkspace is used to isolate
// Terraform state between parallel test runs.
func NewSeaplaneClusterService(tfWorkspace string) (SeaplaneClusterService, error) {
	svc := &seaplaneClusterService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetSeaplaneClusterManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *seaplaneClusterService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *seaplaneClusterService) Apply(args *SeaplaneClusterArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *seaplaneClusterService) Output() (*SeaplaneClusterOutput, error) {
	out := &SeaplaneClusterOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *seaplaneClusterService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *seaplaneClusterService) WriteTFVars(args *SeaplaneClusterArgs) error {
	return svc.tfExecutor.WriteTerraformVars(args)
}

func (svc *seaplaneClusterService) ReadTFVars() (*SeaplaneClusterArgs, error) {
	args := &SeaplaneClusterArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *seaplaneClusterService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
