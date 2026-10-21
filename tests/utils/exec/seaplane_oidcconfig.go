// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// SeaplaneOidcConfigArgs contains the Terraform variable values for the
// rhcs_oidcconfig_seaplane manifest (managed OIDC mode).
type SeaplaneOidcConfigArgs struct {
	SeaplaneURL *string `hcl:"seaplane_url"`
	Name        *string `hcl:"name"`
	Type        *string `hcl:"type"`
}

// SeaplaneOidcConfigOutput holds the Terraform output values from the
// Seaplane OidcConfig manifest.
type SeaplaneOidcConfigOutput struct {
	OidcConfigID string `json:"oidc_config_id,omitempty"`
	Name         string `json:"name,omitempty"`
	IssuerURL    string `json:"issuer_url,omitempty"`
	Phase        string `json:"oidc_config_phase,omitempty"`
	Thumbprint   string `json:"thumbprint,omitempty"`
}

// SeaplaneOidcConfigService manages the lifecycle of an rhcs_oidcconfig_seaplane
// Terraform resource via a TerraformExecutor.
type SeaplaneOidcConfigService interface {
	Init() error
	Apply(args *SeaplaneOidcConfigArgs) (string, error)
	Output() (*SeaplaneOidcConfigOutput, error)
	Destroy() (string, error)
	WriteTFVars(args *SeaplaneOidcConfigArgs) error
	DeleteTFVars() error
}

type seaplaneOidcConfigService struct {
	tfExecutor TerraformExecutor
}

// NewSeaplaneOidcConfigService creates a SeaplaneOidcConfigService rooted at the
// Seaplane OidcConfig tf-manifests directory. tfWorkspace is used to isolate
// Terraform state between parallel test runs.
func NewSeaplaneOidcConfigService(tfWorkspace string) (SeaplaneOidcConfigService, error) {
	svc := &seaplaneOidcConfigService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetSeaplaneOidcConfigManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *seaplaneOidcConfigService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *seaplaneOidcConfigService) Apply(args *SeaplaneOidcConfigArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *seaplaneOidcConfigService) Output() (*SeaplaneOidcConfigOutput, error) {
	out := &SeaplaneOidcConfigOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *seaplaneOidcConfigService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *seaplaneOidcConfigService) WriteTFVars(args *SeaplaneOidcConfigArgs) error {
	return svc.tfExecutor.WriteTerraformVars(args)
}

func (svc *seaplaneOidcConfigService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
