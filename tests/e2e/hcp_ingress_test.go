package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

var internalListeningMethod = "internal"
var externalListeningMethod = "external"

var _ = Describe("HCP Ingress", ci.NonClassicCluster, ci.FeatureIngress, ci.Day2, func() {

	var err error
	var ingressBefore *cmv1.Ingress
	var ingressService *exec.IngressService

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()

		ingressBefore, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		ingressService, err = exec.NewIngressService(constants.HCPIngressDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		listeningMethod := string(ingressBefore.Listening())
		args := exec.IngressArgs{
			Cluster:         &clusterID,
			ListeningMethod: &listeningMethod,
		}
		err = ingressService.Apply(&args)
		Expect(err).ToNot(HaveOccurred())

		err = ingressService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})

	It("can be edited - [id:72517]",
		ci.High,
		func() {
			By("Set Listening method to internal")
			args := exec.IngressArgs{
				Cluster:         &clusterID,
				ListeningMethod: &internalListeningMethod,
			}
			err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("Verify Cluster Ingress")
			ingress, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.Listening())).To(Equal("internal"))

			By("Set Listening method to external")
			args = exec.IngressArgs{
				Cluster:         &clusterID,
				ListeningMethod: &externalListeningMethod,
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

	It("validate edit - [id:72520]", ci.Medium, func() {
		By("Initialize ingress state")
		listeningMethod := string(ingressBefore.Listening())
		args := exec.IngressArgs{
			Cluster:         &clusterID,
			ListeningMethod: &listeningMethod,
		}
		err = ingressService.Apply(&args)
		Expect(err).ToNot(HaveOccurred())

		By("Try to edit with empty cluster")
		args = exec.IngressArgs{
			Cluster:         helper.EmptyStringPointer,
			ListeningMethod: &internalListeningMethod,
		}
		err = ingressService.Apply(&args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Attribute cluster cluster ID may not be empty/blank string"))

		By("Try to edit cluster with other cluster ID")
		clustersResp, err := cms.ListClusters(ci.RHCSConnection)
		Expect(err).ToNot(HaveOccurred())
		var otherClusterID string
		for _, cluster := range clustersResp.Items().Slice() {
			if cluster.ID() != clusterID && cluster.Status().State() == cmv1.ClusterStateReady {
				otherClusterID = cluster.ID()
				break
			}
		}
		if otherClusterID != "" {
			args = exec.IngressArgs{
				Cluster:         &otherClusterID,
				ListeningMethod: &internalListeningMethod,
			}
			err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute cluster, cannot be changed from"))
		} else {
			Logger.Info("No other cluster accessible for testing this change")
		}

		By("Try to edit cluster field with wrong value")
		value := "wrong"
		args = exec.IngressArgs{
			Cluster:         &value,
			ListeningMethod: &internalListeningMethod,
		}
		err = ingressService.Apply(&args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Cluster 'wrong' not"))

		By("Try to edit with empty listening_method")
		args = exec.IngressArgs{
			Cluster:         &clusterID,
			ListeningMethod: helper.EmptyStringPointer,
		}
		err = ingressService.Apply(&args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Expected a valid param"))
		Expect(err.Error()).To(ContainSubstring("Options are"))

		By("Try to edit with wrong listening_method")
		value = "wrong"
		args = exec.IngressArgs{
			Cluster:         &clusterID,
			ListeningMethod: &value,
		}
		err = ingressService.Apply(&args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Expected a valid param"))
		Expect(err.Error()).To(ContainSubstring("Options are"))
	})
})
