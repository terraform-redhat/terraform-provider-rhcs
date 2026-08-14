// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// HyperfleetNodePoolArgs contains the Terraform variable values for the
// rhcs_nodepool_hyperfleet manifest.
type HyperfleetNodePoolArgs struct {
	HyperfleetURL *string            `hcl:"hyperfleet_url"`
	AWSRegion     *string            `hcl:"aws_region"`
	ClusterID     *string            `hcl:"cluster_id"`
	Name          *string            `hcl:"name"`
	SubnetID   *string `hcl:"subnet_id"`
	AutoRepair *bool   `hcl:"auto_repair"`
	Replicas      *int               `hcl:"replicas"`
	InstanceType  *string            `hcl:"instance_type"`
	DiskSize      *int               `hcl:"disk_size"`
	Tags          map[string]string  `hcl:"tags"`
	Labels        map[string]string  `hcl:"labels"`
}

// HyperfleetNodePoolOutput holds the Terraform output values from the nodepool manifest.
type HyperfleetNodePoolOutput struct {
	NodePoolID   string `json:"nodepool_id,omitempty"`
	NodePoolName string `json:"nodepool_name,omitempty"`
	Phase        string `json:"nodepool_phase,omitempty"`
	Replicas     *int   `json:"nodepool_replicas,omitempty"`
}

// HyperfleetNodePoolService manages the lifecycle of an rhcs_nodepool_hyperfleet
// Terraform resource via a TerraformExecutor.
type HyperfleetNodePoolService interface {
	Init() error
	Apply(args *HyperfleetNodePoolArgs) (string, error)
	Output() (*HyperfleetNodePoolOutput, error)
	Destroy() (string, error)
	WriteTFVars(args *HyperfleetNodePoolArgs) error
	DeleteTFVars() error
}

type hyperfleetNodePoolService struct {
	tfExecutor TerraformExecutor
}

// NewHyperfleetNodePoolService creates a HyperfleetNodePoolService rooted at the
// hyperfleet nodepool tf-manifests directory.
func NewHyperfleetNodePoolService(tfWorkspace string) (HyperfleetNodePoolService, error) {
	svc := &hyperfleetNodePoolService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetHyperfleetNodePoolManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *hyperfleetNodePoolService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *hyperfleetNodePoolService) Apply(args *HyperfleetNodePoolArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *hyperfleetNodePoolService) Output() (*HyperfleetNodePoolOutput, error) {
	out := &HyperfleetNodePoolOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *hyperfleetNodePoolService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *hyperfleetNodePoolService) WriteTFVars(args *HyperfleetNodePoolArgs) error {
	return svc.tfExecutor.WriteTerraformVars(args)
}

func (svc *hyperfleetNodePoolService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
