package exec

import (
	"context"
	"encoding/json"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
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

type AccountRoleService struct {
	CreationArgs *AccountRolesArgs
	ManifestDir  string
	Context      context.Context
}

func (acc *AccountRoleService) Init(manifestDirs ...string) error {
	acc.ManifestDir = CON.AccountRolesDir
	if len(manifestDirs) != 0 {
		acc.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	acc.Context = ctx
	err := runTerraformInit(ctx, acc.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (acc *AccountRoleService) Create(createArgs *AccountRolesArgs, extraArgs ...string) error {
	acc.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(acc.Context, acc.ManifestDir, args)
	if err != nil {
		return err
	}
	return nil
}

func (acc *AccountRoleService) Output() (string, error) {
	out, err := runTerraformOutput(acc.Context, acc.ManifestDir)
	if err != nil {
		return "", err
	}
	clusterObj := out["cluster_id"]
	clusterID := h.DigString(clusterObj, "value")
	return clusterID, nil
}

func (acc *AccountRoleService) Destroy(createArgs ...*AccountRolesArgs) error {
	if acc.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := acc.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args := combineStructArgs(destroyArgs)
	err := runTerraformDestroyWithArgs(acc.Context, acc.ManifestDir, args)
	// if err != nil {
	// 	return err
	// }

	// getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id")
	// getClusterIdCmd.Dir = targetDir
	// _, err = getClusterIdCmd.Output()

	return err
}

func NewAccountRoleService(manifestDir ...string) *AccountRoleService {
	acc := &AccountRoleService{}
	acc.Init(manifestDir...)
	return acc
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
