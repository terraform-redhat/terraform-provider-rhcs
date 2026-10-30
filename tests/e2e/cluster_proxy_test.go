// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	sprest "github.com/openshift-online/rosa-seaplane-api/clientset/rest"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var sharedClusterProxyAWSRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

const (
	sharedClusterProxyPollInterval          = 30 * time.Second
	sharedClusterProxyClusterDeletedTimeout = 60 * time.Minute
	sharedClusterProxyTeardownGracePeriod   = 90 * time.Minute
)

var _ = Describe("Shared cluster proxy workflow", ci.Day2, ci.FeatureClusterProxy, func() {
	It("proxy - [id:72489]", ci.High, func() {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profile.Profile().GetClusterType() == constants.SEAPLANE {
			Skip("Seaplane proxy coverage runs through the Seaplane-validated workflow")
		}
		if clusterID == "" {
			Skip("an existing cluster ID is required for the proxy day-2 workflow")
		}
		if !profile.Profile().IsProxy() {
			Skip("No proxy configured")
		}
		runClassicClusterProxyWorkflow(profile)
	})

	It("creates, updates, and clears proxy settings on a retained Seaplane cluster - [id:seaplane-proxy-lifecycle]",
		ci.High, ci.SeaplaneValidated, func(ctx SpecContext) {
			profile, err := profilehandler.NewProfileHandlerFromYamlFile()
			Expect(err).NotTo(HaveOccurred())
			if profile.Profile().GetClusterType() != constants.SEAPLANE {
				Skip("Test requires a Seaplane profile")
			}
			if clusterID == "" {
				Skip("an existing cluster ID is required for the proxy day-2 workflow")
			}
			if !profile.Profile().IsProxy() {
				Skip("No proxy configured")
			}
			runSeaplaneClusterProxyWorkflow(ctx, profile)
		})
})

func runSeaplaneClusterProxyWorkflow(ctx SpecContext, profile profilehandler.ProfileHandler) {
	seaplaneURL := sharedClusterProxyRequiredEnv("SEAPLANE_URL")
	clusterName := os.Getenv("SEAPLANE_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = config.GetRHCSClusterName()
	}
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	Expect(clusterName).NotTo(BeEmpty(), "a Seaplane cluster name is required")
	awsRegion := config.GetRegion()
	if awsRegion == "" {
		awsRegion = profile.Profile().GetRegion()
	}
	if awsRegion == "" {
		awsRegion = sharedClusterProxyAWSRegionRE.FindString(seaplaneURL)
	}
	Expect(awsRegion).NotTo(BeEmpty(), "cannot determine the AWS region from the profile or SEAPLANE_URL")
	workspace := clusterName

	clusterSvc, err := exec.NewSeaplaneClusterService(workspace)
	Expect(err).NotTo(HaveOccurred())
	clusterOutput, err := clusterSvc.Output()
	Expect(err).NotTo(HaveOccurred())
	Expect(clusterOutput.ClusterID).NotTo(BeEmpty(), fmt.Sprintf("proxy checks require existing cluster Terraform state in workspace %q", workspace))
	Expect(clusterOutput.ClusterID).To(Equal(clusterID), "the cluster Terraform state must match the shared E2E fixture")
	clusterArgs, err := clusterSvc.ReadTFVars()
	Expect(err).NotTo(HaveOccurred())
	originalArgs := *clusterArgs
	Expect(clusterArgs.ClusterName).NotTo(BeNil(), "proxy checks require the cluster name for Platform API lookups")
	Expect(*clusterArgs.ClusterName).To(Equal(clusterName))
	runProxyValidationChecks := strings.EqualFold(os.Getenv("SEAPLANE_PROXY_VALIDATION_TEST"), "true")
	invalidHTTPProxy := "http://proxy.example.com/" + strings.Repeat("x", 2049)

	vpcSvc, err := exec.NewSeaplaneVPCService(workspace)
	Expect(err).NotTo(HaveOccurred())
	vpcOutput, err := vpcSvc.Output()
	Expect(err).NotTo(HaveOccurred())
	Expect(vpcOutput.VPCID).NotTo(BeEmpty())
	Expect(vpcOutput.PublicSubnetID).NotTo(BeEmpty())

	proxyWorkspace := workspace + "-proxy-sanity"
	proxySvc, err := exec.NewProxyService(proxyWorkspace, constants.SEAPLANE)
	Expect(err).NotTo(HaveOccurred())
	proxyTrustBundlePath := filepath.Join(config.GetRHCSOutputDir(), proxyWorkspace+"-ca.cert")
	var proxyAttempted bool
	var clusterProxyChangeAttempted bool

	seaplaneClient := buildSharedProxySeaplaneClient(seaplaneURL, awsRegion)
	clusters := seaplaneClient.SeaplaneV1alpha1().Clusters()
	currentCluster, err := clusters.Get(ctx, clusterName, spplatform.GetOptions{})
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
			cluster, getErr := clusters.Get(waitCtx, clusterName, spplatform.GetOptions{})
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
		}).WithContext(waitCtx).WithTimeout(5 * time.Minute).WithPolling(sharedClusterProxyPollInterval).Should(BeTrue())
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
	}, GracePeriod(sharedClusterProxyTeardownGracePeriod))

	if runProxyValidationChecks {
		By("[72491] Verifying an over-limit proxy is rejected during cluster creation")
		invalidWorkspace := fmt.Sprintf("%s-invalid-proxy-%d", workspace, time.Now().UnixNano())
		invalidProxyClusterSvc, createErr := exec.NewSeaplaneClusterService(invalidWorkspace)
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
				deleteErr := clusters.WaitUntil(ctx, invalidClusterName,
					func(cluster *v1alpha1.Cluster) bool { return cluster == nil }, sharedClusterProxyPollInterval, sharedClusterProxyClusterDeletedTimeout)
				Expect(deleteErr).NotTo(HaveOccurred())
			}
			Fail(fmt.Sprintf("Seaplane accepted an HTTP proxy longer than the 2048-character limit (output error: %v)", outputErr))
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
}

