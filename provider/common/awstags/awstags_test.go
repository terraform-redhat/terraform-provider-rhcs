// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package awstags

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"
)

func ptr(value string) *string {
	return &value
}

var _ = Describe("AWS tags", func() {
	Context("ValidateVersion", func() {
		It("accepts 4.20 and later control planes", func() {
			Expect(ValidateVersion("4.20.0")).To(Succeed())
			Expect(ValidateVersion("4.20.0-ec.1")).To(Succeed())
			Expect(ValidateVersion("4.21.3")).To(Succeed())
			Expect(ValidateVersion("openshift-v4.20.1")).To(Succeed())
		})

		It("rejects older or unknown control planes", func() {
			Expect(ValidateVersion("4.19.12")).To(MatchError(ContainSubstring("requires OpenShift version 4.20.0")))
			Expect(ValidateVersion("")).To(MatchError(ContainSubstring("without a control-plane version")))
			Expect(ValidateVersion("not-a-version")).To(MatchError(ContainSubstring("cannot check control-plane version")))
		})
	})

	Context("Validate", func() {
		It("accepts tags with spaces", func() {
			Expect(Validate(map[string]string{"Cost Center": "Finance Team"})).To(Succeed())
			Expect(Validate(nil)).To(Succeed())
		})

		It("rejects empty values", func() {
			Expect(Validate(map[string]string{"team": ""})).To(MatchError(ContainSubstring("cannot be empty")))
		})

		It("rejects reserved keys", func() {
			Expect(Validate(map[string]string{"api.openshift.com/id": "x"})).To(MatchError(ContainSubstring("reserved")))
			Expect(Validate(map[string]string{"red-hat-managed": "x"})).NotTo(Succeed())
			Expect(Validate(map[string]string{"aws:foo": "x"})).NotTo(Succeed())
		})
	})

	Context("Diff", func() {
		It("returns nil for equal sets, treating nil and empty as equal", func() {
			Expect(Diff(nil, map[string]string{})).To(BeNil())
			Expect(Diff(map[string]string{"a": "1"}, map[string]string{"a": "1"})).To(BeNil())
		})

		It("adds, updates, and removes keys with null values", func() {
			changes := Diff(
				map[string]string{"keep": "1", "update": "old", "remove": "x"},
				map[string]string{"keep": "1", "update": "new", "add": "y"},
			)
			Expect(changes).To(HaveLen(3))
			Expect(changes).To(HaveKeyWithValue("update", ptr("new")))
			Expect(changes).To(HaveKeyWithValue("add", ptr("y")))
			Expect(changes).To(HaveKeyWithValue("remove", BeNil()))
		})
	})

	Context("PoolChanges", func() {
		clusterTags := map[string]string{"team": "platform", "env": "prod"}

		It("stores only overrides that differ from cluster tags", func() {
			changes := PoolChanges(clusterTags,
				map[string]string{"team": "platform", "env": "prod"},
				map[string]string{"team": "platform", "pool": "gpu"})
			Expect(changes).To(Equal(map[string]*string{
				"pool": ptr("gpu"),
				// Matching inherited keys are nulled in case an equal-value override is stored.
				"team": nil,
				"env":  nil,
			}))
		})

		It("removes an override to restore the inherited cluster value", func() {
			changes := PoolChanges(clusterTags,
				map[string]string{"team": "data", "env": "prod"},
				map[string]string{})
			Expect(changes).To(HaveKeyWithValue("team", BeNil()))
			Expect(changes).To(HaveKeyWithValue("env", BeNil()))
		})

		It("does not null inherited keys whose effective value is overridden and kept", func() {
			changes := PoolChanges(clusterTags,
				map[string]string{"team": "data", "env": "prod"},
				map[string]string{"team": "data", "env": "staging"})
			Expect(changes).To(Equal(map[string]*string{"env": ptr("staging")}))
		})

		It("reports desired keys that will be inherited", func() {
			Expect(InheritedMatches(clusterTags, map[string]string{"team": "platform", "env": "dev", "x": "y"})).
				To(Equal([]string{"team"}))
		})
	})

	Context("OCM transport", func() {
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

		It("reads customer-only tags and ignores null tombstones", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/c1", "fetchUserTagsOnly=true"),
					sdktesting.RespondWithJSON(http.StatusOK, `{"aws":{"tags":{"team":"platform"}}}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/c1/node_pools/p1",
						"fetchUserTagsOnly=true"),
					sdktesting.RespondWithJSON(http.StatusOK, `{"aws_node_pool":{"tags":{"team":"platform","old":null}}}`),
				),
			)
			clusterTags, err := GetClusterUserTags(ctx, connection, "c1")
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterTags).To(Equal(map[string]string{"team": "platform"}))
			poolTags, err := GetNodePoolUserTags(ctx, connection, "c1", "p1")
			Expect(err).NotTo(HaveOccurred())
			Expect(poolTags).To(Equal(map[string]string{"team": "platform"}))
		})

		It("returns empty tags when the resource has none", func() {
			server.AppendHandlers(sdktesting.RespondWithJSON(http.StatusOK, `{"id":"c1"}`))
			tags, err := GetClusterUserTags(ctx, connection, "c1")
			Expect(err).NotTo(HaveOccurred())
			Expect(tags).To(BeEmpty())
		})

		It("combines nullable tag changes with a typed cluster edit", func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/c1", "fetchUserTagsOnly=true"),
				func(_ http.ResponseWriter, request *http.Request) {
					body, err := io.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					var payload struct {
						AWS struct {
							BillingAccountID string             `json:"billing_account_id"`
							Tags             map[string]*string `json:"tags"`
						} `json:"aws"`
					}
					Expect(json.Unmarshal(body, &payload)).To(Succeed())
					Expect(payload.AWS.BillingAccountID).To(Equal("billing"))
					Expect(payload.AWS.Tags).To(HaveKeyWithValue("remove", BeNil()))
					Expect(payload.AWS.Tags).To(HaveKeyWithValue("Cost Center", ptr("Finance Team")))
				},
				sdktesting.RespondWithJSON(http.StatusOK, `{"id":"c1"}`),
			))
			cluster, err := cmv1.NewCluster().AWS(cmv1.NewAWS().BillingAccountID("billing")).Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(PatchCluster(ctx, connection, "c1", cluster,
				map[string]*string{"remove": nil, "Cost Center": ptr("Finance Team")})).To(Succeed())
		})

		It("adds tag changes to a node pool edit without an AWS section", func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/c1/node_pools/p1"),
				func(_ http.ResponseWriter, request *http.Request) {
					body, err := io.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					var payload struct {
						Replicas    int `json:"replicas"`
						AWSNodePool struct {
							Tags map[string]*string `json:"tags"`
						} `json:"aws_node_pool"`
					}
					Expect(json.Unmarshal(body, &payload)).To(Succeed())
					Expect(payload.Replicas).To(Equal(3))
					Expect(payload.AWSNodePool.Tags).To(HaveKeyWithValue("remove", BeNil()))
				},
				sdktesting.RespondWithJSON(http.StatusOK, `{"id":"p1"}`),
			))
			nodePool, err := cmv1.NewNodePool().ID("p1").Replicas(3).Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(PatchNodePool(ctx, connection, "c1", nodePool, map[string]*string{"remove": nil})).To(Succeed())
		})

		It("reports the API validation reason when a PATCH fails", func() {
			server.AppendHandlers(sdktesting.RespondWithJSON(http.StatusBadRequest,
				`{"kind":"Error","reason":"reserved tag 'red-hat-managed'"}`))
			cluster, err := cmv1.NewCluster().Build()
			Expect(err).NotTo(HaveOccurred())
			err = PatchCluster(ctx, connection, "c1", cluster, map[string]*string{"a": ptr("b")})
			Expect(err).To(MatchError(ContainSubstring("status is 400, reserved tag 'red-hat-managed'")))
		})

		It("reports a failed tag read", func() {
			server.AppendHandlers(sdktesting.RespondWithJSON(http.StatusNotFound, `{"kind":"Error","reason":"not found"}`))
			_, err := GetClusterUserTags(ctx, connection, "c1")
			Expect(err).To(MatchError(ContainSubstring("failed to read AWS tags: status is 404, not found")))
		})
	})
})
