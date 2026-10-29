// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	sprest "github.com/openshift-online/rosa-seaplane-api/clientset/rest"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var seaplaneRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

var _ = Describe("DNS Domain", func() {
	var (
		dnsService     exec.DnsDomainService
		profileHandler profilehandler.ProfileHandler
	)
	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		// Use a dedicated temp workspace so shared-vpc day1 PrepareRoute53 DNS
		// (profile workspace) is not destroyed while the cluster still exists.
		// Same isolation pattern as id:67574 in account_roles_test.go.
		tempWorkspace := helper.GenerateRandomName("ocp-67570"+profileHandler.Profile().GetName(), 2)
		Logger.Infof("Using temp workspace '%s' for creating resources", tempWorkspace)

		dnsService, err = exec.NewDnsDomainService(tempWorkspace, profileHandler.Profile().GetClusterType())
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dnsService.Destroy()
	})

	It("can create and destroy dnsdomain - [id:67570]",
		ci.Day2, ci.Medium, ci.FeatureIDP, ci.SeaplaneValidated, func() {
			if profileHandler.Profile().GetClusterType() == constants.SEAPLANE {
				Skip("Seaplane DNS reservations are covered by the cluster workflow test")
			}

			By("Create/Apply dns-domain resource by terraform")
			_, err := dnsService.Apply(&exec.DnsDomainArgs{})
			Expect(err).ToNot(HaveOccurred())

			By("Destroy dns-domain resource by terraform")
			_, err = dnsService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})
})

var _ = Describe("Seaplane DNS reservation", ci.SeaplaneValidated, func() {
	It("binds a reservation UID to a cluster with an explicitly empty domain prefix", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profileHandler.Profile().GetClusterType() != constants.SEAPLANE {
			Skip("DNS reservation cluster binding requires a Seaplane profile")
		}

		seaplaneURL := os.Getenv("SEAPLANE_URL")
		Expect(seaplaneURL).NotTo(BeEmpty(), "SEAPLANE_URL is required for the Seaplane DNS reservation test")
		region := seaplaneRegionRE.FindString(seaplaneURL)
		if region == "" {
			region = profileHandler.Profile().GetRegion()
		}
		Expect(region).NotTo(BeEmpty(), "cannot derive AWS region from SEAPLANE_URL")

		workspace := helper.GenerateRandomName("sp-dns-"+profileHandler.Profile().GetName(), 2)
		dnsService, err := exec.NewDnsDomainService(workspace+"-reservation", constants.SEAPLANE)
		Expect(err).NotTo(HaveOccurred())

		reservationName := helper.GenerateRandomName("dns-"+profileHandler.Profile().GetName(), 2)
		_, err = dnsService.Apply(&exec.DnsDomainArgs{
			Name:        &reservationName,
			SeaplaneURL: &seaplaneURL,
			AWSRegion:   &region,
		})
		Expect(err).NotTo(HaveOccurred())
		reservation, err := dnsService.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(reservation.DnsDomainId).NotTo(BeEmpty())
		Expect(reservation.BaseDomain).NotTo(BeEmpty())
		Expect(reservation.Phase).To(Equal("Ready"))

		var destroyer clusterworkflow.ClusterDestroyer
		DeferCleanup(func() {
			if destroyer != nil {
				Expect(destroyer.Destroy(context.Background())).To(Succeed())
			}
			_, destroyErr := dnsService.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())
		})

		emptyPrefix := ""
		backend, err := clusterworkflow.NewSeaplaneWithOptions(
			profileHandler,
			workspace,
			clusterworkflow.SeaplaneOptions{
				DNSReservationID: reservation.DnsDomainId,
				BaseDomainPrefix: &emptyPrefix,
			},
		)
		Expect(err).NotTo(HaveOccurred())
		destroyer, err = clusterworkflow.NewDestroyer(profileHandler, "", workspace)
		Expect(err).NotTo(HaveOccurred())

		clusterID, err := exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterID).NotTo(BeEmpty())
		state, err := backend.Lifecycle.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.ClusterName).NotTo(BeEmpty())

		client := newDNSTestSeaplaneClient(seaplaneURL, region)
		Eventually(func(g Gomega) {
			cluster, getErr := client.SeaplaneV1alpha1().Clusters().Get(
				ctx,
				state.ClusterName,
				spplatform.GetOptions{},
			)
			g.Expect(getErr).NotTo(HaveOccurred())
			g.Expect(cluster.Spec.DNSReservationID).To(Equal(reservation.DnsDomainId))
			g.Expect(cluster.Status.BaseDomain).To(Equal(reservation.BaseDomain))
			g.Expect(cluster.Spec.HostedCluster.DNS.BaseDomainPrefix).NotTo(BeNil())
			g.Expect(*cluster.Spec.HostedCluster.DNS.BaseDomainPrefix).To(BeEmpty())
		}).WithContext(ctx).WithTimeout(5 * time.Minute).WithPolling(15 * time.Second).Should(Succeed())
	})
})

func newDNSTestSeaplaneClient(seaplaneURL, region string) *seaplane.Clientset {
	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	Expect(err).NotTo(HaveOccurred())
	identity, err := sts.NewFromConfig(awsCfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	Expect(err).NotTo(HaveOccurred())
	client, err := seaplane.NewForConfig(&sprest.Config{
		Host:      seaplaneURL,
		Region:    region,
		AccountID: *identity.Account,
		CallerARN: *identity.Arn,
		AWSConfig: awsCfg,
	})
	Expect(err).NotTo(HaveOccurred())
	return client
}
