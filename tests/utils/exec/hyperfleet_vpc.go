// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// HyperfleetVPCArgs contains the Terraform variable values for the hyperfleet
// VPC manifest (VPC + subnets + NAT + Route53 private hosted zone).
type HyperfleetVPCArgs struct {
	AWSRegion        *string `hcl:"aws_region"`
	NamePrefix       *string `hcl:"name_prefix"`
	VPCCIDR          *string `hcl:"vpc_cidr"`
	AvailabilityZone *string `hcl:"availability_zone"`
}

// HyperfleetVPCOutput holds the Terraform output values from the hyperfleet
// VPC manifest.
type HyperfleetVPCOutput struct {
	VPCID           string `json:"vpc_id,omitempty"`
	PrivateSubnetID string `json:"private_subnet_id,omitempty"`
	PublicSubnetID  string `json:"public_subnet_id,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
	HostedZoneID    string `json:"hosted_zone_id,omitempty"`
}

// HyperfleetVPCService manages the VPC + networking Terraform resources needed
// by a hyperfleet cluster.
type HyperfleetVPCService interface {
	Init() error
	Apply(args *HyperfleetVPCArgs) (string, error)
	Output() (*HyperfleetVPCOutput, error)
	Destroy() (string, error)
}

type hyperfleetVPCService struct {
	tfExecutor TerraformExecutor
}

// NewHyperfleetVPCService creates a HyperfleetVPCService rooted at the
// hyperfleet VPC tf-manifests directory.
func NewHyperfleetVPCService(tfWorkspace string) (HyperfleetVPCService, error) {
	svc := &hyperfleetVPCService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetHyperfleetVPCManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *hyperfleetVPCService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *hyperfleetVPCService) Apply(args *HyperfleetVPCArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *hyperfleetVPCService) Output() (*HyperfleetVPCOutput, error) {
	out := &HyperfleetVPCOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *hyperfleetVPCService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}
