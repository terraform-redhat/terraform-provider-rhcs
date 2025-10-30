package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type BreakGlassCredentialArgs struct {
	Cluster            *string `hcl:"cluster"`
	Username           *string `hcl:"username"`
	ExpirationDuration *string `hcl:"expiration_duration"`
}

type BreakGlassCredential struct {
	Cluster             string `json:"cluster,omitempty"`
	ID                  string `json:"id,omitempty"`
	Username            string `json:"username,omitempty"`
	ExpirationDuration  string `json:"expiration_duration,omitempty"`
	ExpirationTimestamp string `json:"expiration_timestamp,omitempty"`
	RevocationTimestamp string `json:"revocation_timestamp,omitempty"`
	Status              string `json:"status,omitempty"`
	Kubeconfig          string `json:"kubeconfig,omitempty"`
}
type BreakGlassCredentialService interface {
	Init() error
	Plan(args *BreakGlassCredentialArgs) (string, error)
	Apply(args *BreakGlassCredentialArgs) (string, error)
	Output() (*BreakGlassCredential, error)
	Destroy() (string, error)

	ReadTFVars() (*BreakGlassCredentialArgs, error)
	WriteTFVars(args *BreakGlassCredentialArgs) error
	DeleteTFVars() error
}

type breakGlassCredentialService struct {
	tfExecutor TerraformExecutor
}

func NewBreakGlassCredentialService(tfWorkspace string, clusterType constants.ClusterType) (BreakGlassCredentialService, error) {
	svc := &breakGlassCredentialService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetBreakGlassCredentialManifestsDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *breakGlassCredentialService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *breakGlassCredentialService) Plan(args *BreakGlassCredentialArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *breakGlassCredentialService) Apply(args *BreakGlassCredentialArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *breakGlassCredentialService) Output() (*BreakGlassCredential, error) {
	output := &BreakGlassCredential{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (svc *breakGlassCredentialService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *breakGlassCredentialService) ReadTFVars() (*BreakGlassCredentialArgs, error) {
	args := &BreakGlassCredentialArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *breakGlassCredentialService) WriteTFVars(args *BreakGlassCredentialArgs) error {
	err := svc.tfExecutor.WriteTerraformVars(args)
	return err
}

func (svc *breakGlassCredentialService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
