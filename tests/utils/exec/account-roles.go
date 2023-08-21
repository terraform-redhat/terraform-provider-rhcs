package exec

***REMOVED***
	"context"
	"encoding/json"
***REMOVED***

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
***REMOVED***

type AccountRolesArgs struct {
	AccountRolePrefix string `json:"account_role_prefix,omitempty"`
	OCMENV            string `json:"ocm_environment,omitempty"`
	OpenshiftVersion  string `json:"openshift_version,omitempty"`
	Token             string `json:"token,omitempty"`
	URL               string `json:"url,omitempty"`
	ChannelGroup      string `json:"channel_group,omitempty"`
	// Fake              []string `json:"fake,omitempty"`
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

func (acc *AccountRoleService***REMOVED*** Create(createArgs *AccountRolesArgs, extraArgs ...string***REMOVED*** error {
	acc.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	_, err := runTerraformApplyWithArgs(acc.Context, acc.ManifestDir, args***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

func (acc *AccountRoleService***REMOVED*** Output(***REMOVED*** (string, error***REMOVED*** {
	out, err := runTerraformOutput(acc.Context, acc.ManifestDir***REMOVED***
	if err != nil {
		return "", err
	}
	clusterObj := out["cluster_id"]
	clusterID := h.DigString(clusterObj, "value"***REMOVED***
	return clusterID, nil
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
	// if err != nil {
	// 	return err
	// }

	// getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	// getClusterIdCmd.Dir = targetDir
	// _, err = getClusterIdCmd.Output(***REMOVED***

	return err
}

func NewAccountRoleService(manifestDir ...string***REMOVED*** *AccountRoleService {
	acc := &AccountRoleService{}
	acc.Init(manifestDir...***REMOVED***
	return acc
}

// ********************** AccountRoles CMD ******************************
func CreateAccountRoles(varArgs map[string]interface{}, abArgs ...string***REMOVED*** (string, error***REMOVED*** {
	runTerraformInit(context.TODO(***REMOVED***, CON.AccountRolesDir***REMOVED***

	args := combineArgs(varArgs, abArgs...***REMOVED***
	return runTerraformApplyWithArgs(context.TODO(***REMOVED***, CON.AccountRolesDir, args***REMOVED***
}

func DestroyAccountRoles(varArgs map[string]interface{}, abArgs ...string***REMOVED*** error {
	runTerraformInit(context.TODO(***REMOVED***, CON.AccountRolesDir***REMOVED***

	args := combineArgs(varArgs, abArgs...***REMOVED***
	return runTerraformDestroyWithArgs(context.TODO(***REMOVED***, CON.AccountRolesDir, args***REMOVED***
}

func CreateMyTFAccountRoles(accRoleArgs *AccountRolesArgs, arg ...string***REMOVED*** (string, error***REMOVED*** {
	parambytes, _ := json.Marshal(accRoleArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	return CreateAccountRoles(args, arg...***REMOVED***
}

func DestroyMyTFAccountRoles(accRoleArgs *AccountRolesArgs, arg ...string***REMOVED*** error {
	parambytes, _ := json.Marshal(accRoleArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	return DestroyAccountRoles(args, arg...***REMOVED***
}
