// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"
)

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
})

const (
	waiterTestClusterPath = "/api/clusters_mgmt/v1/clusters/123"
	waiterTestAPIURL      = "https://api.my-cluster.example.com:6443"
	waiterTestConsoleURL  = "https://console.my-cluster.example.com"
)

// waiterTestCluster builds a cluster JSON body; empty URLs are omitted, as OCM does
// before they are assigned.
func waiterTestCluster(state, apiURL, consoleURL string) string {
	body := fmt.Sprintf(`{"kind": "Cluster", "id": "123", "state": %q`, state)
	if apiURL != "" {
		body += fmt.Sprintf(`, "api": {"url": %q}`, apiURL)
	}
	if consoleURL != "" {
		body += fmt.Sprintf(`, "console": {"url": %q}`, consoleURL)
	}
	return body + "}"
}

func waiterTestClusterHandler(state, apiURL, consoleURL string) http.HandlerFunc {
	return ghttp.CombineHandlers(
		ghttp.VerifyRequest(http.MethodGet, waiterTestClusterPath),
		ghttp.RespondWith(http.StatusOK, waiterTestCluster(state, apiURL, consoleURL),
			http.Header{"Content-Type": []string{"application/json"}}),
	)
}

var _ = Describe("Cluster readiness waiter", func() {
	var (
		server     *ghttp.Server
		connection *sdk.Connection
		waiter     ClusterWait
	)

	BeforeEach(func() {
		var ca string
		server, ca = sdktesting.MakeTCPTLSServer()
		token := sdktesting.MakeTokenString("Bearer", 10*time.Minute)
		var err error
		connection, err = sdk.NewConnectionBuilder().
			URL(server.URL()).
			TrustedCAFile(ca).
			Tokens(token).
			BuildContext(context.Background())
		Expect(err).NotTo(HaveOccurred())
		waiter = NewClusterWait(connection.ClustersMgmt().V1().Clusters(), connection)
	})

	AfterEach(func() {
		server.Close()
		connection.Close()
	})

	It("WaitForClusterToBeReady does not require URLs", func() {
		server.AppendHandlers(waiterTestClusterHandler("ready", "", ""))

		cluster, err := waiter.WaitForClusterToBeReady(context.Background(), "123", 60)

		Expect(err).NotTo(HaveOccurred())
		Expect(cluster.State()).To(Equal(cmv1.ClusterStateReady))
		Expect(cluster.API().URL()).To(BeEmpty())
		Expect(server.ReceivedRequests()).To(HaveLen(1))
	})

	It("WaitForClusterToBeReady accepts a polled ready cluster without URLs", func() {
		server.AppendHandlers(
			waiterTestClusterHandler("installing", "", ""),
			waiterTestClusterHandler("ready", "", ""),
		)

		cluster, err := waiter.WaitForClusterToBeReady(context.Background(), "123", 60)

		Expect(err).NotTo(HaveOccurred())
		Expect(cluster.State()).To(Equal(cmv1.ClusterStateReady))
		Expect(server.ReceivedRequests()).To(HaveLen(2))
	})

	Context("WaitForClusterToBeReadyWithURLs", func() {
		It("returns immediately when the cluster is ready and reports both URLs", func() {
			server.AppendHandlers(waiterTestClusterHandler("ready", waiterTestAPIURL, waiterTestConsoleURL))

			cluster, err := waiter.WaitForClusterToBeReadyWithURLs(context.Background(), "123", 60)

			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.API().URL()).To(Equal(waiterTestAPIURL))
			Expect(cluster.Console().URL()).To(Equal(waiterTestConsoleURL))
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		DescribeTable("keeps polling a ready cluster until both URLs are reported",
			func(apiURL, consoleURL string) {
				server.AppendHandlers(
					waiterTestClusterHandler("ready", apiURL, consoleURL),
					waiterTestClusterHandler("ready", waiterTestAPIURL, waiterTestConsoleURL),
				)

				cluster, err := waiter.WaitForClusterToBeReadyWithURLs(context.Background(), "123", 60)

				Expect(err).NotTo(HaveOccurred())
				Expect(cluster.API().URL()).To(Equal(waiterTestAPIURL))
				Expect(cluster.Console().URL()).To(Equal(waiterTestConsoleURL))
				Expect(server.ReceivedRequests()).To(HaveLen(2))
			},
			Entry("API URL missing", "", waiterTestConsoleURL),
			Entry("console URL missing", waiterTestAPIURL, ""),
			Entry("both URLs missing", "", ""),
		)

		It("waits for a cluster that becomes ready with both URLs", func() {
			server.AppendHandlers(
				waiterTestClusterHandler("installing", "", ""),
				waiterTestClusterHandler("ready", waiterTestAPIURL, waiterTestConsoleURL),
			)

			cluster, err := waiter.WaitForClusterToBeReadyWithURLs(context.Background(), "123", 60)

			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.State()).To(Equal(cmv1.ClusterStateReady))
			Expect(cluster.Console().URL()).To(Equal(waiterTestConsoleURL))
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})

		It("returns a terminal error state without waiting for URLs", func() {
			server.AppendHandlers(waiterTestClusterHandler("error", "", ""))

			cluster, err := waiter.WaitForClusterToBeReadyWithURLs(context.Background(), "123", 60)

			Expect(err).To(MatchError(ContainSubstring("Cluster '123' is in state 'error' and will not become ready")))
			Expect(cluster).NotTo(BeNil())
			Expect(cluster.State()).To(Equal(cmv1.ClusterStateError))
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("stops polling when the cluster reaches a terminal state while URLs are missing", func() {
			// The 60 minute timeout leaves room for further 2 minute polling intervals, so
			// a fast return proves the terminal state ended the wait.
			server.AppendHandlers(
				waiterTestClusterHandler("installing", "", ""),
				waiterTestClusterHandler("error", "", ""),
			)

			started := time.Now()
			cluster, err := waiter.WaitForClusterToBeReadyWithURLs(context.Background(), "123", 60)

			Expect(err).To(MatchError("cluster '123' is in state 'error'"))
			Expect(cluster).NotTo(BeNil())
			Expect(cluster.State()).To(Equal(cmv1.ClusterStateError))
			Expect(server.ReceivedRequests()).To(HaveLen(2))
			Expect(time.Since(started)).To(BeNumerically("<", 30*time.Second))
		})

		It("fails when polling ends while the URLs are still missing", func() {
			// With a 1 minute timeout the 2 minute polling interval does not fit, so polling
			// ends after its first request, exactly as it does in WaitForClusterToBeReady.
			server.AppendHandlers(
				waiterTestClusterHandler("ready", "", ""),
				waiterTestClusterHandler("ready", "", ""),
			)

			cluster, err := waiter.WaitForClusterToBeReadyWithURLs(context.Background(), "123", 1)

			Expect(err).To(MatchError("cluster '123' is in state 'ready'"))
			Expect(cluster).NotTo(BeNil())
			Expect(cluster.API().URL()).To(BeEmpty())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})

		It("reports a cluster that is still not ready when polling ends", func() {
			server.AppendHandlers(
				waiterTestClusterHandler("installing", "", ""),
				waiterTestClusterHandler("installing", "", ""),
			)

			cluster, err := waiter.WaitForClusterToBeReadyWithURLs(context.Background(), "123", 1)

			Expect(err).To(MatchError("cluster '123' is in state 'installing'"))
			Expect(cluster).NotTo(BeNil())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})

		It("stops when the context is cancelled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			started := time.Now()
			cluster, err := waiter.WaitForClusterToBeReadyWithURLs(ctx, "123", 60)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, context.Canceled)).To(BeTrue())
			Expect(cluster).To(BeNil())
			Expect(time.Since(started)).To(BeNumerically("<", 5*time.Second))
		})
	})
})
