// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// HyperfleetOidcConfigArgs contains the Terraform variable values for the
// rhcs_oidcconfig_hyperfleet manifest (managed OIDC mode).
type HyperfleetOidcConfigArgs struct {
	HyperfleetURL *string `hcl:"hyperfleet_url"`
	Name          *string `hcl:"name"`
	Type          *string `hcl:"type"`
}

// HyperfleetOidcConfigOutput holds the Terraform output values from the
// hyperfleet OidcConfig manifest.
type HyperfleetOidcConfigOutput struct {
	OidcConfigID string `json:"oidc_config_id,omitempty"`
	Name         string `json:"name,omitempty"`
	IssuerURL    string `json:"issuer_url,omitempty"`
	Phase        string `json:"oidc_config_phase,omitempty"`
	Thumbprint   string `json:"thumbprint,omitempty"`
}

// HyperfleetOidcConfigService manages the lifecycle of an rhcs_oidcconfig_hyperfleet
// Terraform resource via a TerraformExecutor.
type HyperfleetOidcConfigService interface {
	Init() error
	Apply(args *HyperfleetOidcConfigArgs) (string, error)
	Output() (*HyperfleetOidcConfigOutput, error)
	Destroy() (string, error)
	WriteTFVars(args *HyperfleetOidcConfigArgs) error
	DeleteTFVars() error
}

type hyperfleetOidcConfigService struct {
	tfExecutor TerraformExecutor
}

// NewHyperfleetOidcConfigService creates a HyperfleetOidcConfigService rooted at the
// hyperfleet OidcConfig tf-manifests directory. tfWorkspace is used to isolate
// Terraform state between parallel test runs.
func NewHyperfleetOidcConfigService(tfWorkspace string) (HyperfleetOidcConfigService, error) {
	svc := &hyperfleetOidcConfigService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetHyperfleetOidcConfigManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *hyperfleetOidcConfigService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *hyperfleetOidcConfigService) Apply(args *HyperfleetOidcConfigArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *hyperfleetOidcConfigService) Output() (*HyperfleetOidcConfigOutput, error) {
	out := &HyperfleetOidcConfigOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *hyperfleetOidcConfigService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *hyperfleetOidcConfigService) WriteTFVars(args *HyperfleetOidcConfigArgs) error {
	return svc.tfExecutor.WriteTerraformVars(args)
}

func (svc *hyperfleetOidcConfigService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
