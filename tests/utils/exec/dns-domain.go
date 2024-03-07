package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type DnsDomainArgs struct {
	ID string `json:"id,omitempty"`
}

type DnsDomainOutput struct {
	DnsDomainId string `json:"dns_domain_id,omitempty"`
}

type DnsService struct {
	CreationArgs *DnsDomainArgs
	ManifestDir  string
	Context      context.Context
}

func (dns *DnsService) Init(manifestDirs ...string) error {
	dns.ManifestDir = CON.DNSDir
	if len(manifestDirs) != 0 {
		dns.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	dns.Context = ctx
	err := runTerraformInit(ctx, dns.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (dns *DnsService) Create(createArgs *DnsDomainArgs, extraArgs ...string) error {
	dns.CreationArgs = createArgs
	args, _ := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(dns.Context, dns.ManifestDir, args...)
	if err != nil {
		return err
	}
	return nil
}

func (dns *DnsService) Output() (DnsDomainOutput, error) {
	dnsDir := CON.DNSDir
	if dns.ManifestDir != "" {
		dnsDir = dns.ManifestDir
	}
	var dnsOut DnsDomainOutput
	out, err := runTerraformOutput(context.TODO(), dnsDir)
	if err != nil {
		return dnsOut, err
	}
	dnsOut = DnsDomainOutput{
		DnsDomainId: h.DigString(out["dns_domain_id"], "value"),
	}

	return dnsOut, nil
}

func (dns *DnsService) Destroy(createArgs ...*DnsDomainArgs) error {
	if dns.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := dns.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(dns.Context, dns.ManifestDir, args...)

	return err
}

func NewDnsDomainService(manifestDir ...string) (*DnsService, error) {
	dns := &DnsService{}
	err := dns.Init(manifestDir...)
	return dns, err
}
