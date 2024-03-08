package exec

import (
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
	tfExec      TerraformExec
	ManifestDir string
}

func newKMSService(clusterType CON.ClusterType, tfExec TerraformExec) (*KMSService, error) {
	kms := &KMSService{
		ManifestDir: GetKmsManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := kms.Init()
	return kms, err
}

func (kms *KMSService) Init() error {
	return kms.tfExec.RunTerraformInit(kms.ManifestDir)

}

func (kms *KMSService) Apply(createArgs *KMSArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = kms.tfExec.RunTerraformApply(kms.ManifestDir, tfVars)
	return
}

func (kms *KMSService) Output() (*KMSOutput, error) {
	out, err := kms.tfExec.RunTerraformOutput(kms.ManifestDir)
	if err != nil {
		return nil, err
	}
	kmsOutput := &KMSOutput{
		KeyARN: h.DigString(out["arn"], "value"),
	}

	return kmsOutput, nil
}

func (kms *KMSService) Destroy(deleteTFVars bool) (string, error) {
	return kms.tfExec.RunTerraformDestroy(kms.ManifestDir, deleteTFVars)
}
