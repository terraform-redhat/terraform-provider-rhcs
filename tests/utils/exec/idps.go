package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type IDPArgs struct {
	ClusterID      *string           `hcl:"cluster_id"`
	Name           *string           `hcl:"name"`
	ID             *string           `hcl:"id"`
	CA             *string           `hcl:"ca"`
	LDAPAttributes *LDAPAttributes   `hcl:"attributes"`
	ClientID       *string           `hcl:"client_id"`
	ClientSecret   *string           `hcl:"client_secret"`
	Organizations  *[]string         `hcl:"organizations"`
	HostedDomain   *string           `hcl:"hosted_domain"`
	Insecure       *bool             `hcl:"insecure"`
	MappingMethod  *string           `hcl:"mapping_method"`
	HtpasswdUsers  *[]HTPasswordUser `hcl:"htpasswd_users"`
	URL            *string           `hcl:"idp_url"`
}

type HTPasswordUser struct {
	Username *string `cty:"username"`
	Password *string `cty:"password"`
}

type LDAPAttributes struct {
	Emails             *[]string `cty:"email"`
	IDs                *[]string `cty:"id"`
	Names              *[]string `cty:"name"`
	PreferredUsernames *[]string `cty:"preferred_username"`
}

// for now holds only ID, additional vars might be needed in the future
type IDPOutput struct {
	ID       string `json:"idp_id,omitempty"`
	GoogleID string `json:"idp_google_id,omitempty"`
	LDAPID   string `json:"idp_ldap_id,omitempty"`
}

type IDPService interface {
	Init() error
	Plan(args *IDPArgs) (string, error)
	Apply(args *IDPArgs) (string, error)
	Output() (*IDPOutput, error)
	Destroy() (string, error)

	GetStateResource(resourceType string, resoureName string) (interface{}, error)

	ReadTFVars() (*IDPArgs, error)
	DeleteTFVars() error
}

type idpService struct {
	tfExecutor TerraformExecutor
}

func NewIDPService(tfWorkspace string, clusterType constants.ClusterType, idpType constants.IDPType) (IDPService, error) {
	svc := &idpService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetIDPManifestsDir(clusterType, idpType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *idpService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *idpService) Plan(args *IDPArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *idpService) Apply(args *IDPArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *idpService) Output() (*IDPOutput, error) {
	var output IDPOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *idpService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *idpService) GetStateResource(resourceType string, resoureName string) (interface{}, error) {
	return svc.tfExecutor.GetStateResource(resourceType, resoureName)
}

func (svc *idpService) ReadTFVars() (*IDPArgs, error) {
	args := &IDPArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *idpService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
