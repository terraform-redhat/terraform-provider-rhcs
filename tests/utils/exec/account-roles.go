package exec

***REMOVED***
	"context"
***REMOVED***

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
***REMOVED***

type AccountRolesArgs struct {
	AccountRolePrefix string `json:"account_role_prefix,omitempty"`
	OCMENV            string `json:"rhcs_environment,omitempty"`
	OpenshiftVersion  string `json:"openshift_version,omitempty"`
	Token             string `json:"token,omitempty"`
	URL               string `json:"url,omitempty"`
	ChannelGroup      string `json:"channel_group,omitempty"`
}

type AccountRolesOutput struct {
	AccountRolePrefix string        `json:"account_role_prefix,omitempty"`
	MajorVersion      string        `json:"major_version,omitempty"`
	ChannelGroup      string        `json:"channel_group,omitempty"`
	RHCSGatewayUrl    string        `json:"rhcs_gateway_url,omitempty"`
	RHCSVersions      []interface{} `json:"rhcs_versions,omitempty"`
}

type AccountRoleService struct {
	CreationArgs *AccountRolesArgs
	ManifestDir  string
	Context      context.Context
}

func (acc *AccountRoleService***REMOVED*** Init(manifestDirs ...string***REMOVED*** error {
	acc.ManifestDir = CON.AccountRolesDir
	if len(manifestDirs***REMOVED*** != 0 {
		acc.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO(***REMOVED***
	acc.Context = ctx
	err := runTerraformInit(ctx, acc.ManifestDir***REMOVED***
	if err != nil {
		return err
	}
	return nil

}

func (acc *AccountRoleService***REMOVED*** Create(createArgs *AccountRolesArgs, extraArgs ...string***REMOVED*** (*AccountRolesOutput, error***REMOVED*** {
	acc.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	_, err := runTerraformApplyWithArgs(acc.Context, acc.ManifestDir, args***REMOVED***
	if err != nil {
		return nil, err
	}
	output, err := acc.Output(***REMOVED***
	return output, nil
}

func (acc *AccountRoleService***REMOVED*** Output(***REMOVED*** (*AccountRolesOutput, error***REMOVED*** {
	out, err := runTerraformOutput(acc.Context, acc.ManifestDir***REMOVED***
	if err != nil {
		return nil, err
	}
	var accOutput = &AccountRolesOutput{
		AccountRolePrefix: h.DigString(out["account_roles_prefix"], "value"***REMOVED***,
		MajorVersion:      h.DigString(out["major_version"], "value"***REMOVED***,
		ChannelGroup:      h.DigString(out["channel_group"], "value"***REMOVED***,
		RHCSGatewayUrl:    h.DigString(out["rhcs_gateway_url"], "value"***REMOVED***,
		RHCSVersions:      h.DigArray(out["rhcs_versions"], "value"***REMOVED***,
	}
	return accOutput, nil
}

func (acc *AccountRoleService***REMOVED*** Destroy(createArgs ...*AccountRolesArgs***REMOVED*** error {
	if acc.CreationArgs == nil && len(createArgs***REMOVED*** == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter"***REMOVED***
	}
	destroyArgs := acc.CreationArgs
	if len(createArgs***REMOVED*** != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs***REMOVED***
	err := runTerraformDestroyWithArgs(acc.Context, acc.ManifestDir, args***REMOVED***
	return err
}

func NewAccountRoleService(manifestDir ...string***REMOVED*** (*AccountRoleService, error***REMOVED*** {
	acc := &AccountRoleService{}
	err := acc.Init(manifestDir...***REMOVED***
	return acc, err
}
