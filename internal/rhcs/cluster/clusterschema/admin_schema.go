package clusterschema

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/idps"
)

func AdminFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"username": {
			Description: "Admin username that will be created with the cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"password": {
			Description: "Admin password that will be created with the cluster.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			ForceNew:    true,
		},
	}
}

type AdminCredentials struct {
	Username string `tfsdk:"username"`
	Password string `tfsdk:"password"`
}

func ExpandAdminCredentialsFromResourceData(resourceData *schema.ResourceData) *AdminCredentials {
	list, ok := resourceData.GetOk("admin_credentials")
	if !ok {
		return nil
	}

	return ExpandAdminCredentialsFromInterface(list)
}

func ExpandAdminCredentialsFromInterface(i interface{}) *AdminCredentials {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	adminCredsMap := l[0].(map[string]interface{})
	return &AdminCredentials{
		Username: adminCredsMap["username"].(string),
		Password: adminCredsMap["password"].(string),
	}
}

func FlatAdminCredentials(object *cmv1.Cluster) []interface{} {
	htpasswd, ok := object.GetHtpasswd()
	if !ok || htpasswd == nil {
		return nil
	}

	result := make(map[string]interface{})
	if username, ok := htpasswd.GetUsername(); ok {
		result["username"] = username
	}
	if password, ok := htpasswd.GetPassword(); ok {
		result["password"] = password
	}

	return []interface{}{result}
}

func AdminCredsValidators(i interface{}) error {
	errSumm := "Invalid admin_creedntials"

	// Validate admin username
	adminCreds := ExpandAdminCredentialsFromInterface(i)
	if adminCreds == nil {
		return nil
	}
	if adminCreds.Username == "" {
		return fmt.Errorf(errSumm, "Username can't be empty")
	}
	if err := idps.ValidateHTPasswdUsername(adminCreds.Username); err != nil {
		return fmt.Errorf(errSumm, err.Error())
	}

	//  Validate admin password
	if adminCreds.Password == "" {
		return fmt.Errorf(errSumm, "Password can't be empty")
	}
	if err := idps.ValidateHTPasswdPassword(adminCreds.Password); err != nil {
		return fmt.Errorf(errSumm, err.Error())
	}

	return nil
}
