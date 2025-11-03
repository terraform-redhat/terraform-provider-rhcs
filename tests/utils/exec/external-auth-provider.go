package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type ExternalAuthProviderArgs struct {
	Cluster                      *string                        `hcl:"cluster"`
	ID                          *string                        `hcl:"id"`
	IssuerURL                   *string                        `hcl:"issuer_url"`
	IssuerAudiences             *[]string                      `hcl:"issuer_audiences"`
	IssuerCA                    *string                        `hcl:"issuer_ca"`
	ConsoleClientID             *string                        `hcl:"console_client_id"`
	ConsoleClientSecret         *string                        `hcl:"console_client_secret"`
	ClaimMappingUsernameKey     *string                        `hcl:"claim_mapping_username_key"`
	ClaimMappingGroupsKey       *string                        `hcl:"claim_mapping_groups_key"`
	ClaimValidationRules        *[]ClaimValidationRule         `hcl:"claim_validation_rules"`
}

type ClaimValidationRule struct {
	Claim         *string `cty:"claim"`
	RequiredValue *string `cty:"required_value"`
}

type ExternalAuthProviderOutput struct {
	ID      string `json:"id,omitempty"`
	Cluster string `json:"cluster,omitempty"`
}

type ExternalAuthProviderService interface {
	Init() error
	Plan(args *ExternalAuthProviderArgs) (string, error)
	Apply(args *ExternalAuthProviderArgs) (string, error)
	Output() (*ExternalAuthProviderOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*ExternalAuthProviderArgs, error)
	DeleteTFVars() error
}

type externalAuthProviderService struct {
	tfExecutor TerraformExecutor
}

func NewExternalAuthProviderService(tfWorkspace string, clusterType constants.ClusterType) (ExternalAuthProviderService, error) {
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

func (svc *externalAuthProviderService) Plan(args *ExternalAuthProviderArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *externalAuthProviderService) Apply(args *ExternalAuthProviderArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *externalAuthProviderService) Output() (*ExternalAuthProviderOutput, error) {
	var output ExternalAuthProviderOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *externalAuthProviderService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *externalAuthProviderService) ReadTFVars() (*ExternalAuthProviderArgs, error) {
	args := &ExternalAuthProviderArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *externalAuthProviderService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTFVars()
}

func NewClaimValidationRule(claim, requiredValue string) ClaimValidationRule {
	return ClaimValidationRule{
		Claim:         helper.StringPointer(claim),
		RequiredValue: helper.StringPointer(requiredValue),
	}
}