func runClassicClusterProxyWorkflow(profile profilehandler.ProfileHandler) {
	clusterService, err := profile.Services().GetClusterService()
	Expect(err).NotTo(HaveOccurred())
	clusterArgs, err := clusterService.ReadTFVars()
	Expect(err).NotTo(HaveOccurred())
	originalClusterArgs := *clusterArgs

	proxyService, err := profile.Services().GetProxyService()
	Expect(err).NotTo(HaveOccurred())
	proxyArgs, err := proxyService.ReadTFVars()
	Expect(err).NotTo(HaveOccurred())
	proxyArgs.ProxyCount = new(2)
	var proxyAttempted bool
	var clusterProxyChangeAttempted bool
	DeferCleanup(func() {
		if !proxyAttempted {
			return
		}
		if clusterProxyChangeAttempted {
			By("Restore the original cluster settings")
			_, applyErr := clusterService.Apply(&originalClusterArgs)
			Expect(applyErr).NotTo(HaveOccurred())
		}
		By("Restore the original proxy endpoint")
		proxyArgs.ProxyCount = new(1)
		_, applyErr := proxyService.Apply(proxyArgs)
		Expect(applyErr).NotTo(HaveOccurred())
	})
	proxyAttempted = true
	_, err = proxyService.Apply(proxyArgs)
	Expect(err).NotTo(HaveOccurred())

	proxiesOutput, err := proxyService.Output()
	Expect(err).NotTo(HaveOccurred())
	Expect(proxiesOutput.Proxies).To(HaveLen(2))
	newProxyOutput := proxiesOutput.Proxies[1]

	By("Apply the new proxy endpoint")
	clusterArgs.Proxy = &exec.Proxy{
		AdditionalTrustBundle: new(newProxyOutput.AdditionalTrustBundle),
		HTTPSProxy:            new(newProxyOutput.HttpsProxy),
		HTTPProxy:             new(newProxyOutput.HttpProxy),
		NoProxy:               new(newProxyOutput.NoProxy),
	}
	clusterProxyChangeAttempted = true
	_, err = clusterService.Apply(clusterArgs)
	Expect(err).NotTo(HaveOccurred())

	By("Verify the new proxy settings")
	clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
	Expect(err).NotTo(HaveOccurred())
	clusterProxy := clusterResp.Body().Proxy()
	Expect(clusterProxy.HTTPProxy()).To(Equal(newProxyOutput.HttpProxy))
	Expect(clusterProxy.HTTPSProxy()).To(Equal(newProxyOutput.HttpsProxy))
	Expect(clusterProxy.NoProxy()).To(Equal(newProxyOutput.NoProxy))

	By("Remove the proxy settings")
	clusterArgs.Proxy = &exec.Proxy{
		AdditionalTrustBundle: helper.EmptyStringPointer,
		HTTPSProxy:            helper.EmptyStringPointer,
		HTTPProxy:             helper.EmptyStringPointer,
		NoProxy:               helper.EmptyStringPointer,
	}
	_, err = clusterService.Apply(clusterArgs)
	Expect(err).NotTo(HaveOccurred())

	By("Verify the proxy settings are removed")
	clusterResp, err = cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
	Expect(err).NotTo(HaveOccurred())
	clusterProxy = clusterResp.Body().Proxy()
	Expect(clusterProxy.HTTPProxy()).To(BeEmpty())
	Expect(clusterProxy.HTTPSProxy()).To(BeEmpty())
	Expect(clusterProxy.NoProxy()).To(BeEmpty())
}

func buildSharedProxySeaplaneClient(seaplaneURL, region string) *seaplane.Clientset {
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background())
	Expect(err).NotTo(HaveOccurred())
	identity, err := sts.NewFromConfig(awsConfig).GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	Expect(err).NotTo(HaveOccurred())
	client, err := seaplane.NewForConfig(&sprest.Config{
		Host:      seaplaneURL,
		Region:    region,
		AccountID: *identity.Account,
		CallerARN: *identity.Arn,
		AWSConfig: awsConfig,
	})
	Expect(err).NotTo(HaveOccurred())
	return client
}

func sharedClusterProxyRequiredEnv(key string) string {
	value := os.Getenv(key)
	Expect(value).NotTo(BeEmpty(), fmt.Sprintf("required env var %s is not set", key))
	return value
}
