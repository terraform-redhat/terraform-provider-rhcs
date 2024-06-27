package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type ProxyArgs struct {
	ProxyCount          *int    `hcl:"proxy_count"`
	Region              *string `hcl:"aws_region"`
	VPCID               *string `hcl:"vpc_id"`
	PublicSubnetID      *string `hcl:"subnet_public_id"`
	TrustBundleFilePath *string `hcl:"trust_bundle_path"`
	KeyPairID           *string `hcl:"key_pair_id"`
}

// for now holds only ID, additional vars might be needed in the future
type ProxyServiceOutput struct {
	HttpProxies            []string `json:"http_proxies,omitempty"`
	HttpsProxies           []string `json:"https_proxies,omitempty"`
	NoProxies              []string `json:"no_proxies,omitempty"`
	AdditionalTrustBundles []string `json:"additional_trust_bundles,omitempty"`
}

type ProxyOutput struct {
	HttpProxy             string
	HttpsProxy            string
	NoProxy               string
	AdditionalTrustBundle string
}

type ProxiesOutput struct {
	Proxies []*ProxyOutput
}

type ProxyService interface {
	Init() error
	Plan(args *ProxyArgs) (string, error)
	Apply(args *ProxyArgs) (string, error)
	Output() (*ProxiesOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*ProxyArgs, error)
	DeleteTFVars() error
}

type proxyService struct {
	tfExecutor TerraformExecutor
}

func NewProxyService(manifestsDirs ...string) (ProxyService, error) {
	manifestsDir := constants.ProxyDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &proxyService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *proxyService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *proxyService) Plan(args *ProxyArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *proxyService) Apply(args *ProxyArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *proxyService) Output() (*ProxiesOutput, error) {
	var svcOutput ProxyServiceOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&svcOutput)
	if err != nil {
		return nil, err
	}

	var proxies []*ProxyOutput
	for index, _ := range svcOutput.HttpProxies {
		proxies = append(proxies, &ProxyOutput{
			HttpProxy:             svcOutput.HttpProxies[index],
			HttpsProxy:            svcOutput.HttpsProxies[index],
			NoProxy:               svcOutput.NoProxies[index],
			AdditionalTrustBundle: svcOutput.AdditionalTrustBundles[index],
		})
	}

	return &ProxiesOutput{
		Proxies: proxies,
	}, nil
}

func (svc *proxyService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *proxyService) ReadTFVars() (*ProxyArgs, error) {
	args := &ProxyArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *proxyService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
