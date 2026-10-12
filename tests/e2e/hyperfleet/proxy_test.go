// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hfplatform "github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("Hyperfleet proxy", func() {
	It("creates, updates, and clears proxy settings on a retained cluster", func(ctx SpecContext) {
		hyperfleetURL := requireEnv("HYPERFLEET_URL")
		clusterName := requireEnv("HYPERFLEET_CLUSTER_NAME")
		awsRegion := awsRegionRE.FindString(hyperfleetURL)
		Expect(awsRegion).NotTo(BeEmpty(), fmt.Sprintf("cannot derive AWS region from HYPERFLEET_URL %q", hyperfleetURL))
		workspace := clusterName

		clusterSvc, err := exec.NewHyperfleetClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())
		clusterOutput, err := clusterSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterOutput.ClusterID).NotTo(BeEmpty(), fmt.Sprintf("proxy checks require existing cluster Terraform state in workspace %q", workspace))
		clusterID := clusterOutput.ClusterID
		clusterArgs, err := clusterSvc.ReadTFVars()
		Expect(err).NotTo(HaveOccurred())
		originalArgs := *clusterArgs
		runProxyValidationChecks := strings.EqualFold(os.Getenv("HYPERFLEET_PROXY_VALIDATION_TEST"), "true")
		invalidHTTPProxy := "http://proxy.example.com/" + strings.Repeat("x", 2049)

		vpcSvc, err := exec.NewHyperfleetVPCService(workspace)
		Expect(err).NotTo(HaveOccurred())
		vpcOutput, err := vpcSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(vpcOutput.VPCID).NotTo(BeEmpty())
		Expect(vpcOutput.PublicSubnetID).NotTo(BeEmpty())

		proxyWorkspace := workspace + "-proxy-sanity"
		proxySvc, err := exec.NewProxyService(proxyWorkspace, constants.HYPERFLEET)
		Expect(err).NotTo(HaveOccurred())
		proxyTrustBundlePath := filepath.Join(config.GetRHCSOutputDir(), proxyWorkspace+"-ca.cert")
		var proxyAttempted bool
		var clusterProxyChangeAttempted bool

		hfClient := mustBuildHyperfleetClient(hyperfleetURL, awsRegion)
		clusters := hfClient.HyperfleetV1alpha1().Clusters()
		currentCluster, err := clusters.Get(ctx, clusterID, hfplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		originalArgs.HTTPProxy = nil
		originalArgs.HTTPSProxy = nil
		originalArgs.NoProxy = nil
		if currentCluster.Spec.HostedCluster.Configuration != nil && currentCluster.Spec.HostedCluster.Configuration.Proxy != nil {
			httpProxy := currentCluster.Spec.HostedCluster.Configuration.Proxy.HTTPProxy
			httpsProxy := currentCluster.Spec.HostedCluster.Configuration.Proxy.HTTPSProxy
			noProxy := currentCluster.Spec.HostedCluster.Configuration.Proxy.NoProxy
			originalArgs.HTTPProxy = &httpProxy
			originalArgs.HTTPSProxy = &httpsProxy
			originalArgs.NoProxy = &noProxy
		}
		verifyClusterProxy := func(waitCtx context.Context, wantHTTP, wantHTTPS, wantNo string) {
			Eventually(func() bool {
				cluster, getErr := clusters.Get(waitCtx, clusterID, hfplatform.GetOptions{})
				if getErr != nil {
					return false
				}
				configuration := cluster.Spec.HostedCluster.Configuration
				if wantHTTP == "" && wantHTTPS == "" && wantNo == "" {
					return configuration == nil || configuration.Proxy == nil
				}
				if configuration == nil || configuration.Proxy == nil {
					return false
				}
				proxy := configuration.Proxy
				return proxy.HTTPProxy == wantHTTP && proxy.HTTPSProxy == wantHTTPS && proxy.NoProxy == wantNo
			}).WithContext(waitCtx).WithTimeout(5 * time.Minute).WithPolling(pollInterval).Should(BeTrue())
		}
		proxyString := func(value *string) string {
			if value == nil {
				return ""
			}
			return *value
		}
		DeferCleanup(func(cleanupCtx SpecContext) {
			if !proxyAttempted {
				return
			}
			if clusterProxyChangeAttempted {
				By("Restoring the retained cluster's original proxy settings")
				if _, restoreErr := clusterSvc.Apply(&originalArgs); restoreErr != nil {
					Fail(fmt.Sprintf("restoring the cluster proxy settings failed; leaving proxy endpoints in place: %v", restoreErr))
					return
				}
				verifyClusterProxy(cleanupCtx, proxyString(originalArgs.HTTPProxy), proxyString(originalArgs.HTTPSProxy), proxyString(originalArgs.NoProxy))
			}
			By("Destroying the proxy endpoints created by this spec")
			_, destroyErr := proxySvc.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())
		}, GracePeriod(teardownGracePeriod))

		if runProxyValidationChecks {
			By("[72491] Verifying an over-limit proxy is rejected during cluster creation")
			invalidWorkspace := fmt.Sprintf("%s-invalid-proxy-%d", workspace, time.Now().UnixNano())
			invalidProxyClusterSvc, createErr := exec.NewHyperfleetClusterService(invalidWorkspace)
			Expect(createErr).NotTo(HaveOccurred())
			invalidClusterName := fmt.Sprintf("ip-%012d", time.Now().UnixNano()%1_000_000_000_000)
			invalidExpiration := time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339)
			invalidArgs := originalArgs
			invalidArgs.ClusterName = &invalidClusterName
			invalidArgs.ExpirationTimestamp = &invalidExpiration
			invalidArgs.HTTPProxy = &invalidHTTPProxy
			invalidArgs.HTTPSProxy = nil
			invalidArgs.NoProxy = nil
			// A configured OidcConfig can back only one cluster. Leave it unset here
			// so the API validates the intentionally over-limit proxy value instead
			// of rejecting the request because the retained cluster already uses it.
			invalidArgs.OIDCConfigID = nil
			_, invalidCreateErr := invalidProxyClusterSvc.Apply(&invalidArgs)
			if invalidCreateErr == nil {
				unexpectedCluster, outputErr := invalidProxyClusterSvc.Output()
				_, destroyErr := invalidProxyClusterSvc.Destroy()
				Expect(destroyErr).NotTo(HaveOccurred())
				if outputErr == nil && unexpectedCluster.ClusterID != "" {
					deleteErr := clusters.WaitUntil(ctx, unexpectedCluster.ClusterID,
						func(cluster *v1alpha1.Cluster) bool { return cluster == nil }, pollInterval, clusterDeletedTimeout)
					Expect(deleteErr).NotTo(HaveOccurred())
				}
				Fail(fmt.Sprintf("Hyperfleet accepted an HTTP proxy longer than the 2048-character limit (output error: %v)", outputErr))
			}
			Expect(invalidCreateErr).To(HaveOccurred())
			invalidCreateMessage := strings.ToLower(invalidCreateErr.Error())
			Expect(invalidCreateMessage).To(Or(
				ContainSubstring("2048"),
				ContainSubstring("too long"),
				ContainSubstring("length"),
			), "cluster creation should fail because the HTTP proxy exceeds the API limit")
			_, destroyErr := invalidProxyClusterSvc.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())
		}

		By("Creating two proxy endpoints in the retained cluster VPC")
		proxyCount := 2
		proxyAttempted = true // Ensure partially created AWS resources are cleaned up on failure.
		region := awsRegion
		_, err = proxySvc.Apply(&exec.ProxyArgs{
			ProxyCount:          &proxyCount,
			Region:              &region,
			VPCID:               &vpcOutput.VPCID,
			PublicSubnetID:      &vpcOutput.PublicSubnetID,
			TrustBundleFilePath: &proxyTrustBundlePath,
		})
		Expect(err).NotTo(HaveOccurred())
		proxies, err := proxySvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(proxies.Proxies).To(HaveLen(2))
		initialProxy := proxies.Proxies[0]
		updatedProxy := proxies.Proxies[1]
		Expect(initialProxy.HttpProxy).NotTo(BeEmpty())
		Expect(initialProxy.HttpsProxy).NotTo(BeEmpty())
		Expect(updatedProxy.HttpProxy).NotTo(BeEmpty())
		Expect(updatedProxy.HttpsProxy).NotTo(BeEmpty())

		By("[72489] Applying and verifying the first proxy endpoint")
		updatedArgs := originalArgs
		updatedArgs.HTTPProxy = &initialProxy.HttpProxy
		updatedArgs.HTTPSProxy = &initialProxy.HttpsProxy
		updatedArgs.NoProxy = &initialProxy.NoProxy
		clusterProxyChangeAttempted = true
		_, err = clusterSvc.Apply(&updatedArgs)
		Expect(err).NotTo(HaveOccurred())
		verifyClusterProxy(ctx, initialProxy.HttpProxy, initialProxy.HttpsProxy, initialProxy.NoProxy)

		if runProxyValidationChecks {
			By("[72490] Rejecting an over-limit proxy update without changing stored settings")
			invalidEditArgs := updatedArgs
			invalidEditArgs.HTTPProxy = &invalidHTTPProxy
			_, err = clusterSvc.Apply(&invalidEditArgs)
			Expect(err).To(HaveOccurred())
			verifyClusterProxy(ctx, initialProxy.HttpProxy, initialProxy.HttpsProxy, initialProxy.NoProxy)
		}

		By("[72489] Updating proxy endpoints and no_proxy")
		updatedNoProxy := updatedProxy.NoProxy + ",example.com"
		updatedArgs.HTTPProxy = &updatedProxy.HttpProxy
		updatedArgs.HTTPSProxy = &updatedProxy.HttpsProxy
		updatedArgs.NoProxy = &updatedNoProxy
		_, err = clusterSvc.Apply(&updatedArgs)
		Expect(err).NotTo(HaveOccurred())
		verifyClusterProxy(ctx, updatedProxy.HttpProxy, updatedProxy.HttpsProxy, updatedNoProxy)

		By("Clearing proxy settings from the retained cluster")
		updatedArgs.HTTPProxy = nil
		updatedArgs.HTTPSProxy = nil
		updatedArgs.NoProxy = nil
		_, err = clusterSvc.Apply(&updatedArgs)
		Expect(err).NotTo(HaveOccurred())
		verifyClusterProxy(ctx, "", "", "")
	})
})
