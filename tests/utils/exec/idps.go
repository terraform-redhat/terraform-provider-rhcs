package exec

import (
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type IDPArgs struct {
	ClusterID     string        `json:"cluster_id,omitempty"`
	Name          string        `json:"name,omitempty"`
	ID            string        `json:"id,omitempty"`
	OCMENV        string        `json:"ocm_environment,omitempty"`
	URL           string        `json:"url,omitempty"`
	CA            string        `json:"ca,omitempty"`
	Attributes    interface{}   `json:"attributes,omitempty"`
	ClientID      string        `json:"client_id,omitempty"`
	ClientSecret  string        `json:"client_secret,omitempty"`
	Organizations []string      `json:"organizations,omitempty"`
	HostedDomain  string        `json:"hosted_domain,omitempty"`
	Insecure      bool          `json:"insecure,omitempty"`
	MappingMethod string        `json:"mapping_method,omitempty"`
	HtpasswdUsers []interface{} `json:"htpasswd_users,omitempty"`
}

func (args *IDPArgs) appendURL() {
	args.URL = CON.GateWayURL
}

type IDPService struct {
	tfExec      TerraformExec
	ManifestDir string
	idpType     string
}

// for now holds only ID, additional vars might be needed in the future
type IDPOutput struct {
	ID string `json:"idp_id,omitempty"`
}

func newIDPService(idpType CON.IDPType, clusterType CON.ClusterType, tfExec TerraformExec) (*IDPService, error) {
	idp := &IDPService{
		ManifestDir: GetIDPsManifestDir(idpType, clusterType),
		tfExec:      tfExec,
	}
	err := idp.Init()
	return idp, err
}

func (idps *IDPService) Init() error {
	return idps.tfExec.RunTerraformInit(idps.ManifestDir)
}

func (idps *IDPService) Apply(createArgs *IDPArgs) (output string, err error) {
	createArgs.appendURL()

	var tfVars *TFVars
	tfVars, err = NewTFArgs(createArgs)
	if err != nil {
		return
	}
	output, err = idps.tfExec.RunTerraformApply(idps.ManifestDir, tfVars)
	return
}

func (idps *IDPService) Output() (*IDPOutput, error) {
	out, err := idps.tfExec.RunTerraformOutput(idps.ManifestDir)
	if err != nil {
		return nil, err
	}

	// right now only "holds" id, more vars might be needed in the future
	output := &IDPOutput{
		ID: h.DigString(out["idp_id"], "value"),
	}
	return output, nil
}

func (idps *IDPService) Destroy(deleteTFVars bool) (string, error) {
	return idps.tfExec.RunTerraformDestroy(idps.ManifestDir, deleteTFVars)
}
