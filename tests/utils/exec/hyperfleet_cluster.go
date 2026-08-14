// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// HyperfleetClusterArgs contains the Terraform variable values for the
// rhcs_cluster_hyperfleet manifest.
type HyperfleetClusterArgs struct {
	HyperfleetURL       *string `hcl:"hyperfleet_url"`
	AWSRegion           *string `hcl:"aws_region"`
	ClusterName         *string `hcl:"cluster_name"`
	OperatorRolesPrefix *string `hcl:"operator_roles_prefix"`
	SubnetID            *string `hcl:"subnet_id"`
	VPCID               *string `hcl:"vpc_id"`
	AvailabilityZone    *string `hcl:"availability_zone"`
	ExpirationTimestamp *string `hcl:"expiration_timestamp"`
}

// HyperfleetClusterOutput holds the Terraform output values from the hyperfleet
// cluster manifest.
type HyperfleetClusterOutput struct {
	ClusterID   string `json:"cluster_id,omitempty"`
	ClusterName string `json:"cluster_name,omitempty"`
	Phase       string `json:"cluster_phase,omitempty"`
	APIURL      string `json:"cluster_api_url,omitempty"`
	OIDCIssuer  string `json:"oidc_issuer,omitempty"`
}

// HyperfleetClusterService manages the lifecycle of an rhcs_cluster_hyperfleet
// Terraform resource via a TerraformExecutor.
type HyperfleetClusterService interface {
	Init() error
	Apply(args *HyperfleetClusterArgs) (string, error)
	Output() (*HyperfleetClusterOutput, error)
	Destroy() (string, error)
	WriteTFVars(args *HyperfleetClusterArgs) error
	DeleteTFVars() error
}

type hyperfleetClusterService struct {
	tfExecutor TerraformExecutor
}

// NewHyperfleetClusterService creates a HyperfleetClusterService rooted at the
// hyperfleet cluster tf-manifests directory. tfWorkspace is used to isolate
// Terraform state between parallel test runs.
func NewHyperfleetClusterService(tfWorkspace string) (HyperfleetClusterService, error) {
	svc := &hyperfleetClusterService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetHyperfleetClusterManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *hyperfleetClusterService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *hyperfleetClusterService) Apply(args *HyperfleetClusterArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *hyperfleetClusterService) Output() (*HyperfleetClusterOutput, error) {
	out := &HyperfleetClusterOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *hyperfleetClusterService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *hyperfleetClusterService) WriteTFVars(args *HyperfleetClusterArgs) error {
	return svc.tfExecutor.WriteTerraformVars(args)
}

func (svc *hyperfleetClusterService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
