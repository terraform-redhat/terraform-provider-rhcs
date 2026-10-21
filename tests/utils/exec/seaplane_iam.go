// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// SeaplaneIAMArgs contains the Terraform variable values for the Seaplane
// IAM roles manifest (OIDC provider + operator roles + worker role).
// Apply AFTER the cluster resource so that OIDCIssuerURL is known.
type SeaplaneIAMArgs struct {
	AWSRegion           *string `hcl:"aws_region"`
	OperatorRolesPrefix *string `hcl:"operator_roles_prefix"`
	OIDCIssuerURL       *string `hcl:"oidc_issuer_url"`
	OIDCThumbprint      *string `hcl:"oidc_thumbprint"`
}

// SeaplaneIAMOutput holds the Terraform output values from the Seaplane
// IAM manifest.
type SeaplaneIAMOutput struct {
	OIDCProviderARN           string `json:"oidc_provider_arn,omitempty"`
	WorkerRoleARN             string `json:"worker_role_arn,omitempty"`
	WorkerInstanceProfileName string `json:"worker_instance_profile_name,omitempty"`
}

// SeaplaneIAMService manages the OIDC provider and IAM roles Terraform
// resources needed by a Seaplane cluster.
type SeaplaneIAMService interface {
	Init() error
	Apply(args *SeaplaneIAMArgs) (string, error)
	Output() (*SeaplaneIAMOutput, error)
	Destroy() (string, error)
}

type seaplaneIAMService struct {
	tfExecutor TerraformExecutor
}

// NewSeaplaneIAMService creates a SeaplaneIAMService rooted at the
// Seaplane IAM tf-manifests directory.
func NewSeaplaneIAMService(tfWorkspace string) (SeaplaneIAMService, error) {
	svc := &seaplaneIAMService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetSeaplaneIAMManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *seaplaneIAMService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *seaplaneIAMService) Apply(args *SeaplaneIAMArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *seaplaneIAMService) Output() (*SeaplaneIAMOutput, error) {
	out := &SeaplaneIAMOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *seaplaneIAMService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}
