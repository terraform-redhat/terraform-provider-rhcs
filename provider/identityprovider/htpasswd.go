package identityprovider

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	idputils "github.com/openshift-online/ocm-common/pkg/idp/utils"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/identityprovider/htpasswd"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const (
	HTPasswdMinPassLength = 14
)

var (
	HTPasswdPassRegexAscii          = regexp.MustCompile(`^[\x20-\x7E]+$`)
	HTPasswdPassRegexHasUpper       = regexp.MustCompile(`[A-Z]`)
	HTPasswdPassRegexHasLower       = regexp.MustCompile(`[a-z]`)
	HTPasswdPassRegexHasNumOrSymbol = regexp.MustCompile(`[^a-zA-Z]`)

	HTPasswdPasswordValidators = []validator.String{
		stringvalidator.LengthAtLeast(HTPasswdMinPassLength),
		stringvalidator.RegexMatches(HTPasswdPassRegexAscii, "password should use ASCII-standard characters only"),
		stringvalidator.RegexMatches(HTPasswdPassRegexHasUpper, "password must contain uppercase characters"),
		stringvalidator.RegexMatches(HTPasswdPassRegexHasLower, "password must contain lowercase characters"),
		stringvalidator.RegexMatches(HTPasswdPassRegexHasNumOrSymbol, "password must contain numbers or symbols"),
	}

	HTPasswdUsernameValidators = []validator.String{
		stringvalidator.RegexMatches(regexp.MustCompile(`^[^/:%]*$`), "username may not contain the characters: '/:%'"),
	}
)

type HTPasswdUser struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type HTPasswdIdentityProvider struct {
	Users []HTPasswdUser `tfsdk:"users"`
}

var htpasswdSchema = map[string]schema.Attribute{
	"users": schema.ListNestedAttribute{
		Description: "A list of htpasswd user credentials",
		NestedObject: schema.NestedAttributeObject{
			Attributes: htpasswdUserList,
		},
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			uniqueUsernameValidator(),
		},
		Required: true,
	},
}

var htpasswdUserList = map[string]schema.Attribute{
	"username": schema.StringAttribute{
		Description: "User username.",
		Required:    true,
		Validators:  HTPasswdUsernameValidators,
	},
	"password": schema.StringAttribute{
		Description: "User password.",
		Required:    true,
		Sensitive:   true,
		Validators:  HTPasswdPasswordValidators,
	},
}

func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider) (*cmv1.HTPasswdIdentityProviderBuilder, error) {
	builder := cmv1.NewHTPasswdIdentityProvider()
	userListBuilder := cmv1.NewHTPasswdUserList()
	userList := []*cmv1.HTPasswdUserBuilder{}
	for _, user := range state.Users {
		hashedPwd, err := idputils.GenerateHTPasswdCompatibleHash(user.Password.ValueString())
		if err != nil {
			return nil, err
		}
		if os.Getenv("IS_TEST") == "true" {
			hashedPwd = fmt.Sprintf("hash(%s)", user.Password.ValueString())
		}
		userBuilder := &cmv1.HTPasswdUserBuilder{}
		userBuilder.Username(user.Username.ValueString())
		userBuilder.HashedPassword(hashedPwd)
		userList = append(userList, userBuilder)
	}
	userListBuilder.Items(userList...)
	builder.Users(userListBuilder)
	return builder, nil
}

func uniqueUsernameValidator() validator.List {
	return attrvalidators.NewListValidator("userlist unique username", func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
		usersList := req.ConfigValue
		htusers := []HTPasswdUser{}
		err := usersList.ElementsAs(ctx, &htusers, true)
		if err != nil {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid list conversion", "Failed to parse userlist")
			return
		}
		usernames := make(map[string]bool)
		for _, user := range htusers {
			if _, ok := usernames[user.Username.ValueString()]; ok {
				// Username already exists
				resp.Diagnostics.AddAttributeError(req.Path, fmt.Sprintf("Found duplicate username: '%s'", user.Username.ValueString()), "Usernames in HTPasswd user list must be unique")
				return
			}
			usernames[user.Username.ValueString()] = true
		}
	})
}

func htPasswdUserListToStringMaps(ctx context.Context, users []HTPasswdUser,
	resource *cmv1.IdentityProviderClient) (map[string]htpasswd.HtPasswdUserWithId, error) {
	csUserMap := make(map[string]htpasswd.HtPasswdUserWithId)
	get, err := resource.HtpasswdUsers().List().SendContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, user := range get.Items().Slice() {
		csUserMap[user.Username()] = htpasswd.HtPasswdUserWithId{user.ID(), user.Username(), ""}
	}
	for _, user := range users {
		csUserMap[user.Username.ValueString()] = htpasswd.HtPasswdUserWithId{csUserMap[user.Username.ValueString()].Id,
			csUserMap[user.Username.ValueString()].Username, user.Password.ValueString()}
	}
	finalUserMap := make(map[string]htpasswd.HtPasswdUserWithId)
	// Remove deleted users
	for _, user := range users {
		if _, ok := csUserMap[user.Username.ValueString()]; ok {
			finalUserMap[user.Username.ValueString()] = csUserMap[user.Username.ValueString()]
		}
	}
	return finalUserMap, nil
}

func UpdateHTPasswd(ctx context.Context, resource *cmv1.IdentityProviderClient, state *IdentityProviderState,
	plan *IdentityProviderState, response *resource.UpdateResponse) {
	if reflect.DeepEqual(state.HTPasswd.Users, plan.HTPasswd.Users) {
		return
	}
	stateUserMap, err := htPasswdUserListToStringMaps(ctx, state.HTPasswd.Users, resource)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't get identity provider user ID",
			fmt.Sprintf(
				"Can't get identity provider user ID with identifier for "+
					"cluster '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return
	}
	planUserMap, err := htPasswdUserListToStringMaps(ctx, plan.HTPasswd.Users, resource)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't get identity provider user ID",
			fmt.Sprintf(
				"Can't get identity provider user ID with identifier for "+
					"cluster '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return
	}

	patchParams := htpasswd.PatchParams{
		ctx, stateUserMap, planUserMap, resource,
		[]string{}, state.Cluster.ValueString(), response,
	}

	patchParams.RemovedUsers, err = htpasswd.DeleteUserFromState(patchParams)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete identity provider user",
			fmt.Sprintf(
				"Can't delete identity provider user: '%v'", err,
			),
		)
		return
	}

	err = htpasswd.PatchOrAddUserInState(patchParams)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't patch/add identity provider user",
			fmt.Sprintf(
				"Can't patch/add identity provider user: '%v'", err,
			),
		)
		return
	}
}
