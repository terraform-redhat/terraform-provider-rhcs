package exec

import (
	"context"
	"encoding/json"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

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
func CreateAccountRoles(varArgs map[string]interface{}, abArgs ...string) (string, error) {
	runTerraformInit(context.TODO(), CON.AccountRolesDir)

	args := combineArgs(varArgs, abArgs...)
	return runTerraformApplyWithArgs(context.TODO(), CON.AccountRolesDir, args)
}

func DestroyAccountRoles(varArgs map[string]interface{}, abArgs ...string) error {
	runTerraformInit(context.TODO(), CON.AccountRolesDir)

	args := combineArgs(varArgs, abArgs...)
	return runTerraformDestroyWithArgs(context.TODO(), CON.AccountRolesDir, args)
}

func CreateMyTFAccountRoles(accRoleArgs *AccountRolesArgs, arg ...string) (string, error) {
	parambytes, _ := json.Marshal(accRoleArgs)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	return CreateAccountRoles(args, arg...)
}

func DestroyMyTFAccountRoles(accRoleArgs *AccountRolesArgs, arg ...string) error {
	parambytes, _ := json.Marshal(accRoleArgs)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	return DestroyAccountRoles(args, arg...)
}
