package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("HCP Ingress", ci.NonClassicCluster, ci.FeatureIngress, func() {

	var err error
	var ingressBefore *cmv1.Ingress
	var ingressService *exec.IngressService

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()

		ingressBefore, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		ingressService, err = exec.NewIngressService(con.HCPIngressDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		args := exec.IngressArgs{
			Cluster:         clusterID,
			ListeningMethod: string(ingressBefore.Listening()),
		}
		ingressService.Apply(&args)
	})

	It("can be edited - [id:72517]",
		ci.Day2,
		ci.High,
		func() {
			By("Set Listening method to internal")
			args := exec.IngressArgs{
				Cluster:         clusterID,
				ListeningMethod: "internal",
			}
			err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("Verify Cluster Ingress")
			ingress, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.Listening())).To(Equal("internal"))

			By("Set Listening method to external")
			args = exec.IngressArgs{
				Cluster:         clusterID,
				ListeningMethod: "external",
			}
			err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("Verify Cluster Ingress")
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.Listening())).To(Equal("external"))

			By("Destroy Cluster Ingress")
			err = ingressService.Destroy()
			Expect(err).ToNot(HaveOccurred())

			By("Verify Cluster Ingress is still present")
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.Listening())).To(Equal("external"))
		})
})
