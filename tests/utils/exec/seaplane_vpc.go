// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// SeaplaneVPCArgs contains the Terraform variable values for the Seaplane
// VPC manifest (VPC + subnets + NAT + Route53 private hosted zone).
type SeaplaneVPCArgs struct {
	AWSRegion        *string `hcl:"aws_region"`
	NamePrefix       *string `hcl:"name_prefix"`
	VPCCIDR          *string `hcl:"vpc_cidr"`
	AvailabilityZone *string `hcl:"availability_zone"`
}

// SeaplaneVPCOutput holds the Terraform output values from the Seaplane
// VPC manifest.
type SeaplaneVPCOutput struct {
	VPCID            string `json:"vpc_id,omitempty"`
	PrivateSubnetID  string `json:"private_subnet_id,omitempty"`
	PublicSubnetID   string `json:"public_subnet_id,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
	HostedZoneID     string `json:"hosted_zone_id,omitempty"`
}

// SeaplaneVPCService manages the VPC + networking Terraform resources needed
// by a Seaplane cluster.
type SeaplaneVPCService interface {
	Init() error
	Apply(args *SeaplaneVPCArgs) (string, error)
	Output() (*SeaplaneVPCOutput, error)
	Destroy() (string, error)
}

type seaplaneVPCService struct {
	tfExecutor TerraformExecutor
}

// NewSeaplaneVPCService creates a SeaplaneVPCService rooted at the
// Seaplane VPC tf-manifests directory.
func NewSeaplaneVPCService(tfWorkspace string) (SeaplaneVPCService, error) {
	svc := &seaplaneVPCService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetSeaplaneVPCManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *seaplaneVPCService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *seaplaneVPCService) Apply(args *SeaplaneVPCArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *seaplaneVPCService) Output() (*SeaplaneVPCOutput, error) {
	out := &SeaplaneVPCOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *seaplaneVPCService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}
