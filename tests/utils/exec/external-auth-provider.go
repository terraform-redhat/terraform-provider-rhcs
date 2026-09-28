// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type ExternalAuthProviderArgs struct {
	Cluster                   *string   `hcl:"cluster"`
	Name                      *string   `hcl:"name"`
	IssuerURL                 *string   `hcl:"issuer_url"`
	IssuerAudiences           *[]string `hcl:"issuer_audiences"`
	ClaimMappingGroupsClaim   *string   `hcl:"claim_mapping_groups_claim"`
	ClaimMappingUsernameClaim *string   `hcl:"claim_mapping_username_claim"`
	ClaimValidationRule       *[]string `hcl:"claim_validation_rule"`
}

type ExternalAuthProviderOutput struct {
	Cluster                   string   `json:"cluster,omitempty"`
	ID                        string   `json:"id,omitempty"`
	Name                      string   `json:"name,omitempty"`
	IssuerURL                 string   `json:"issuer_url,omitempty"`
	IssuerAudiences           []string `json:"issuer_audiences,omitempty"`
	ClaimMappingGroupsClaim   string   `json:"claim_mapping_groups_claim,omitempty"`
	ClaimMappingUsernameClaim string   `json:"claim_mapping_username_claim,omitempty"`
	ClaimValidationRule       []string `json:"claim_validation_rule,omitempty"`
}

type ExternalAuthProviderService interface {
	Init() error
	Apply(args *ExternalAuthProviderArgs) (string, error)
	Output() (*ExternalAuthProviderOutput, error)
	Destroy() (string, error)
}

type externalAuthProviderService struct {
	tfExecutor TerraformExecutor
}

func NewExternalAuthProviderService(
	tfWorkspace string, clusterType constants.ClusterType,
) (ExternalAuthProviderService, error) {
	svc := &externalAuthProviderService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetExternalAuthProviderManifestsDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *externalAuthProviderService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *externalAuthProviderService) Apply(args *ExternalAuthProviderArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *externalAuthProviderService) Output() (*ExternalAuthProviderOutput, error) {
	output := &ExternalAuthProviderOutput{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (svc *externalAuthProviderService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}
