package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type KMSArgs struct {
	KMSName           string `json:"kms_name,omitempty"`
	AWSRegion         string `json:"aws_region,omitempty"`
	AccountRolePrefix string `json:"account_role_prefix,omitempty"`
	AccountRolePath   string `json:"path,omitempty"`
	TagKey            string `json:"tag_key,omitempty"`
	TagValue          string `json:"tag_value,omitempty"`
	TagDescription    string `json:"tag_description,omitempty"`
	HCP               bool   `json:"hcp,omitempty"`
}

type KMSOutput struct {
	KeyARN string `json:"arn,omitempty"`
}

type KMSService struct {
	CreationArgs *KMSArgs
	ManifestDir  string
	Context      context.Context
}

func (kms *KMSService) Init(manifestDirs ...string) error {
	kms.ManifestDir = CON.KMSDir
	if len(manifestDirs) != 0 {
		kms.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	kms.Context = ctx
	err := runTerraformInit(ctx, kms.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (kms *KMSService) Apply(createArgs *KMSArgs, recordtfvars bool, extraArgs ...string) error {
	kms.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(kms.Context, kms.ManifestDir, args)
	if err != nil {
		return err
	}
	if recordtfvars {
		recordTFvarsFile(kms.ManifestDir, tfvars)
	}

	return nil
}

func (kms *KMSService) Output() (KMSOutput, error) {
	kmsDir := CON.KMSDir
	if kms.ManifestDir != "" {
		kmsDir = kms.ManifestDir
	}
	var kmsOutput KMSOutput
	out, err := runTerraformOutput(context.TODO(), kmsDir)
	if err != nil {
		return kmsOutput, err
	}
	kmsOutput = KMSOutput{
		KeyARN: h.DigString(out["arn"], "value"),
	}

	return kmsOutput, nil
}

func (kms *KMSService) Destroy(createArgs ...*KMSArgs) error {
	if kms.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := kms.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroyWithArgs(kms.Context, kms.ManifestDir, args)

	return err
}

func NewKMSService(manifestDir ...string) (*KMSService, error) {
	kms := &KMSService{}
	err := kms.Init(manifestDir...)
	return kms, err
}
