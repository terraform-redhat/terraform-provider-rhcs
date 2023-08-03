package idps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type HTPasswdUser struct {
	Username string `tfsdk:"username"`
	Password string `tfsdk:"password"`
}

type HTPasswdIdentityProvider struct {
	Users []HTPasswdUser `tfsdk:"users"`
}

func HtpasswdSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"users": {
			Description: "A list of htpasswd user credentials",
			Type:        schema.TypeList,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: HTPasswdUserList(),
			},
			Required: true,
		},
	}
}
func HTPasswdUserList() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"username": {
			Description: "User username.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"password": {
			Description: "User password.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
		},
	}
}

func ExpandHTPASSFromResourceData(resourceData *schema.ResourceData) *HTPasswdIdentityProvider {
	list, ok := resourceData.GetOk("htpasswd")
	if !ok {
		return nil
	}

	return ExpandHTPASSFromInterface(list)
}

func ExpandHTPASSFromInterface(i interface{}) *HTPasswdIdentityProvider {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	htpassMap := l[0].(map[string]interface{})
	return &HTPasswdIdentityProvider{
		Users: ExpandHTPasswdUserInterface(htpassMap["users"]),
	}
}

func ExpandHTPasswdUserInterface(i interface{}) []HTPasswdUser {
	l := i.([]interface{})
	if len(l) == 0 {
		return nil
	}

	result := []HTPasswdUser{}
	for _, user := range l {
		userMap := user.(map[string]interface{})
		result = append(result, HTPasswdUser{
			Username: userMap["username"].(string),
			Password: userMap["password"].(string),
		})
	}
	return result

}

func CreateHTPasswdIDPBuilder(state *HTPasswdIdentityProvider) *cmv1.HTPasswdIdentityProviderBuilder {
	builder := cmv1.NewHTPasswdIdentityProvider()
	userListBuilder := cmv1.NewHTPasswdUserList()
	userList := []*cmv1.HTPasswdUserBuilder{}
	for _, user := range state.Users {
		userBuilder := &cmv1.HTPasswdUserBuilder{}
		userBuilder.Username(user.Username)
		userBuilder.Password(user.Password)
		userList = append(userList, userBuilder)
	}
	userListBuilder.Items(userList...)
	builder.Users(userListBuilder)
	return builder
}

func FlatHtpasswd(object *cmv1.IdentityProvider) []interface{} {
	htpasswdObject, ok := object.GetHtpasswd()
	if !ok {
		return nil
	}
	users, ok := htpasswdObject.GetUsers()
	if !ok {
		return nil
	}

	listOfUsers := []interface{}{}
	for _, user := range users.Slice() {
		if user == nil {
			continue
		}
		userMap := make(map[string]string)
		userMap["username"] = user.Username()
		userMap["password"] = user.Password()
		listOfUsers = append(listOfUsers, userMap)
	}
	return listOfUsers
}

func HTPasswdValidators(i interface{}) error {
	errSumm := "Invalid HTPasswd IDP resource configuration, %s"
	htpass := ExpandHTPASSFromInterface(i)
	if htpass == nil {
		return nil
	}
	if len(htpass.Users) < 1 {
		return fmt.Errorf(errSumm, "Must provide at least one user.")
	}
	for index, user := range htpass.Users {
		if err := ValidateHTPasswdUsername(user.Username); err != nil {
			return fmt.Errorf(errSumm, fmt.Sprintf("Invalid username @ index %d. Error: %s", index, err.Error()))
		}
		if err := ValidateHTPasswdPassword(user.Password); err != nil {
			return fmt.Errorf(errSumm, fmt.Sprintf("Invalid password @ index %d. Error: %s", index, err.Error()))
		}
	}

	return nil
}

func ValidateHTPasswdUsername(username string) error {
	if strings.ContainsAny(username, "/:%") {
		return fmt.Errorf("invalid username '%s': "+
			"username must not contain /, :, or %%", username)
	}
	return nil
}

func ValidateHTPasswdPassword(password string) error {
	notAsciiOnly, _ := regexp.MatchString(`[^\x20-\x7E]`, password)
	containsSpace := strings.Contains(password, " ")
	tooShort := len(password) < 14
	if notAsciiOnly || containsSpace || tooShort {
		return fmt.Errorf(
			"password must be at least 14 characters (ASCII-standard) without whitespaces")
	}
	hasUppercase, _ := regexp.MatchString(`[A-Z]`, password)
	hasLowercase, _ := regexp.MatchString(`[a-z]`, password)
	hasNumberOrSymbol, _ := regexp.MatchString(`[^a-zA-Z]`, password)
	if !hasUppercase || !hasLowercase || !hasNumberOrSymbol {
		return fmt.Errorf(
			"password must include uppercase letters, lowercase letters, and numbers " +
				"or symbols (ASCII-standard characters only)")
	}
	return nil
}
