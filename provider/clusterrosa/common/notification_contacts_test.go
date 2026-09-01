// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"
)

var _ = Describe("Notification contacts helpers", func() {
	Context("GetSubscriptionID", func() {
		It("returns the subscription ID when present", func() {
			cluster, err := cmv1.NewCluster().
				Subscription(cmv1.NewSubscription().ID("sub-123")).
				Build()
			Expect(err).NotTo(HaveOccurred())

			id, ok := GetSubscriptionID(cluster)
			Expect(ok).To(BeTrue())
			Expect(id).To(Equal("sub-123"))
		})

		It("returns false when cluster is nil", func() {
			id, ok := GetSubscriptionID(nil)
			Expect(ok).To(BeFalse())
			Expect(id).To(BeEmpty())
		})

		It("returns false when subscription is not set", func() {
			cluster, err := cmv1.NewCluster().Build()
			Expect(err).NotTo(HaveOccurred())

			id, ok := GetSubscriptionID(cluster)
			Expect(ok).To(BeFalse())
			Expect(id).To(BeEmpty())
		})
	})

	Context("API operations", func() {
		var (
			server     *ghttp.Server
			ca         string
			connection *sdk.Connection
			ctx        context.Context
		)

		BeforeEach(func() {
			server, ca = sdktesting.MakeTCPTLSServer()
			token := sdktesting.MakeTokenString("Bearer", 10*time.Minute)
			ctx = context.Background()
			var err error
			connection, err = sdk.NewConnectionBuilder().
				URL(server.URL()).
				TrustedCAFile(ca).
				Tokens(token).
				BuildContext(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			server.Close()
			connection.Close()
		})

		Context("FetchNotificationContacts", func() {
			It("returns sorted usernames from the sub-resource", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-z", "username": "zuser"},
								{"kind": "Account", "id": "acc-a", "username": "auser"}
							],
							"size": 2,
							"total": 2
						}`),
					),
				)

				usernames, err := FetchNotificationContacts(ctx, connection, "sub-123")
				Expect(err).NotTo(HaveOccurred())
				Expect(usernames).To(Equal([]string{"auser", "zuser"}))
			})

			It("returns nil when no contacts are set", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 0
						}`),
					),
				)

				usernames, err := FetchNotificationContacts(ctx, connection, "sub-123")
				Expect(err).NotTo(HaveOccurred())
				Expect(usernames).To(BeNil())
			})

			It("returns an error when the API call fails", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusNotFound,
							`{"kind":"Error","id":"404","href":"/api/accounts_mgmt/v1/errors/404","code":"AMS-404","reason":"not found"}`),
					),
				)

				_, err := FetchNotificationContacts(ctx, connection, "sub-123")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can't read notification contacts"))
			})

			It("paginates when size field exceeds actual item count", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						ghttp.VerifyFormKV("page", "1"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-b", "username": "buser"},
								{"kind": "Account", "id": "acc-c", "username": "cuser"}
							],
							"size": 100,
							"total": 3
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						ghttp.VerifyFormKV("page", "2"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-a", "username": "auser"}
							],
							"size": 100,
							"total": 3
						}`),
					),
				)

				usernames, err := FetchNotificationContacts(ctx, connection, "sub-123")
				Expect(err).NotTo(HaveOccurred())
				Expect(usernames).To(Equal([]string{"auser", "buser", "cuser"}))
			})

			It("paginates across two pages and combines results", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						ghttp.VerifyFormKV("page", "1"),
						ghttp.VerifyFormKV("size", "100"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-b", "username": "buser"},
								{"kind": "Account", "id": "acc-c", "username": "cuser"}
							],
							"size": 2,
							"total": 3
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						ghttp.VerifyFormKV("page", "2"),
						ghttp.VerifyFormKV("size", "100"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-a", "username": "auser"}
							],
							"size": 1,
							"total": 3
						}`),
					),
				)

				usernames, err := FetchNotificationContacts(ctx, connection, "sub-123")
				Expect(err).NotTo(HaveOccurred())
				Expect(usernames).To(Equal([]string{"auser", "buser", "cuser"}))
			})

			It("returns error when server returns empty page before total is reached", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						ghttp.VerifyFormKV("page", "1"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-a", "username": "auser"}
							],
							"size": 1,
							"total": 3
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						ghttp.VerifyFormKV("page", "2"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 3
						}`),
					),
				)

				_, err := FetchNotificationContacts(ctx, connection, "sub-123")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty page 2"))
				Expect(err.Error()).To(ContainSubstring("1 of 3"))
			})
		})

		Context("UpdateNotificationContacts", func() {
			It("reports no mutation when initial fetch fails", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusForbidden,
							`{"kind":"Error","id":"403","reason":"forbidden"}`),
					),
				)

				mutated, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1"})
				Expect(err).To(HaveOccurred())
				Expect(mutated).To(BeFalse())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("adds contacts that don't exist yet", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 0
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "user1"),
						sdktesting.RespondWithJSON(http.StatusCreated, `{
							"kind": "Account", "id": "acc-1", "username": "user1"
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "user2"),
						sdktesting.RespondWithJSON(http.StatusCreated, `{
							"kind": "Account", "id": "acc-2", "username": "user2"
						}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1", "user2"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("removes contacts not in desired list", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"},
								{"kind": "Account", "id": "acc-2", "username": "user2"}
							],
							"size": 2,
							"total": 2
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-2"),
						sdktesting.RespondWithJSON(http.StatusNoContent, ""),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("clears all contacts with empty list", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"}
							],
							"size": 1,
							"total": 1
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-1"),
						sdktesting.RespondWithJSON(http.StatusNoContent, ""),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("handles mixed add and remove with add-before-delete ordering", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"},
								{"kind": "Account", "id": "acc-2", "username": "user2"}
							],
							"size": 2,
							"total": 2
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "user3"),
						sdktesting.RespondWithJSON(http.StatusCreated, `{
							"kind": "Account", "id": "acc-3", "username": "user3"
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-1"),
						sdktesting.RespondWithJSON(http.StatusNoContent, ""),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user2", "user3"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("is a no-op when desired matches current", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"}
							],
							"size": 1,
							"total": 1
						}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error when username is not found", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 0
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusNotFound,
							`{"kind":"Error","id":"404","reason":"account not found"}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"nonexistent"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to add 1 of 1"))
			})

			It("continues adding remaining contacts after one fails and reports summary", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 0
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusBadRequest,
							`{"kind":"Error","id":"400","reason":"invalid username"}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusCreated, `{
							"kind": "Account", "id": "acc-2", "username": "good_user"
						}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"bad_user", "good_user"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to add 1 of 2"))
				Expect(err.Error()).To(ContainSubstring("Successfully added: good_user"))
			})

			It("deduplicates contacts and sends only one POST per unique username", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 0
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "user1"),
						sdktesting.RespondWithJSON(http.StatusCreated, `{
							"kind": "Account", "id": "acc-1", "username": "user1"
						}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1", "user1", "user1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})

			It("skips removals when an addition fails", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-old", "username": "old_user"}
							],
							"size": 1,
							"total": 1
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusBadRequest,
							`{"kind":"Error","id":"400","reason":"invalid username"}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"bad_user"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to add 1 of 1"))
				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})

			It("treats POST 409 as success when contact already exists", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 0
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "user1"),
						sdktesting.RespondWithJSON(http.StatusCreated, `{
							"kind": "Account", "id": "acc-1", "username": "user1"
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "user1@example.com"),
						sdktesting.RespondWithJSON(http.StatusConflict,
							`{"kind":"Error","id":"409","reason":"notification contact already exists"}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"}
							],
							"size": 1,
							"total": 1
						}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1", "user1@example.com"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not remove existing contact when 409 alias resolves to it", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-alice", "username": "alice"}
							],
							"size": 1,
							"total": 1
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "alice@example.com"),
						sdktesting.RespondWithJSON(http.StatusConflict,
							`{"kind":"Error","id":"409","reason":"notification contact already exists"}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-alice", "username": "alice"}
							],
							"size": 1,
							"total": 1
						}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"alice@example.com"})
				Expect(err).NotTo(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(3))
			})

			It("returns error when 409 re-fetch fails", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [],
							"size": 0,
							"total": 0
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "alias@example.com"),
						sdktesting.RespondWithJSON(http.StatusConflict,
							`{"kind":"Error","id":"409","reason":"notification contact already exists"}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusForbidden,
							`{"kind":"Error","id":"403","reason":"forbidden"}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"alias@example.com"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can't resolve canonical username"))
			})

			It("returns error when 409 cannot resolve canonical among multiple unseen contacts", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-alice", "username": "alice"},
								{"kind": "Account", "id": "acc-zoe", "username": "zoe"}
							],
							"size": 2,
							"total": 2
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "zoe@example.com"),
						sdktesting.RespondWithJSON(http.StatusConflict,
							`{"kind":"Error","id":"409","reason":"notification contact already exists"}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-alice", "username": "alice"},
								{"kind": "Account", "id": "acc-zoe", "username": "zoe"}
							],
							"size": 2,
							"total": 2
						}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"zoe@example.com"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can't resolve canonical username"))
				Expect(err.Error()).To(ContainSubstring("multiple unresolved contacts"))
				Expect(server.ReceivedRequests()).To(HaveLen(3))
			})

			It("includes username in removal error messages", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"}
							],
							"size": 1,
							"total": 1
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-1"),
						sdktesting.RespondWithJSON(http.StatusForbidden,
							`{"kind":"Error","id":"403","reason":"forbidden"}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("user1"))
				Expect(err.Error()).To(ContainSubstring("acc-1"))
			})

			It("returns an error when the DELETE fails", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"}
							],
							"size": 1,
							"total": 1
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-1"),
						sdktesting.RespondWithJSON(http.StatusForbidden,
							`{"kind":"Error","id":"403","reason":"forbidden"}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can't remove notification contact"))
			})

			It("reports mutation when one DELETE succeeds and a later DELETE fails", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"},
								{"kind": "Account", "id": "acc-2", "username": "user2"}
							],
							"size": 2,
							"total": 2
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", MatchRegexp(`/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-`)),
						sdktesting.RespondWithJSON(http.StatusNoContent, ""),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", MatchRegexp(`/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-`)),
						sdktesting.RespondWithJSON(http.StatusForbidden,
							`{"kind":"Error","id":"403","reason":"forbidden"}`),
					),
				)

				mutated, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{})
				Expect(err).To(HaveOccurred())
				Expect(mutated).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("can't remove notification contact"))
				Expect(server.ReceivedRequests()).To(HaveLen(3))
			})

			It("treats DELETE 404 as success when contact already removed", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-1", "username": "user1"}
							],
							"size": 1,
							"total": 1
						}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-1"),
						sdktesting.RespondWithJSON(http.StatusNotFound,
							`{"kind":"Error","id":"404","reason":"not found"}`),
					),
				)

				_, err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{})
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})
})
