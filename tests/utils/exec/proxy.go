package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type ProxyArgs struct {
	Region              string `json:"aws_region,omitempty"`
	VPCID               string `json:"vpc_id,omitempty"`
	PublicSubnetID      string `json:"subnet_public_id,omitempty"`
	TrustBundleFilePath string `json:"trust_bundle_path,omitempty"`
	KeyPairID           string `json:"key_pair_id",omitempty`
}

type ProxyService struct {
	CreationArgs *ProxyArgs
	ManifestDir  string
	Context      context.Context
}

// for now holds only ID, additional vars might be needed in the future
type ProxyOutput struct {
	HttpProxy             string `json:"http_proxy,omitempty"`
	HttpsProxy            string `json:"https_proxy,omitempty"`
	NoProxy               string `json:"no_proxy,omitempty"`
	AdditionalTrustBundle string `json:"additional_trust_bundle,omitempty"`
}

func (proxy *ProxyService) Init(manifestDirs ...string) error {
	proxy.ManifestDir = CON.ProxyDir
	if len(manifestDirs) != 0 {
		proxy.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	proxy.Context = ctx
	err := runTerraformInit(ctx, proxy.ManifestDir)
	if err != nil {
		return err
	}
	return nil
}

func (proxy *ProxyService) Apply(createArgs *ProxyArgs, recordtfvars bool, extraArgs ...string) error {
	proxy.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(proxy.Context, proxy.ManifestDir, args...)
	if err != nil {
		return err
	}
	if recordtfvars {
		recordTFvarsFile(proxy.ManifestDir, tfvars)
	}

	return nil
}

func (proxy *ProxyService) Output() (output *ProxyOutput, err error) {
	idpDir := CON.IDPsDir
	if proxy.ManifestDir != "" {
		idpDir = proxy.ManifestDir
	}
	var out map[string]interface{}
	out, err = runTerraformOutput(context.TODO(), idpDir)
	if err != nil {
		return
	}
	httpProxy := h.DigString(out["http_proxy"], "value")
	httpsProxy := h.DigString(out["https_proxy"], "value")
	noProxy := h.DigString(out["no_proxy"], "value")
	additionalTrustBundle := h.DigString(out["additional_trust_bundle"], "value")

	output = &ProxyOutput{
		HttpProxy:             httpProxy,
		HttpsProxy:            httpsProxy,
		NoProxy:               noProxy,
		AdditionalTrustBundle: additionalTrustBundle,
	}
	return
}

func (proxy *ProxyService) Destroy(createArgs ...*ProxyArgs) error {
	if proxy.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := proxy.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(proxy.Context, proxy.ManifestDir, args...)

	return err
}

func NewProxyService(manifestDir ...string) (*ProxyService, error) {
	proxy := &ProxyService{}
	err := proxy.Init(manifestDir...)
	return proxy, err
}
