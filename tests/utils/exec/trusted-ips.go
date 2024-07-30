package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type TrustedIPsArgs struct {
}

type TrustedIP struct {
	Enabled bool   `json:"enabled,omitempty"`
	Id      string `json:"id,omitempty"`
}

type TrustedIPList struct {
	Items []TrustedIP `json:"items,omitempty"`
}

type TrustedIPsOutput struct {
	// TrustedIPs map[string][]TrustedIP `json:"trusted_ips,omitempty"`
	TrustedIPs TrustedIPList `json:"trusted_ips,omitempty"`
}

type TrustedIPsService interface {
	Init() error
	Plan(args *TrustedIPsArgs) (string, error)
	Apply(args *TrustedIPsArgs) (string, error)
	Output() (*TrustedIPsOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*TrustedIPsArgs, error)
	DeleteTFVars() error
}

type trustedIPsService struct {
	tfExecutor TerraformExecutor
}

func NewTrustedIPsService(manifestsDirs ...string) (TrustedIPsService, error) {
	manifestsDir := constants.TrustedIPsDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &trustedIPsService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *trustedIPsService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *trustedIPsService) Plan(args *TrustedIPsArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *trustedIPsService) Apply(args *TrustedIPsArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *trustedIPsService) Output() (*TrustedIPsOutput, error) {
	var output TrustedIPsOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *trustedIPsService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *trustedIPsService) ReadTFVars() (*TrustedIPsArgs, error) {
	args := &TrustedIPsArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *trustedIPsService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
