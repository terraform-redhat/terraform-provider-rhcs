// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
