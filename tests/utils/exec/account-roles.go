package exec

***REMOVED***
	"context"
	"encoding/json"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
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
