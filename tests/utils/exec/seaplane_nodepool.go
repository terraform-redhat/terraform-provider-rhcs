// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// SeaplaneNodePoolArgs contains the Terraform variable values for the
// rhcs_nodepool_seaplane manifest.
type SeaplaneNodePoolArgs struct {
	SeaplaneURL                   *string              `hcl:"seaplane_url"`
	AWSRegion                     *string              `hcl:"aws_region"`
	ClusterID                     *string              `hcl:"cluster_id"`
	ClusterName                   *string              `hcl:"cluster_name"`
	Name                          *string              `hcl:"name"`
	SubnetID                      *string              `hcl:"subnet_id"`
	AutoRepair                    *bool                `hcl:"auto_repair"`
	Replicas                      *int                 `hcl:"replicas"`
	Autoscaling                   *NodePoolAutoscaling `hcl:"autoscaling"`
	InstanceType                  *string              `hcl:"instance_type"`
	PlacementMarketType           *string              `hcl:"placement_market_type"`
	MaxPrice                      *string              `hcl:"max_price"`
	DiskSize                      *int                 `hcl:"disk_size"`
	EC2InstanceMetadataHttpTokens *string              `hcl:"ec2_instance_metadata_http_tokens"`

	// Operating system image family for the nodes in this pool. The image family
	// is baked into every instance the pool launches, so it can only be set on
	// create; changing it would replace the pool. Valid values: "Linux" (RHCOS)
	// and "Windows" (Windows Server, License Included).
	ImageType *string `hcl:"image_type"`

	// Maximum time to wait for a node to drain before forcibly terminating it
	// during an upgrade or replacement, as a Go duration string (e.g., "5m", "1h").
	NodeDrainTimeout *string `hcl:"node_drain_timeout"`

	Tags   map[string]string `hcl:"tags"`
	Labels map[string]string `hcl:"labels"`
}

type NodePoolAutoscaling struct {
	Enabled     *bool `cty:"enabled"`
	MinReplicas *int  `cty:"min_replicas"`
	MaxReplicas *int  `cty:"max_replicas"`
}

// SeaplaneNodePoolOutput holds the Terraform output values from the nodepool manifest.
type SeaplaneNodePoolOutput struct {
	NodePoolID          string `json:"nodepool_id,omitempty"`
	NodePoolName        string `json:"nodepool_name,omitempty"`
	Phase               string `json:"nodepool_phase,omitempty"`
	Replicas            *int   `json:"nodepool_replicas,omitempty"`
	PlacementMarketType string `json:"nodepool_placement_market_type,omitempty"`
	MaxPrice            string `json:"nodepool_max_price,omitempty"`
	MinReplicas         *int   `json:"nodepool_min_replicas,omitempty"`
	MaxReplicas         *int   `json:"nodepool_max_replicas,omitempty"`
}

// SeaplaneNodePoolService manages the lifecycle of an rhcs_nodepool_seaplane
// Terraform resource via a TerraformExecutor.
type SeaplaneNodePoolService interface {
	Init() error
	Apply(args *SeaplaneNodePoolArgs) (string, error)
	Output() (*SeaplaneNodePoolOutput, error)
	Destroy() (string, error)
	WriteTFVars(args *SeaplaneNodePoolArgs) error
	DeleteTFVars() error
}

type seaplaneNodePoolService struct {
	tfExecutor TerraformExecutor
}

// NewSeaplaneNodePoolService creates a SeaplaneNodePoolService rooted at the
// Seaplane nodepool tf-manifests directory.
func NewSeaplaneNodePoolService(tfWorkspace string) (SeaplaneNodePoolService, error) {
	svc := &seaplaneNodePoolService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetSeaplaneNodePoolManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *seaplaneNodePoolService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *seaplaneNodePoolService) Apply(args *SeaplaneNodePoolArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *seaplaneNodePoolService) Output() (*SeaplaneNodePoolOutput, error) {
	out := &SeaplaneNodePoolOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *seaplaneNodePoolService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *seaplaneNodePoolService) WriteTFVars(args *SeaplaneNodePoolArgs) error {
	return svc.tfExecutor.WriteTerraformVars(args)
}

func (svc *seaplaneNodePoolService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
