package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type ProxyArgs struct {
	Region              string `json:"aws_region,omitempty"`
	VPCID               string `json:"vpc_id,omitempty"`
	PublicSubnetID      string `json:"subnet_public_id,omitempty"`
	TrustBundleFilePath string `json:"trust_bundle_path,omitempty"`
}

// for now holds only ID, additional vars might be needed in the future
type ProxyOutput struct {
	HttpProxy             string `json:"http_proxy,omitempty"`
	HttpsProxy            string `json:"https_proxy,omitempty"`
	NoProxy               string `json:"no_proxy,omitempty"`
	AdditionalTrustBundle string `json:"additional_trust_bundle,omitempty"`
}

type ProxyService struct {
	tfExec      TerraformExec
	ManifestDir string
}

func newProxyService(clusterType CON.ClusterType, tfExec TerraformExec) (*ProxyService, error) {
	proxy := &ProxyService{
		ManifestDir: GetProxyManifestDir(clusterType),
	}
	err := proxy.Init()
	return proxy, err
}

func (proxy *ProxyService) Init() error {
	return proxy.tfExec.RunTerraformInit(proxy.ManifestDir)
}

func (proxy *ProxyService) Apply(createArgs *ProxyArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = proxy.tfExec.RunTerraformApply(proxy.ManifestDir, tfVars)
	return
}

func (proxy *ProxyService) Output() (*ProxyOutput, error) {
	out, err := proxy.tfExec.RunTerraformOutput(proxy.ManifestDir)
	if err != nil {
		return nil, err
	}
	output := &ProxyOutput{
		HttpProxy:             h.DigString(out["http_proxy"], "value"),
		HttpsProxy:            h.DigString(out["https_proxy"], "value"),
		NoProxy:               h.DigString(out["no_proxy"], "value"),
		AdditionalTrustBundle: h.DigString(out["additional_trust_bundle"], "value"),
	}
	return output, nil
}

func (proxy *ProxyService) Destroy(deleteTFVars bool) (string, error) {
	return proxy.tfExec.RunTerraformDestroy(proxy.ManifestDir, deleteTFVars)
}
