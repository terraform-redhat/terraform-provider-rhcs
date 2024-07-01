package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type DnsDomainArgs struct {
	ID *string `hcl:"id"`
}

type DnsDomainOutput struct {
	DnsDomainId string `json:"dns_domain_id,omitempty"`
}

type DnsDomainService interface {
	Init() error
	Plan(args *DnsDomainArgs) (string, error)
	Apply(args *DnsDomainArgs) (string, error)
	Output() (*DnsDomainOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*DnsDomainArgs, error)
	DeleteTFVars() error
}

type dnsDomainService struct {
	tfExecutor TerraformExecutor
}

func NewDnsDomainService(tfWorkspace string, clusterType constants.ClusterType) (DnsDomainService, error) {
	svc := &dnsDomainService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetDnsDomainManifestsDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *dnsDomainService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *dnsDomainService) Plan(args *DnsDomainArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *dnsDomainService) Apply(args *DnsDomainArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *dnsDomainService) Output() (*DnsDomainOutput, error) {
	var output DnsDomainOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *dnsDomainService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *dnsDomainService) ReadTFVars() (*DnsDomainArgs, error) {
	args := &DnsDomainArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *dnsDomainService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
