// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// HyperfleetIAMArgs contains the Terraform variable values for the hyperfleet
// IAM roles manifest (OIDC provider + operator roles + worker role).
// Apply AFTER the cluster resource so that OIDCIssuerURL is known.
type HyperfleetIAMArgs struct {
	AWSRegion           *string `hcl:"aws_region"`
	OperatorRolesPrefix *string `hcl:"operator_roles_prefix"`
	OIDCIssuerURL       *string `hcl:"oidc_issuer_url"`
}

// HyperfleetIAMOutput holds the Terraform output values from the hyperfleet
// IAM manifest.
type HyperfleetIAMOutput struct {
	OIDCProviderARN           string `json:"oidc_provider_arn,omitempty"`
	WorkerRoleARN             string `json:"worker_role_arn,omitempty"`
	WorkerInstanceProfileName string `json:"worker_instance_profile_name,omitempty"`
}

// HyperfleetIAMService manages the OIDC provider and IAM roles Terraform
// resources needed by a hyperfleet cluster.
type HyperfleetIAMService interface {
	Init() error
	Apply(args *HyperfleetIAMArgs) (string, error)
	Output() (*HyperfleetIAMOutput, error)
	Destroy() (string, error)
}

type hyperfleetIAMService struct {
	tfExecutor TerraformExecutor
}

// NewHyperfleetIAMService creates a HyperfleetIAMService rooted at the
// hyperfleet IAM tf-manifests directory.
func NewHyperfleetIAMService(tfWorkspace string) (HyperfleetIAMService, error) {
	svc := &hyperfleetIAMService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetHyperfleetIAMManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *hyperfleetIAMService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *hyperfleetIAMService) Apply(args *HyperfleetIAMArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *hyperfleetIAMService) Output() (*HyperfleetIAMOutput, error) {
	out := &HyperfleetIAMOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *hyperfleetIAMService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}
