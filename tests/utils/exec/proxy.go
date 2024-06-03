package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type ProxyArgs struct {
	ProxyCount          int    `json:"proxy_count,omitempty"`
	Region              string `json:"aws_region,omitempty"`
	VPCID               string `json:"vpc_id,omitempty"`
	PublicSubnetID      string `json:"subnet_public_id,omitempty"`
	TrustBundleFilePath string `json:"trust_bundle_path,omitempty"`
	KeyPairID           string `json:"key_pair_id,omitempty"`
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

type ProxiesOutput struct {
	Proxies []*ProxyOutput
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

func (proxy *ProxyService) Output() (output *ProxiesOutput, err error) {
	idpDir := CON.IDPsDir
	if proxy.ManifestDir != "" {
		idpDir = proxy.ManifestDir
	}
	var out map[string]interface{}
	out, err = runTerraformOutput(context.TODO(), idpDir)
	if err != nil {
		return
	}
	httpProxies := h.DigArrayToString(out["http_proxies"], "value")
	httpsProxies := h.DigArrayToString(out["https_proxies"], "value")
	noProxies := h.DigArrayToString(out["no_proxies"], "value")
	additionalTrustBundles := h.DigArrayToString(out["additional_trust_bundles"], "value")

	var proxies []*ProxyOutput
	for index, _ := range httpProxies {
		proxy := &ProxyOutput{
			HttpProxy:             httpProxies[index],
			HttpsProxy:            httpsProxies[index],
			NoProxy:               noProxies[index],
			AdditionalTrustBundle: additionalTrustBundles[index],
		}
		proxies = append(proxies, proxy)
	}
	output = &ProxiesOutput{Proxies: proxies}
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
