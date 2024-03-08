package exec

import (
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
	tfExec      TerraformExec
	ManifestDir string
}

func newDnsDomainService(clusterType CON.ClusterType, tfExec TerraformExec) (*DnsService, error) {
	dns := &DnsService{
		ManifestDir: GetDnsDomainManifestDir(clusterType),
		tfExec:      tfExec,
	}
	err := dns.Init()
	return dns, err
}

func (dns *DnsService) Init(manifestDirs ...string) error {
	return dns.tfExec.RunTerraformInit(dns.ManifestDir)

}

func (dns *DnsService) Apply(createArgs *DnsDomainArgs) (output string, err error) {
	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = dns.tfExec.RunTerraformApply(dns.ManifestDir, tfVars)
	return
}

func (dns *DnsService) Output() (*DnsDomainOutput, error) {
	out, err := dns.tfExec.RunTerraformOutput(dns.ManifestDir)
	if err != nil {
		return nil, err
	}
	dnsOut := &DnsDomainOutput{
		DnsDomainId: h.DigString(out["dns_domain_id"], "value"),
	}

	return dnsOut, err
}

func (dns *DnsService) Destroy(deleteTFVars bool) (string, error) {
	return dns.tfExec.RunTerraformDestroy(dns.ManifestDir, deleteTFVars)
}
