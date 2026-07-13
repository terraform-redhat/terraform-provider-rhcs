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

var _ = Describe("Delete protection helpers", func() {
	Context("ValidateDeleteAllowed", func() {
		It("returns no diagnostics when delete protection is disabled", func() {
			diags := ValidateDeleteAllowed("123", false)
			Expect(diags.HasError()).To(BeFalse())
		})

		It("returns an error when delete protection is enabled", func() {
			diags := ValidateDeleteAllowed("123", true)
			Expect(diags.HasError()).To(BeTrue())
			Expect(diags.Errors()[0].Detail()).To(ContainSubstring("delete_protection = false"))
		})
	})

	Context("Sub-resource operations", func() {
		var (
			server        *ghttp.Server
			ca            string
			connection    *sdk.Connection
			clusterClient *cmv1.ClusterClient
			ctx           context.Context
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
			clusterClient = connection.ClustersMgmt().V1().Clusters().Cluster("123")
		})

		AfterEach(func() {
			server.Close()
			connection.Close()
		})

		Context("FetchDeleteProtection", func() {
			It("returns true when the sub-resource reports enabled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/clusters_mgmt/v1/clusters/123/delete_protection"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"enabled": true,
						}),
					),
				)
				enabled, err := FetchDeleteProtection(ctx, clusterClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(enabled).To(BeTrue())
			})

			It("returns false when the sub-resource reports disabled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/clusters_mgmt/v1/clusters/123/delete_protection"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"enabled": false,
						}),
					),
				)
				enabled, err := FetchDeleteProtection(ctx, clusterClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(enabled).To(BeFalse())
			})

			It("returns an error when the API call fails", func() {
				server.RouteToHandler("GET",
					"/api/clusters_mgmt/v1/clusters/123/delete_protection",
					sdktesting.RespondWithJSON(http.StatusNotFound, `{"kind":"Error","id":"404","href":"/api/clusters_mgmt/v1/errors/404","code":"CLUSTERS-MGMT-404","reason":"not found"}`),
				)
				_, err := FetchDeleteProtection(ctx, clusterClient)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("UpdateDeleteProtection", func() {
			It("sends a PATCH with the enabled value", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/clusters_mgmt/v1/clusters/123/delete_protection"),
						sdktesting.VerifyJQ(".enabled", true),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"enabled": true,
						}),
					),
				)
				err := UpdateDeleteProtection(ctx, clusterClient, true)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error when the API call fails", func() {
				server.RouteToHandler("PATCH",
					"/api/clusters_mgmt/v1/clusters/123/delete_protection",
					sdktesting.RespondWithJSON(http.StatusNotFound, `{"kind":"Error","id":"404","href":"/api/clusters_mgmt/v1/errors/404","code":"CLUSTERS-MGMT-404","reason":"not found"}`),
				)
				err := UpdateDeleteProtection(ctx, clusterClient, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can't update cluster delete protection"))
			})
		})

		Context("ResolveDeleteProtection", func() {
			It("returns the inline value when present", func() {
				cluster, err := cmv1.NewCluster().
					DeleteProtection(cmv1.NewDeleteProtection().Enabled(true)).
					Build()
				Expect(err).NotTo(HaveOccurred())

				val, diags := ResolveDeleteProtection(ctx, clusterClient, cluster)
				Expect(diags).To(BeEmpty())
				Expect(val.ValueBool()).To(BeTrue())
			})

			It("falls back to the sub-resource when inline is absent", func() {
				cluster, err := cmv1.NewCluster().Build()
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/clusters_mgmt/v1/clusters/123/delete_protection"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"enabled": true,
						}),
					),
				)

				val, diags := ResolveDeleteProtection(ctx, clusterClient, cluster)
				Expect(diags).To(BeEmpty())
				Expect(val.ValueBool()).To(BeTrue())
			})

			It("returns a warning when the sub-resource call fails", func() {
				cluster, err := cmv1.NewCluster().Build()
				Expect(err).NotTo(HaveOccurred())

				server.RouteToHandler("GET",
					"/api/clusters_mgmt/v1/clusters/123/delete_protection",
					sdktesting.RespondWithJSON(http.StatusNotFound, `{"kind":"Error","id":"404","href":"/api/clusters_mgmt/v1/errors/404","code":"CLUSTERS-MGMT-404","reason":"not found"}`),
				)

				val, diags := ResolveDeleteProtection(ctx, clusterClient, cluster)
				Expect(diags).To(HaveLen(1))
				Expect(diags[0].Detail()).To(ContainSubstring("Could not read delete protection"))
				Expect(val.ValueBool()).To(BeFalse())
			})
		})

		Context("CheckDeleteProtectionEnabled", func() {
			It("returns the sub-resource value when available", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/clusters_mgmt/v1/clusters/123/delete_protection"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"enabled": true,
						}),
					),
				)
				enabled, diags := CheckDeleteProtectionEnabled(ctx, "123", clusterClient)
				Expect(diags).To(BeEmpty())
				Expect(enabled).To(BeTrue())
			})

			It("falls back to the cluster GET when the sub-resource fails", func() {
				clusterJSON := `{
					"kind": "Cluster",
					"id": "123",
					"name": "test-cluster",
					"delete_protection": {"enabled": true}
				}`
				server.RouteToHandler("GET",
					"/api/clusters_mgmt/v1/clusters/123/delete_protection",
					sdktesting.RespondWithJSON(http.StatusNotFound, `{"kind":"Error","id":"404","href":"/api/clusters_mgmt/v1/errors/404","code":"CLUSTERS-MGMT-404","reason":"not found"}`),
				)
				server.RouteToHandler("GET",
					"/api/clusters_mgmt/v1/clusters/123",
					sdktesting.RespondWithJSON(http.StatusOK, clusterJSON),
				)
				enabled, diags := CheckDeleteProtectionEnabled(ctx, "123", clusterClient)
				Expect(diags).To(BeEmpty())
				Expect(enabled).To(BeTrue())
			})

			It("returns an error when both sub-resource and cluster GET fail", func() {
				server.RouteToHandler("GET",
					"/api/clusters_mgmt/v1/clusters/123/delete_protection",
					sdktesting.RespondWithJSON(http.StatusNotFound, `{"kind":"Error","id":"404","href":"/api/clusters_mgmt/v1/errors/404","code":"CLUSTERS-MGMT-404","reason":"not found"}`),
				)
				server.RouteToHandler("GET",
					"/api/clusters_mgmt/v1/clusters/123",
					sdktesting.RespondWithJSON(http.StatusNotFound, `{"kind":"Error","id":"404","href":"/api/clusters_mgmt/v1/errors/404","code":"CLUSTERS-MGMT-404","reason":"cluster error"}`),
				)
				_, diags := CheckDeleteProtectionEnabled(ctx, "123", clusterClient)
				Expect(diags.HasError()).To(BeTrue())
				Expect(diags.Errors()[0].Detail()).To(ContainSubstring("cluster GET fallback failed"))
			})
		})
	})
})
