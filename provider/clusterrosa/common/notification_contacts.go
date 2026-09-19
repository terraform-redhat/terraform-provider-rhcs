// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const notificationContactsBasePath = "/api/accounts_mgmt/v1/subscriptions"

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
	Size  int                           `json:"size"`
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
// from the subscription's notification_contacts sub-resource, paginating
// through all results.
func fetchNotificationContactsWithIDs(
	ctx context.Context,
	connection *sdk.Connection,
	subscriptionID string,
) (map[string]string, error) {
	const pageSize = 100
	result := make(map[string]string)
	path := notificationContactsPath(subscriptionID)
	page := 1
	for {
		resp, err := connection.Get().
			Path(path).
			Parameter("page", page).
			Parameter("size", pageSize).
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
		for _, item := range listResp.Items {
			if item.Username != "" && item.ID != "" {
				result[item.Username] = item.ID
			}
		}
		if listResp.Size < pageSize {
			break
		}
		page++
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
	body, err := json.Marshal(map[string]string{"account_identifier": username})
	if err != nil {
		return fmt.Errorf("can't marshal notification contact '%s': %w", username, err)
	}
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
	username string,
) error {
	deletePath := fmt.Sprintf("%s/%s", notificationContactsPath(subscriptionID), accountID)
	resp, err := connection.Delete().
		Path(deletePath).
		SendContext(ctx)
	if err != nil {
		return fmt.Errorf("can't remove notification contact '%s' (account %s): %w", username, accountID, err)
	}
	if resp.Status() >= 400 {
		return fmt.Errorf("can't remove notification contact '%s' (account %s): HTTP %d: %s",
			username, accountID, resp.Status(), resp.String())
	}
	return nil
}

// UpdateNotificationContacts synchronizes the subscription's notification contacts
// to match the desired list of usernames. It fetches current contacts, computes
// additions and removals, then adds new contacts before deleting removed ones.
// All additions are attempted; failures are collected and reported as a summary error.
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

	seen := make(map[string]bool, len(usernames))
	deduplicated := make([]string, 0, len(usernames))
	for _, u := range usernames {
		if u != "" && !seen[u] {
			seen[u] = true
			deduplicated = append(deduplicated, u)
		}
	}

	var added []string
	var addErrors []string
	for _, username := range deduplicated {
		if _, exists := current[username]; !exists {
			if err := addNotificationContact(ctx, connection, subscriptionID, username); err != nil {
				addErrors = append(addErrors, fmt.Sprintf("'%s': %v", username, err))
				continue
			}
			added = append(added, username)
		}
	}

	if len(addErrors) > 0 {
		summary := fmt.Sprintf("failed to add %d of %d notification contact(s): %s",
			len(addErrors), len(addErrors)+len(added), strings.Join(addErrors, "; "))
		if len(added) > 0 {
			summary += fmt.Sprintf(". Successfully added: %s", strings.Join(added, ", "))
		}
		return fmt.Errorf("%s", summary)
	}

	for username, accountID := range current {
		if !seen[username] {
			if err := removeNotificationContact(ctx, connection, subscriptionID, accountID, username); err != nil {
				return err
			}
		}
	}

	return nil
}
