// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const notificationContactsResourceDescription = "Set of Red Hat OpenShift " +
	"Cluster Manager (OCM) account usernames to receive cluster notification " +
	"emails. Values must be OCM usernames (not email addresses). " +
	"While email addresses are accepted by the API, they are " +
	"resolved to usernames internally, which will cause " +
	"persistent plan diffs and may lead to incorrect contact " +
	"removal on subsequent applies. " +
	"All contacts must belong to the same Red Hat organization " +
	"as the cluster. " +
	"This attribute is configured after cluster creation (Day 2). " +
	"By default, the cluster creator is set as the notification " +
	"contact. For clusters created with service accounts, " +
	"no default contact is set."

const notificationContactsDatasourceDescription = "Set of Red Hat OpenShift " +
	"Cluster Manager (OCM) account usernames configured to receive cluster " +
	"notification emails."

const notificationContactsBasePath = "/api/accounts_mgmt/v1/subscriptions"

// NotificationContactsResourceSchema returns the schema definition for the notification_contacts
// resource attribute.
func NotificationContactsResourceSchema() schema.SetAttribute {
	return schema.SetAttribute{
		Description: notificationContactsResourceDescription,
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
	}
}

// NotificationContactsDatasourceSchema returns the schema definition for the notification_contacts
// data source attribute.
func NotificationContactsDatasourceSchema() schema.SetAttribute {
	return schema.SetAttribute{
		Description: notificationContactsDatasourceDescription,
		ElementType: types.StringType,
		Computed:    true,
	}
}

// GetSubscriptionID extracts the subscription ID from a ClustersMgmt Cluster object.
func GetSubscriptionID(cluster *cmv1.Cluster) (string, bool) {
	if cluster == nil {
		return "", false
	}
	sub := cluster.Subscription()
	if sub == nil {
		return "", false
	}
	id := sub.ID()
	return id, id != ""
}

// notificationContactResponse represents a single notification contact from the API.
type notificationContactResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// notificationContactListResponse represents the API response containing a list of notification contacts.
type notificationContactListResponse struct {
	Items []notificationContactResponse `json:"items"`
}

// notificationContactsPath returns the API path for a subscription's notification contacts.
func notificationContactsPath(subscriptionID string) string {
	return fmt.Sprintf("%s/%s/notification_contacts", notificationContactsBasePath, subscriptionID)
}

// FetchNotificationContacts reads notification contacts from the subscription's
// notification_contacts sub-resource.
func FetchNotificationContacts(
	ctx context.Context,
	connection *sdk.Connection,
	subscriptionID string,
) ([]string, error) {
	contactMap, err := fetchNotificationContactsWithIDs(ctx, connection, subscriptionID)
	if err != nil {
		return nil, err
	}
	if len(contactMap) == 0 {
		return nil, nil
	}
	usernames := make([]string, 0, len(contactMap))
	for username := range contactMap {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)
	return usernames, nil
}

// fetchNotificationContactsWithIDs returns a map of username -> account ID
// from the subscription's notification_contacts sub-resource.
func fetchNotificationContactsWithIDs(
	ctx context.Context,
	connection *sdk.Connection,
	subscriptionID string,
) (map[string]string, error) {
	resp, err := connection.Get().
		Path(notificationContactsPath(subscriptionID)).
		SendContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't read notification contacts for subscription '%s': %w", subscriptionID, err)
	}
	if resp.Status() >= 400 {
		return nil, fmt.Errorf("can't read notification contacts for subscription '%s': HTTP %d: %s",
			subscriptionID, resp.Status(), resp.String())
	}
	var listResp notificationContactListResponse
	if err := json.Unmarshal(resp.Bytes(), &listResp); err != nil {
		return nil, fmt.Errorf("can't parse notification contacts response: %w", err)
	}
	result := make(map[string]string, len(listResp.Items))
	for _, item := range listResp.Items {
		if item.Username != "" && item.ID != "" {
			result[item.Username] = item.ID
		}
	}
	return result, nil
}

