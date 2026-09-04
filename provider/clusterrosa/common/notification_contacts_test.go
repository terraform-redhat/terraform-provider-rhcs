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
		})

		Context("UpdateNotificationContacts", func() {
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

				err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1", "user2"})
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

				err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1"})
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

				err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("handles mixed add and remove", func() {
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
						ghttp.VerifyRequest("DELETE", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts/acc-1"),
						sdktesting.RespondWithJSON(http.StatusNoContent, ""),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.VerifyJQ(".account_identifier", "user3"),
						sdktesting.RespondWithJSON(http.StatusCreated, `{
							"kind": "Account", "id": "acc-3", "username": "user3"
						}`),
					),
				)

				err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user2", "user3"})
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

				err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"user1"})
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

				err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{"nonexistent"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can't add notification contact"))
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

				err := UpdateNotificationContacts(ctx, connection, "sub-123", []string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can't remove notification contact"))
			})
		})

		Context("ResolveNotificationContacts", func() {
			It("returns contacts from the subscription", func() {
				cluster, err := cmv1.NewCluster().
					Subscription(cmv1.NewSubscription().ID("sub-123")).
					Build()
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"),
						sdktesting.RespondWithJSON(http.StatusOK, `{
							"kind": "AccountList",
							"items": [
								{"kind": "Account", "id": "acc-a", "username": "alice"},
								{"kind": "Account", "id": "acc-b", "username": "bob"}
							],
							"size": 2,
							"total": 2
						}`),
					),
				)

				setVal, diags := ResolveNotificationContacts(ctx, connection, cluster)
				Expect(diags).To(BeEmpty())
				Expect(setVal.IsNull()).To(BeFalse())
				Expect(setVal.Elements()).To(HaveLen(2))
			})

			It("returns a warning when subscription ID is not available", func() {
				cluster, err := cmv1.NewCluster().Build()
				Expect(err).NotTo(HaveOccurred())

				setVal, diags := ResolveNotificationContacts(ctx, connection, cluster)
				Expect(diags).To(HaveLen(1))
				Expect(diags[0].Detail()).To(ContainSubstring("subscription ID is not available"))
				Expect(setVal.IsNull()).To(BeTrue())
			})

			It("returns a warning when the API call fails", func() {
				cluster, err := cmv1.NewCluster().
					Subscription(cmv1.NewSubscription().ID("sub-123")).
					Build()
				Expect(err).NotTo(HaveOccurred())

				server.RouteToHandler("GET",
					"/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts",
					sdktesting.RespondWithJSON(http.StatusNotFound,
						`{"kind":"Error","id":"404","href":"/api/accounts_mgmt/v1/errors/404","code":"AMS-404","reason":"not found"}`),
				)

				setVal, diags := ResolveNotificationContacts(ctx, connection, cluster)
				Expect(diags).To(HaveLen(1))
				Expect(diags[0].Detail()).To(ContainSubstring("Could not read notification contacts"))
				Expect(setVal.IsNull()).To(BeTrue())
			})

			It("returns empty set when no contacts are set", func() {
				cluster, err := cmv1.NewCluster().
					Subscription(cmv1.NewSubscription().ID("sub-123")).
					Build()
				Expect(err).NotTo(HaveOccurred())

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

				setVal, diags := ResolveNotificationContacts(ctx, connection, cluster)
				Expect(diags).To(BeEmpty())
				Expect(setVal.IsNull()).To(BeFalse())
				Expect(setVal.Elements()).To(BeEmpty())
			})
		})
	})
})
