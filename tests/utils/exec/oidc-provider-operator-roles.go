package exec

***REMOVED***
	"context"
***REMOVED***

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
***REMOVED***

type OIDCProviderOperatorRolesArgs struct {
	AccountRolePrefix  string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix string `json:"operator_role_prefix,omitempty"`
	Token              string `json:"token,omitempty"`
	URL                string `json:"url,omitempty"`
	OIDCConfig         string `json:"oidc_config,omitempty"`
	AWSRegion          string `json:"aws_region,omitempty"`
	OCMENV             string `json:"rhcs_environment,omitempty"`
}

type OIDCProviderOperatorRolesOutput struct {
	OIDCConfigID       string `json:"oidc_config_id,omitempty"`
	AccountRolePrefix  string `json:"account_role_prefix,omitempty"`
	OperatorRolePrefix string `json:"operator_role_prefix,omitempty"`
}

type OIDCProviderOperatorRolesService struct {
	CreationArgs *OIDCProviderOperatorRolesArgs
	ManifestDir  string
	Context      context.Context
}

func (oidcOP *OIDCProviderOperatorRolesService***REMOVED*** Init(manifestDirs ...string***REMOVED*** error {
	oidcOP.ManifestDir = CON.OIDCProviderOperatorRolesManifestDir
	if len(manifestDirs***REMOVED*** != 0 {
		oidcOP.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO(***REMOVED***
	oidcOP.Context = ctx
	err := runTerraformInit(ctx, oidcOP.ManifestDir***REMOVED***
	if err != nil {
		return err
	}
	return nil

}

func (oidcOP *OIDCProviderOperatorRolesService***REMOVED*** Create(createArgs *OIDCProviderOperatorRolesArgs, extraArgs ...string***REMOVED*** (
	*OIDCProviderOperatorRolesOutput, error***REMOVED*** {
	createArgs.URL = CON.GateWayURL
	createArgs.OCMENV = CON.OCMENV
	oidcOP.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	_, err := runTerraformApplyWithArgs(oidcOP.Context, oidcOP.ManifestDir, args***REMOVED***
	if err != nil {
		return nil, err
	}
	output, err := oidcOP.Output(***REMOVED***
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (oidcOP *OIDCProviderOperatorRolesService***REMOVED*** Output(***REMOVED*** (*OIDCProviderOperatorRolesOutput, error***REMOVED*** {
	out, err := runTerraformOutput(oidcOP.Context, oidcOP.ManifestDir***REMOVED***
	if err != nil {
		return nil, err
	}
	var oidcOPOutput = &OIDCProviderOperatorRolesOutput{
		AccountRolePrefix:  h.DigString(out["account_roles_prefix"], "value"***REMOVED***,
		OIDCConfigID:       h.DigString(out["oidc_config_id"], "value"***REMOVED***,
		OperatorRolePrefix: h.DigString(out["operator_role_prefix"], "value"***REMOVED***,
	}
	return oidcOPOutput, nil
}

func (oidcOP *OIDCProviderOperatorRolesService***REMOVED*** Destroy(createArgs ...*OIDCProviderOperatorRolesArgs***REMOVED*** error {
	if oidcOP.CreationArgs == nil && len(createArgs***REMOVED*** == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter"***REMOVED***
	}
	destroyArgs := oidcOP.CreationArgs
	if len(createArgs***REMOVED*** != 0 {
		destroyArgs = createArgs[0]
	}
	destroyArgs.URL = CON.GateWayURL
	destroyArgs.OCMENV = CON.OCMENV
	args := combineStructArgs(destroyArgs***REMOVED***
	err := runTerraformDestroyWithArgs(oidcOP.Context, oidcOP.ManifestDir, args***REMOVED***
	return err
}

func NewOIDCProviderOperatorRolesService(manifestDir ...string***REMOVED*** (*OIDCProviderOperatorRolesService, error***REMOVED*** {
	oidcOP := &OIDCProviderOperatorRolesService{}
	err := oidcOP.Init(manifestDir...***REMOVED***
	return oidcOP, err
}
