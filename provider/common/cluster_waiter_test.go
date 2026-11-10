// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"
)

const clusterReadyNoURLs = `{"id":"123","name":"my-cluster","state":"ready"}`

const clusterReadyWithURLs = `{"id":"123","name":"my-cluster","state":"ready",` +
	`"api":{"url":"https://my-api.example.com"},` +
	`"console":{"url":"https://my-console.example.com"}}`

var _ = Describe("Cluster waiter", func() {
	var waiter *DefaultClusterWait

	BeforeEach(func() {
		token := sdktesting.MakeTokenString("Bearer", 10*time.Minute)
		connection, err := sdk.NewConnectionBuilder().Tokens(token).BuildContext(context.Background())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(connection.Close)
		waiter = &DefaultClusterWait{connection: connection}
	})

	It("uses one deadline for polling", func() {
		started := time.Now()
		cluster, err := waiter.pollClusterWithRetry(
			context.Background(), "123", 50*time.Millisecond,
			func(ctx context.Context) (*cmv1.Cluster, error) {
				<-ctx.Done()
				return nil, ctx.Err()
			},
		)

		Expect(cluster).To(BeNil())
		Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
		Expect(time.Since(started)).To(BeNumerically("<", 500*time.Millisecond))
	})

	It("stops retry backoff at the deadline", func() {
		started := time.Now()
		cluster, err := waiter.pollClusterWithRetry(
			context.Background(), "123", 50*time.Millisecond,
			func(context.Context) (*cmv1.Cluster, error) {
				return nil, errors.New("temporary error")
			},
		)

		Expect(cluster).To(BeNil())
		Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
		Expect(time.Since(started)).To(BeNumerically("<", 500*time.Millisecond))
	})

	Context("WaitForClusterEndpoints", func() {
		var (
			server     *ghttp.Server
			connection *sdk.Connection
		)

		BeforeEach(func() {
			var ca string
			server, ca = sdktesting.MakeTCPTLSServer()
			DeferCleanup(server.Close)
			token := sdktesting.MakeTokenString("Bearer", 10*time.Minute)
			var err error
			connection, err = sdk.NewConnectionBuilder().
				URL(server.URL()).
				TrustedCAFile(ca).
				Tokens(token).
				BuildContext(context.Background())
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(connection.Close)
		})

		It("returns the cluster when endpoint URLs are already populated", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					sdktesting.RespondWithJSON(http.StatusOK, clusterReadyWithURLs),
				),
			)

			cluster, err := WaitForClusterEndpoints(
				context.Background(), connection.ClustersMgmt().V1().Clusters(), "123", time.Minute,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.API().URL()).To(Equal("https://my-api.example.com"))
			Expect(cluster.Console().URL()).To(Equal("https://my-console.example.com"))
		})

		It("re-fetches the cluster until endpoint URLs are populated", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					sdktesting.RespondWithJSON(http.StatusOK, clusterReadyNoURLs),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					sdktesting.RespondWithJSON(http.StatusOK, clusterReadyWithURLs),
				),
			)

			cluster, err := WaitForClusterEndpoints(
				context.Background(), connection.ClustersMgmt().V1().Clusters(), "123", time.Minute,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.API().URL()).To(Equal("https://my-api.example.com"))
			Expect(cluster.Console().URL()).To(Equal("https://my-console.example.com"))
		})

		It("times out with an error when endpoint URLs are never populated", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					sdktesting.RespondWithJSON(http.StatusOK, clusterReadyNoURLs),
				),
			)

			started := time.Now()
			_, err := WaitForClusterEndpoints(
				context.Background(), connection.ClustersMgmt().V1().Clusters(), "123", 50*time.Millisecond,
			)

			Expect(err).To(MatchError(ContainSubstring("api_url/console_url")))
			Expect(time.Since(started)).To(BeNumerically("<", 500*time.Millisecond))
		})
	})
})
