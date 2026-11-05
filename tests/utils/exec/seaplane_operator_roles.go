// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

// SeaplaneOperatorRolesArgs contains the Terraform variable values for the
// rhcs_seaplane_operator_roles manifest.
type SeaplaneOperatorRolesArgs struct {
	SeaplaneURL         *string `hcl:"seaplane_url"`
	OperatorRolesPrefix *string `hcl:"operator_roles_prefix"`
	OIDCIssuerURL       *string `hcl:"oidc_issuer_url"`
}

// SeaplaneOperatorRolesOutput holds the Terraform output values from the
// rhcs_seaplane_operator_roles manifest.
type SeaplaneOperatorRolesOutput struct {
	DefinitionVersion   string `json:"definition_version,omitempty"`
	InstanceProfileName string `json:"instance_profile_name,omitempty"`
	WorkerRoleARN       string `json:"worker_role_arn,omitempty"`
	RoleCount           int    `json:"role_count,omitempty"`
}

// SeaplaneOperatorRolesService manages the rhcs_seaplane_operator_roles
// Terraform resource via a TerraformExecutor.
type SeaplaneOperatorRolesService interface {
	Init() error
	Apply(args *SeaplaneOperatorRolesArgs) (string, error)
	Plan(args *SeaplaneOperatorRolesArgs) (string, error)
	Output() (*SeaplaneOperatorRolesOutput, error)
	Import(args *SeaplaneOperatorRolesArgs, prefix string) (string, error)
	Destroy() (string, error)
}

type seaplaneOperatorRolesService struct {
	tfExecutor TerraformExecutor
}

// NewSeaplaneOperatorRolesService creates a SeaplaneOperatorRolesService
// rooted at the Seaplane operator roles tf-manifests directory. tfWorkspace
// is used to isolate Terraform state between parallel test runs.
func NewSeaplaneOperatorRolesService(tfWorkspace string) (SeaplaneOperatorRolesService, error) {
	svc := &seaplaneOperatorRolesService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetSeaplaneOperatorRolesManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *seaplaneOperatorRolesService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *seaplaneOperatorRolesService) Apply(args *SeaplaneOperatorRolesArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *seaplaneOperatorRolesService) Plan(args *SeaplaneOperatorRolesArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *seaplaneOperatorRolesService) Output() (*SeaplaneOperatorRolesOutput, error) {
	out := &SeaplaneOperatorRolesOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Import imports an existing role set (by operator-roles prefix) into the
// Terraform state of this service's workspace.
func (svc *seaplaneOperatorRolesService) Import(args *SeaplaneOperatorRolesArgs, prefix string) (string, error) {
	if err := svc.tfExecutor.WriteTerraformVars(args); err != nil {
		return "", err
	}
	return svc.tfExecutor.RunTerraformImport("rhcs_seaplane_operator_roles.roles", prefix)
}

func (svc *seaplaneOperatorRolesService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

// SeaplaneOperatorRolesDiscoveryArgs contains the Terraform variable values
// for the rhcs_seaplane_operator_roles discovery data source manifest.
type SeaplaneOperatorRolesDiscoveryArgs struct {
	OperatorRolesPrefix *string `hcl:"operator_roles_prefix"`
}

// SeaplaneOperatorRolesDiscoveryOutput holds the Terraform output values from
// the discovery data source manifest.
type SeaplaneOperatorRolesDiscoveryOutput struct {
	RoleCount         int    `json:"role_count,omitempty"`
	DefinitionVersion string `json:"definition_version,omitempty"`
}

// SeaplaneOperatorRolesDiscoveryService manages the
// rhcs_seaplane_operator_roles data source via a TerraformExecutor.
type SeaplaneOperatorRolesDiscoveryService interface {
	Init() error
	Apply(args *SeaplaneOperatorRolesDiscoveryArgs) (string, error)
	Output() (*SeaplaneOperatorRolesDiscoveryOutput, error)
	Destroy() (string, error)
}

type seaplaneOperatorRolesDiscoveryService struct {
	tfExecutor TerraformExecutor
}

// NewSeaplaneOperatorRolesDiscoveryService creates a
// SeaplaneOperatorRolesDiscoveryService rooted at the Seaplane operator roles
// discovery tf-manifests directory.
func NewSeaplaneOperatorRolesDiscoveryService(tfWorkspace string) (SeaplaneOperatorRolesDiscoveryService, error) {
	svc := &seaplaneOperatorRolesDiscoveryService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetSeaplaneOperatorRolesDiscoveryManifestsDir()),
	}
	return svc, svc.Init()
}

func (svc *seaplaneOperatorRolesDiscoveryService) Init() error {
	_, err := svc.tfExecutor.RunTerraformInit()
	return err
}

func (svc *seaplaneOperatorRolesDiscoveryService) Apply(args *SeaplaneOperatorRolesDiscoveryArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *seaplaneOperatorRolesDiscoveryService) Output() (*SeaplaneOperatorRolesDiscoveryOutput, error) {
	out := &SeaplaneOperatorRolesDiscoveryOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (svc *seaplaneOperatorRolesDiscoveryService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}
