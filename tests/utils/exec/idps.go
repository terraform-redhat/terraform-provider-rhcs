package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type IDPArgs struct {
	ClusterID     string        `json:"cluster_id,omitempty"`
	Name          string        `json:"name,omitempty"`
	ID            string        `json:"id,omitempty"`
	Token         string        `json:"token,omitempty"`
	OCMENV        string        `json:"ocm_environment,omitempty"`
	URL           string        `json:"url,omitempty"`
	MappingMethod string        `json:"mapping_method,omitempty"`
	HtpasswdUsers []interface{} `json:"htpasswd_users,omitempty"`
}

type IDPService struct {
	CreationArgs *IDPArgs
	ManifestDir  string
	Context      context.Context
}

// for now holds only ID, additional vars might be needed in the future
type IDPOutput struct {
	ID string `json:"idp_id,omitempty"`
}

func (idp *IDPService) Init(manifestDirs ...string) error {
	idp.ManifestDir = CON.IDPsDir
	if len(manifestDirs) != 0 {
		idp.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	idp.Context = ctx
	err := runTerraformInit(ctx, idp.ManifestDir)
	if err != nil {
		return err
	}
	return nil
}

func (idp *IDPService) Create(createArgs *IDPArgs, extraArgs ...string) error {
	idp.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(idp.Context, idp.ManifestDir, args)
	if err != nil {
		return err
	}
	return nil
}

func (idp *IDPService) Output() (IDPOutput, error) {
	idpDir := CON.IDPsDir
	if idp.ManifestDir != "" {
		idpDir = idp.ManifestDir
	}
	var output IDPOutput
	out, err := runTerraformOutput(context.TODO(), idpDir)
	if err != nil {
		return output, err
	}
	if err != nil {
		return output, err
	}
	id := h.DigString(out["idp_id"], "value")

	// right now only "holds" id, more vars might be needed in the future
	output = IDPOutput{
		ID: id,
	}
	return output, nil
}

func (idp *IDPService) Destroy(createArgs ...*IDPArgs) error {
	if idp.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := idp.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs)
	err := runTerraformDestroyWithArgs(idp.Context, idp.ManifestDir, args)

	return err
}

func NewIDPService(manifestDir ...string) *IDPService {
	idp := &IDPService{}
	idp.Init(manifestDir...)
	return idp
}
