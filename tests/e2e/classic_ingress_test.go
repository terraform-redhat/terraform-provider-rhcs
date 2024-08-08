package e2e

import (

	// nolint

	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmsv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("Classic Ingress", ci.FeatureIngress, func() {

	var (
		profile        *ci.Profile
		err            error
		ingressBefore  *cmsv1.Ingress
		ingressService exec.IngressService
	)

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()

		if profile.GetClusterType().HCP {
			Skip("Test can run only on Classic cluster")
		}

		ingressBefore, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		ingressService, err = exec.NewIngressService(constants.ClassicIngressDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		var componentRoutes *map[string]*exec.IngressComponentRoute
		if ingressBefore.ComponentRoutes() == nil {
			componentRoutes = nil
		} else {
			crs := map[string]*exec.IngressComponentRoute{}

			if ingressBefore.ComponentRoutes()["oauth"] != nil {
				crs["oauth"] = exec.NewIngressComponentRoute(
					helper.StringPointer(ingressBefore.ComponentRoutes()["oauth"].Hostname()),
					helper.StringPointer(ingressBefore.ComponentRoutes()["oauth"].TlsSecretRef()),
				)
			}
			if ingressBefore.ComponentRoutes()["console"] != nil {
				crs["console"] = exec.NewIngressComponentRoute(
					helper.StringPointer(ingressBefore.ComponentRoutes()["console"].Hostname()),
					helper.StringPointer(ingressBefore.ComponentRoutes()["console"].TlsSecretRef()),
				)
			}
			if ingressBefore.ComponentRoutes()["downloads"] != nil {
				crs["downloads"] = exec.NewIngressComponentRoute(
					helper.StringPointer(ingressBefore.ComponentRoutes()["downloads"].Hostname()),
					helper.StringPointer(ingressBefore.ComponentRoutes()["downloads"].TlsSecretRef()),
				)
			}
			componentRoutes = &crs
		}
		args := exec.IngressArgs{
			Cluster:                       helper.StringPointer(clusterID),
			ExcludedNamespaces:            helper.StringSlicePointer(ingressBefore.ExcludedNamespaces()),
			LoadBalancerType:              helper.StringPointer(string(ingressBefore.LoadBalancerType())),
			RouteSelectors:                helper.StringMapPointer(ingressBefore.RouteSelectors()),
			RouteNamespaceOwnershipPolicy: helper.StringPointer(string(ingressBefore.RouteNamespaceOwnershipPolicy())),
			RouteWildcardPolicy:           helper.StringPointer(string(ingressBefore.RouteWildcardPolicy())),
			ComponentRoutes:               componentRoutes,
		}
		ingressService.Apply(&args)
	})

	It("allows LB configuration - [id:70336]",
		ci.Day2,
		ci.High,
		func() {
			By("update the LB type to classic")
			args := exec.IngressArgs{
				LoadBalancerType: helper.StringPointer("classic"),
				Cluster:          helper.StringPointer(clusterID),
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("use API to check if ingress LB type updated")
			ingress, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(ingress.LoadBalancerType())).To(Equal("classic"))

			By("update the LB type to back to NLB")
			args = exec.IngressArgs{
				LoadBalancerType: helper.StringPointer("nlb"),
				Cluster:          helper.StringPointer(clusterID),
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.LoadBalancerType())).To(Equal("nlb"))
		})

	It("update component routes - [id:73610]",
		ci.Day2,
		ci.High,
		func() {
			By("set ingress component routes")
			componentRoutes := map[string]*exec.IngressComponentRoute{
				"oauth": exec.NewIngressComponentRoute(
					helper.StringPointer("oauth.example.com"),
					helper.StringPointer("oauth"),
				),
				"console": exec.NewIngressComponentRoute(
					helper.StringPointer("console.example.com"),
					helper.StringPointer("console"),
				),
				"downloads": exec.NewIngressComponentRoute(
					helper.StringPointer("downloads.example.com"),
					helper.StringPointer("downloads"),
				),
			}
			args := exec.IngressArgs{
				ComponentRoutes: &componentRoutes,
				Cluster:         &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("use ocm API to check if component routes updated")
			ingress, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(ingress.ComponentRoutes()["oauth"].Hostname()).To(Equal(*componentRoutes["oauth"].Hostname))
			Expect(ingress.ComponentRoutes()["oauth"].TlsSecretRef()).To(Equal(*componentRoutes["oauth"].TlsSecretRef))
			Expect(ingress.ComponentRoutes()["console"].Hostname()).To(Equal(*componentRoutes["console"].Hostname))
			Expect(ingress.ComponentRoutes()["console"].TlsSecretRef()).To(Equal(*componentRoutes["console"].TlsSecretRef))
			Expect(ingress.ComponentRoutes()["downloads"].Hostname()).To(Equal(*componentRoutes["downloads"].Hostname))
			Expect(ingress.ComponentRoutes()["downloads"].TlsSecretRef()).To(Equal(*componentRoutes["downloads"].TlsSecretRef))

			By("update ingress component routes")
			componentRoutes = map[string]*exec.IngressComponentRoute{
				"oauth": exec.NewIngressComponentRoute(
					helper.StringPointer("oauth.test.example.com"),
					helper.StringPointer("test-oauth"),
				),
				"console": exec.NewIngressComponentRoute(
					helper.StringPointer("console.test.example.com"),
					helper.StringPointer("test-console"),
				),
				"downloads": exec.NewIngressComponentRoute(
					helper.StringPointer("downloads.test.example.com"),
					helper.StringPointer("test-downloads"),
				),
			}
			args = exec.IngressArgs{
				ComponentRoutes: &componentRoutes,
				Cluster:         &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("use ocm API to check if component routes updated")
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(ingress.ComponentRoutes()["oauth"].Hostname()).To(Equal(*componentRoutes["oauth"].Hostname))
			Expect(ingress.ComponentRoutes()["oauth"].TlsSecretRef()).To(Equal(*componentRoutes["oauth"].TlsSecretRef))
			Expect(ingress.ComponentRoutes()["console"].Hostname()).To(Equal(*componentRoutes["console"].Hostname))
			Expect(ingress.ComponentRoutes()["console"].TlsSecretRef()).To(Equal(*componentRoutes["console"].TlsSecretRef))
			Expect(ingress.ComponentRoutes()["downloads"].Hostname()).To(Equal(*componentRoutes["downloads"].Hostname))
			Expect(ingress.ComponentRoutes()["downloads"].TlsSecretRef()).To(Equal(*componentRoutes["downloads"].TlsSecretRef))

			By("remove some component routes")
			componentRoutes = map[string]*exec.IngressComponentRoute{
				"oauth": exec.NewIngressComponentRoute(
					helper.StringPointer("oauth.test.example.com"),
					helper.StringPointer("test-oauth"),
				),
			}
			args = exec.IngressArgs{
				ComponentRoutes: &componentRoutes,
				Cluster:         &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("use ocm API to check if component routes updated")
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(ingress.ComponentRoutes()["oauth"].Hostname()).To(Equal(*componentRoutes["oauth"].Hostname))
			Expect(ingress.ComponentRoutes()["oauth"].TlsSecretRef()).To(Equal(*componentRoutes["oauth"].TlsSecretRef))
			Expect(ingress.ComponentRoutes()["console"]).To(BeNil())
			Expect(ingress.ComponentRoutes()["downloads"]).To(BeNil())

			By("Remove component routes")
			args = exec.IngressArgs{
				ComponentRoutes: nil,
				Cluster:         &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("use ocm API to check if component routes deleted")
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(ingress.ComponentRoutes())).To(Equal(0))
		})

	It("validate ingress components - [id:75067]",
		ci.Day2,
		ci.Medium,
		func() {
			By("Try to remove only the hostname")
			out, err := ingressService.Output()
			Expect(err).ToNot(HaveOccurred())
			args := exec.IngressArgs{
				ComponentRoutes: &map[string]*exec.IngressComponentRoute{
					"oauth": exec.NewIngressComponentRoute(
						helper.StringPointer(""),
						helper.StringPointer("test-oauth"),
					),
					"console": exec.NewIngressComponentRoute(
						helper.StringPointer("console.test.example.com"),
						helper.StringPointer("test-console"),
					),
					"downloads": exec.NewIngressComponentRoute(
						helper.StringPointer("downloads.test.example.com"),
						helper.StringPointer("test-downloads"),
					),
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("Can't remove 'oauth' component route hostname for ingress '%s' in cluster '%s' without also removing TLS secret reference", out.ID, clusterID))

			By("Try to remove only the tls_secret_ref")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]*exec.IngressComponentRoute{
					"oauth": exec.NewIngressComponentRoute(
						helper.StringPointer("oauth.test.example.com"),
						helper.StringPointer(""),
					),
					"console": exec.NewIngressComponentRoute(
						helper.StringPointer("console.test.example.com"),
						helper.StringPointer("test-console"),
					),
					"downloads": exec.NewIngressComponentRoute(
						helper.StringPointer("downloads.test.example.com"),
						helper.StringPointer("test-downloads"),
					),
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("Can't update 'oauth' component route hostname for ingress '%s' in cluster '%s' without also supplying a new TLS secret reference", out.ID, clusterID))

			By("Try with blank component routes")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]*exec.IngressComponentRoute{},
				Cluster:         &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Error: Provider produced inconsistent result after apply")

			By("Try with invalid cluster ID")
			invalidID := "asdf"
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]*exec.IngressComponentRoute{
					"oauth": exec.NewIngressComponentRoute(
						helper.StringPointer("oauth.test.example.com"),
						helper.StringPointer("test-oauth"),
					),
					"console": exec.NewIngressComponentRoute(
						helper.StringPointer("console.test.example.com"),
						helper.StringPointer("test-console"),
					),
					"downloads": exec.NewIngressComponentRoute(
						helper.StringPointer("downloads.test.example.com"),
						helper.StringPointer("test-downloads"),
					),
				},
				Cluster: &invalidID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("Cluster '%s' not found", invalidID))
		})

	It("allows day2 configuration - [id:70337]",
		ci.Day2,
		ci.High,
		func() {
			By("update ingress attributes")
			args := exec.IngressArgs{
				ExcludedNamespaces: helper.StringSlicePointer([]string{
					"qe",
					"test"}),
				Cluster: &clusterID,
				RouteSelectors: helper.StringMapPointer(map[string]string{
					"route": "internal",
				}),
				RouteNamespaceOwnershipPolicy: helper.StringPointer("Strict"),
				RouteWildcardPolicy:           helper.StringPointer("WildcardsAllowed"),
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("use ocm API to check if ingress config updated")
			ingress, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ingress.RouteNamespaceOwnershipPolicy())).To(Equal(*args.RouteNamespaceOwnershipPolicy))
			Expect(string(ingress.RouteWildcardPolicy())).To(Equal(*args.RouteWildcardPolicy))
			Expect(ingress.RouteSelectors()["route"]).To(Equal((*args.RouteSelectors)["route"]))
			Expect(ingress.ExcludedNamespaces()).To(Equal(*args.ExcludedNamespaces))

			By("just update one of cluster_routes_tls_secret_ref and cluster_routes_hostname, not update both together.")
			args = exec.IngressArgs{
				ClusterRoutesHostename: helper.StringPointer("test.example.com"),
				Cluster:                helper.StringPointer(clusterID),
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("must be set together"))

		})
})
