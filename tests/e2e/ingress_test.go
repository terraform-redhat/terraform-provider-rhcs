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

var _ = Describe("Ingress updating", func() {

	var err error
	var profile *ci.Profile
	var ingress *cmv1.Ingress
	var ingressService *exec.IngressService

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
		ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		ingressService, err = exec.NewIngressService(con.ClassicIngressDir)
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		// For future implementation
	})
	It("allow ingress LB config via terraform provider - [id:70336]",
		ci.Day2,
		ci.High,
		func() {
			By("update the LB type to classic")
			args := exec.IngressArgs{
				LoadBalancerType: "classic",
				Cluster:          clusterID,
			}
			err = ingressService.Apply(&args)
			if profile.GetClusterType().HCP {
				By("prepare one ROSA HCP cluster, export cluster ID then apply")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(MatchRegexp(`Can't update load balancer type on[\s\S]?Hosted Control Plane cluster '%s'`, clusterID))
				return
			}
			Expect(err).ToNot(HaveOccurred())

			defer func() {
				args := exec.IngressArgs{
					LoadBalancerType: string(ingress.LoadBalancerType()),
					Cluster:          clusterID,
				}
				ingressService.Apply(&args)
			}()

			By("use API to check if ingress LB type updated")
			ingress, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(ingress.LoadBalancerType())).To(Equal("classic"))

			By("update the LB type to back to NLB")
			args = exec.IngressArgs{
				LoadBalancerType: "nlb",
				Cluster:          clusterID,
			}
			err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.LoadBalancerType())).To(Equal("nlb"))
		})
	It("allow ingress day2 configuration in the RHCS terraform provider - [id:70337]",
		ci.Day2,
		ci.High,
		func() {
			By("update ingress attributes")

			args := exec.IngressArgs{
				ExcludedNamespaces: []string{
					"qe",
					"test"},
				Cluster: clusterID,
				RouteSelectors: map[string]string{
					"route": "internal",
				},
				RouteNamespaceOwnershipPolicy: "Strict",
				RouteWildcardPolicy:           "WildcardsAllowed",
			}
			err = ingressService.Apply(&args)
			if profile.GetClusterType().HCP {
				By("prepare one ROSA HCP cluster, export cluster ID then apply")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(MatchRegexp(`Can't update route selectors on[\s\S]?Hosted Control Plane cluster '%s'`, clusterID))
				return
			}
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				args := exec.IngressArgs{
					ExcludedNamespaces:            ingress.ExcludedNamespaces(),
					Cluster:                       clusterID,
					RouteSelectors:                ingress.RouteSelectors(),
					RouteNamespaceOwnershipPolicy: string(ingress.RouteNamespaceOwnershipPolicy()),
					RouteWildcardPolicy:           string(ingress.RouteWildcardPolicy()),
				}
				ingressService.Apply(&args)
			}()
			By("use ocm API to check if ingress config updated")
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.RouteNamespaceOwnershipPolicy())).To(Equal(args.RouteNamespaceOwnershipPolicy))
			Expect(string(ingress.RouteWildcardPolicy())).To(Equal(args.RouteWildcardPolicy))
			Expect(ingress.RouteSelectors()["route"]).To(Equal(args.RouteSelectors["route"]))
			Expect(ingress.ExcludedNamespaces()).To(Equal(args.ExcludedNamespaces))

			By("just update one of cluster_routes_tls_secret_ref and cluster_routes_hostname, not update both together.")
			args = exec.IngressArgs{
				ClusterRoutesHostename: "test.example.com",
				Cluster:                clusterID,
			}
			err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("must be set together"))

		})
})
