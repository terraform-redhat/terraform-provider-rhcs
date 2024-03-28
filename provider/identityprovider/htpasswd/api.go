package htpasswd

import (
	"context"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func DeleteUser(ctx context.Context, id string, resource *v1.IdentityProviderClient) error {
	userToDelete := resource.HtpasswdUsers().HtpasswdUser(id)
	deleteRequest := userToDelete.Delete()
	_, err := deleteRequest.SendContext(ctx)
	return err
}

func UpdateUser(ctx context.Context, password string, id string, resource *v1.IdentityProviderClient) error {
	userToPatch, err := (&v1.HTPasswdUserBuilder{}).
		Password(password).Build()
	if err != nil {
		return err
	}
	_, err = resource.HtpasswdUsers().HtpasswdUser(id).Update().
		Body(userToPatch).SendContext(ctx)
	return err
}

func AddUser(ctx context.Context, username string, password string, resource *v1.IdentityProviderClient) error {
	userToAdd, err := (&v1.HTPasswdUserBuilder{}).Username(username).
		Password(password).Build()
	if err != nil {
		return err
	}
	addRequest := resource.HtpasswdUsers().Add()
	_, err = addRequest.Body(userToAdd).SendContext(ctx)
	return err
}
