// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmsv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var internalListeningMethod = "internal"
var externalListeningMethod = "external"

var _ = Describe("HCP Ingress", ci.FeatureIngress, ci.Day2, func() {

	var (
		profileHandler profilehandler.ProfileHandler
		ingressBefore  *cmsv1.Ingress
		ingressService exec.IngressService
		ingressArgs    *exec.IngressArgs
	)

	getOriginalIngressArgs := func() *exec.IngressArgs {
		args, err := ingressService.ReadTFVars()
		Expect(err).ToNot(HaveOccurred())
		if args.Cluster == nil {
			args.Cluster = new(clusterID)
		}
		return args
	}

	BeforeEach(func() {
		var err error

		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if !profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Hosted cluster")
		}

		ingressBefore, err = cms.RetrieveClusterIngress(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		ingressService, err = profileHandler.Services().GetIngressService()
		Expect(err).ToNot(HaveOccurred())

		ingressArgs = getOriginalIngressArgs()
	})

	AfterEach(func() {
		ingressArgs.ListeningMethod = new(string(ingressBefore.Listening()))
		_, err := ingressService.Apply(ingressArgs)
		Expect(err).ToNot(HaveOccurred())
	})

	It("can be edited - [id:72517]",
		ci.High,
		func() {
			By("Set Listening method to internal")
			ingressArgs.ListeningMethod = new(internalListeningMethod)
			_, err := ingressService.Apply(ingressArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify Cluster Ingress")
			ingress, err := cms.RetrieveClusterIngress(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.Listening())).To(Equal("internal"))

			By("Set Listening method to external")
			ingressArgs.ListeningMethod = new(externalListeningMethod)
			_, err = ingressService.Apply(ingressArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify Cluster Ingress")
			ingress, err = cms.RetrieveClusterIngress(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.Listening())).To(Equal("external"))

			By("Destroy Cluster Ingress")
			_, err = ingressService.Destroy()
			Expect(err).ToNot(HaveOccurred())

			By("Verify Cluster Ingress is still present")
			ingress, err = cms.RetrieveClusterIngress(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.Listening())).To(Equal("external"))
		})

	It("validate edit - [id:72520]", ci.Medium, func() {
		By("Initialize ingress state")
		ingressArgs.ListeningMethod = new(string(ingressBefore.Listening()))
		_, err := ingressService.Apply(ingressArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Try to edit with empty cluster")
		ingressArgs.Cluster = helper.EmptyStringPointer
		_, err = ingressService.Apply(ingressArgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Attribute cluster cluster ID may not be empty/blank string"))

		By("Try to edit cluster with other cluster ID")
		clustersResp, err := cms.ListClusters(cms.RHCSConnection)
		Expect(err).ToNot(HaveOccurred())
		var otherClusterID string
		for _, cluster := range clustersResp.Items().Slice() {
			if cluster.ID() != clusterID && cluster.Status().State() == cmsv1.ClusterStateReady {
				otherClusterID = cluster.ID()
				break
			}
		}
		if otherClusterID != "" {
			ingressArgs = getOriginalIngressArgs()
			ingressArgs.Cluster = new(otherClusterID)
			_, err = ingressService.Apply(ingressArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute cluster, cannot be changed from"))
		} else {
			Logger.Info("No other cluster accessible for testing this change")
		}

		By("Try to edit cluster field with wrong value")
		ingressArgs = getOriginalIngressArgs()
		ingressArgs.Cluster = new("wrong")
		_, err = ingressService.Apply(ingressArgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cannot be changed"))

		By("Try to edit with empty listening_method")
		ingressArgs = getOriginalIngressArgs()
		ingressArgs.ListeningMethod = helper.EmptyStringPointer
		_, err = ingressService.Apply(ingressArgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Expected a valid param"))
		Expect(err.Error()).To(ContainSubstring("Options are"))

		By("Try to edit with wrong listening_method")
		ingressArgs = getOriginalIngressArgs()
		ingressArgs.ListeningMethod = new("wrong")
		_, err = ingressService.Apply(ingressArgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Expected a valid param"))
		Expect(err.Error()).To(ContainSubstring("Options are"))
	})
})

var _ = Describe("Shared component routes workflow", ci.Day2, ci.FeatureIngress, func() {
	It("updates component routes through the selected HCP or HyperFleet backend", ci.High, func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		clusterType := profile.Profile().GetClusterType()
		if clusterType != constants.ROSA_HCP && clusterType != constants.HYPERFLEET {
			Skip("component routes are covered here only for HCP and HyperFleet")
		}

		backend, err := clusterworkflow.NewExisting(profile)
		Expect(err).NotTo(HaveOccurred())
		before, err := backend.Lifecycle.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		originalRoutes := before.ComponentRoutes
		if originalRoutes == "" {
			if clusterType == constants.HYPERFLEET {
				originalRoutes = "[]"
			} else {
				originalRoutes = "{}"
			}
		}
		DeferCleanup(func() {
			Expect(backend.Lifecycle.Update(exec.ClusterSpec{ComponentRoutes: &originalRoutes})).To(Succeed())
		})

		routes := `[{"namespace":"openshift-console","name":"console","hostname":"console-shared.example.com","servingCertKeyPairSecret":{"name":"console-tls"}}]`
		Expect(backend.Lifecycle.Update(exec.ClusterSpec{ComponentRoutes: &routes})).To(Succeed())
		Eventually(func() bool {
			state, readErr := backend.Lifecycle.Read(ctx)
			return readErr == nil && state.ComponentRoutes != "" &&
				strings.Contains(state.ComponentRoutes, "console-shared.example.com")
		}).WithContext(ctx).WithTimeout(5 * time.Minute).WithPolling(15 * time.Second).Should(BeTrue())
	})
})
