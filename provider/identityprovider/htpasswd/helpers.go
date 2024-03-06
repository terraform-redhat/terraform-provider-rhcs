package htpasswd

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/pkg/errors"
)

type HtPasswdUserWithId struct {
	Id       string
	Username string
	Password string
}

type PatchParams struct {
	Ctx          context.Context
	StateUserMap map[string]HtPasswdUserWithId
	PlanUserMap  map[string]HtPasswdUserWithId
	Resource     *v1.IdentityProviderClient
	RemovedUsers []string
	ClusterId    string
	Response     *resource.UpdateResponse
}

func DeleteUserFromState(params PatchParams) ([]string, error) {
	if params.Ctx == nil || params.StateUserMap == nil || params.PlanUserMap == nil || params.Resource == nil ||
		params.Response == nil {
		return []string{}, errors.Errorf("Unable to delete user from htpasswd IDP, nil param")
	}

	removedUsers := []string{}
	for user, htpasswdUser := range params.StateUserMap {
		if params.PlanUserMap[user].Username == "" { // Not in plan, delete
			err := DeleteUser(params.Ctx, htpasswdUser.Id, params.Resource)
			if err != nil {
				return removedUsers, err
			}
			removedUsers = append(removedUsers, user)
		}
	}
	return removedUsers, nil
}

func PatchOrAddUserInState(params PatchParams) error {
	if params.Ctx == nil || params.StateUserMap == nil || params.PlanUserMap == nil || params.Resource == nil ||
		params.RemovedUsers == nil || params.Response == nil {
		return errors.Errorf("Unable to patch or add user to htpasswd IDP, nil param")
	}
	for user, planValue := range params.PlanUserMap {
		if stateValue, ok := params.StateUserMap[user]; ok { // Is in current state
			if planValue != stateValue && !slices.Contains(params.RemovedUsers, user) { // Password changed, update
				err := UpdateUser(params.Ctx, planValue.Password, planValue.Id, params.Resource)
				if err != nil {
					return err
				}
			}
		} else { // Should be added (not in current state)
			if !slices.Contains(params.RemovedUsers, user) {
				err := AddUser(params.Ctx, user, planValue.Password, params.Resource)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
