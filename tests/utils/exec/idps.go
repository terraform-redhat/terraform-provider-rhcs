package exec

***REMOVED***
	"context"
***REMOVED***

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
***REMOVED***

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

func (idp *IDPService***REMOVED*** Init(manifestDirs ...string***REMOVED*** error {
	idp.ManifestDir = CON.IDPsDir
	if len(manifestDirs***REMOVED*** != 0 {
		idp.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO(***REMOVED***
	idp.Context = ctx
	err := runTerraformInit(ctx, idp.ManifestDir***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

func (idp *IDPService***REMOVED*** Create(createArgs *IDPArgs, extraArgs ...string***REMOVED*** error {
	idp.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	_, err := runTerraformApplyWithArgs(idp.Context, idp.ManifestDir, args***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

func (idp *IDPService***REMOVED*** Output(***REMOVED*** (IDPOutput, error***REMOVED*** {
	idpDir := CON.IDPsDir
	if idp.ManifestDir != "" {
		idpDir = idp.ManifestDir
	}
	var output IDPOutput
	out, err := runTerraformOutput(context.TODO(***REMOVED***, idpDir***REMOVED***
	if err != nil {
		return output, err
	}
	if err != nil {
		return output, err
	}
	id := h.DigString(out["idp_id"], "value"***REMOVED***

	// right now only "holds" id, more vars might be needed in the future
	output = IDPOutput{
		ID: id,
	}
	return output, nil
}

func (idp *IDPService***REMOVED*** Destroy(createArgs ...*IDPArgs***REMOVED*** error {
	if idp.CreationArgs == nil && len(createArgs***REMOVED*** == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter"***REMOVED***
	}
	destroyArgs := idp.CreationArgs
	if len(createArgs***REMOVED*** != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs***REMOVED***
	err := runTerraformDestroyWithArgs(idp.Context, idp.ManifestDir, args***REMOVED***

	return err
}

func NewIDPService(manifestDir ...string***REMOVED*** *IDPService {
	idp := &IDPService{}
	idp.Init(manifestDir...***REMOVED***
	return idp
}