// addNotificationContact adds a single notification contact by username using
// POST to the notification_contacts sub-resource.
func addNotificationContact(
	ctx context.Context,
	connection *sdk.Connection,
	subscriptionID string,
	username string,
) error {
	body, _ := json.Marshal(map[string]string{"account_identifier": username})
	resp, err := connection.Post().
		Path(notificationContactsPath(subscriptionID)).
		Bytes(body).
		SendContext(ctx)
	if err != nil {
		return fmt.Errorf("can't add notification contact '%s': %w", username, err)
	}
	if resp.Status() >= 400 {
		return fmt.Errorf("can't add notification contact '%s': HTTP %d: %s",
			username, resp.Status(), resp.String())
	}
	return nil
}

// removeNotificationContact removes a single notification contact by account ID using
// DELETE on the notification_contacts sub-resource.
func removeNotificationContact(
	ctx context.Context,
	connection *sdk.Connection,
	subscriptionID string,
	accountID string,
) error {
	deletePath := fmt.Sprintf("%s/%s", notificationContactsPath(subscriptionID), accountID)
	resp, err := connection.Delete().
		Path(deletePath).
		SendContext(ctx)
	if err != nil {
		return fmt.Errorf("can't remove notification contact '%s': %w", accountID, err)
	}
	if resp.Status() >= 400 {
		return fmt.Errorf("can't remove notification contact '%s': HTTP %d: %s",
			accountID, resp.Status(), resp.String())
	}
	return nil
}

// UpdateNotificationContacts synchronizes the subscription's notification contacts
// to match the desired list of usernames. It fetches current contacts, computes
// additions and removals, then POSTs/DELETEs as needed.
func UpdateNotificationContacts(
	ctx context.Context,
	connection *sdk.Connection,
	subscriptionID string,
	usernames []string,
) error {
	current, err := fetchNotificationContactsWithIDs(ctx, connection, subscriptionID)
	if err != nil {
		return err
	}

	desired := make(map[string]bool, len(usernames))
	for _, u := range usernames {
		desired[u] = true
	}

	for username, accountID := range current {
		if !desired[username] {
			if err := removeNotificationContact(ctx, connection, subscriptionID, accountID); err != nil {
				return err
			}
		}
	}

	for _, username := range usernames {
		if _, exists := current[username]; !exists {
			if err := addNotificationContact(ctx, connection, subscriptionID, username); err != nil {
				return err
			}
		}
	}

	return nil
}

// ResolveNotificationContacts resolves notification contacts from the cluster's subscription,
// returning a types.Set suitable for Terraform state.
func ResolveNotificationContacts(
	ctx context.Context,
	connection *sdk.Connection,
	cluster *cmv1.Cluster,
) (types.Set, diag.Diagnostics) {
	subID, ok := GetSubscriptionID(cluster)
	if !ok {
		return types.SetNull(types.StringType), diag.Diagnostics{
			diag.NewWarningDiagnostic(
				"Can't read notification contacts",
				"Cluster subscription ID is not available. "+
					"Notification contacts will be populated on the next terraform apply.",
			),
		}
	}
	return ResolveNotificationContactsBySubID(ctx, connection, subID)
}

// ResolveNotificationContactsBySubID resolves notification contacts using a known subscription ID,
// returning a types.Set suitable for Terraform state. Use this when the subscription ID was
// captured earlier (e.g. from the Create response) and the current cluster object may not
// include the subscription link.
func ResolveNotificationContactsBySubID(
	ctx context.Context,
	connection *sdk.Connection,
	subscriptionID string,
) (types.Set, diag.Diagnostics) {
	usernames, err := FetchNotificationContacts(ctx, connection, subscriptionID)
	if err != nil {
		return types.SetNull(types.StringType), diag.Diagnostics{
			diag.NewWarningDiagnostic(
				"Can't read notification contacts",
				fmt.Sprintf(
					"Could not read notification contacts from the API: %v. "+
						"Run terraform apply again to refresh.",
					err,
				),
			),
		}
	}

	if len(usernames) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{}), nil
	}

	vals := make([]attr.Value, len(usernames))
	for i, u := range usernames {
		vals[i] = types.StringValue(u)
	}
	return types.SetValueMust(types.StringType, vals), nil
}
