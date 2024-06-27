package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type KMSArgs struct {
	KMSName           *string `hcl:"kms_name"`
	AWSRegion         *string `hcl:"aws_region"`
	AccountRolePrefix *string `hcl:"account_role_prefix"`
	AccountRolePath   *string `hcl:"path"`
	TagKey            *string `hcl:"tag_key"`
	TagValue          *string `hcl:"tag_value"`
	TagDescription    *string `hcl:"tag_description"`
	HCP               *bool   `hcl:"hcp"`
}

type KMSOutput struct {
	KeyARN string `json:"arn,omitempty"`
}

type KMSService interface {
	Init() error
	Plan(args *KMSArgs) (string, error)
	Apply(args *KMSArgs) (string, error)
	Output() (*KMSOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*KMSArgs, error)
	DeleteTFVars() error
}

type kmsService struct {
	tfExecutor TerraformExecutor
}

func NewKMSService(manifestsDirs ...string) (KMSService, error) {
	manifestsDir := constants.KMSDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &kmsService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *kmsService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *kmsService) Plan(args *KMSArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *kmsService) Apply(args *KMSArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *kmsService) Output() (*KMSOutput, error) {
	var output KMSOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *kmsService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *kmsService) ReadTFVars() (*KMSArgs, error) {
	args := &KMSArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *kmsService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
