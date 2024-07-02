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

		ingressBefore, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		ingressService, err = exec.NewIngressService(constants.ClassicIngressDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		var componentRoutes map[string]map[string]string
		if ingressBefore.ComponentRoutes() == nil {
			componentRoutes = nil
		} else {
			componentRoutes = map[string]map[string]string{
				"oauth": {
					"hostname":       ingressBefore.ComponentRoutes()["oauth"].Hostname(),
					"tls_secret_ref": ingressBefore.ComponentRoutes()["oauth"].TlsSecretRef(),
				},
				"console": {
					"hostname":       ingressBefore.ComponentRoutes()["console"].Hostname(),
					"tls_secret_ref": ingressBefore.ComponentRoutes()["console"].TlsSecretRef(),
				},
				"downloads": {
					"hostname":       ingressBefore.ComponentRoutes()["downloads"].Hostname(),
					"tls_secret_ref": ingressBefore.ComponentRoutes()["downloads"].TlsSecretRef(),
				},
			}
		}
		args := exec.IngressArgs{
			Cluster:                       helper.StringPointer(clusterID),
			ExcludedNamespaces:            helper.StringSlicePointer(ingressBefore.ExcludedNamespaces()),
			LoadBalancerType:              helper.StringPointer(string(ingressBefore.LoadBalancerType())),
			RouteSelectors:                helper.StringMapPointer(ingressBefore.RouteSelectors()),
			RouteNamespaceOwnershipPolicy: helper.StringPointer(string(ingressBefore.RouteNamespaceOwnershipPolicy())),
			RouteWildcardPolicy:           helper.StringPointer(string(ingressBefore.RouteWildcardPolicy())),
			ComponentRoutes:               &componentRoutes,
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
			if profile.GetClusterType().HCP {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(MatchRegexp(`Can't update load balancer type on[\s\S]?Hosted Control Plane cluster '%s'`, clusterID))
				return
			}
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

	It("update ingress components - [id:73610]",
		ci.Day2,
		ci.High,
		func() {
			By("set ingress component routes")
			args := exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "oauth.example.com",
						"tls_secret_ref": "oauth",
					},
					"console": {
						"hostname":       "console.example.com",
						"tls_secret_ref": "console",
					},
					"downloads": {
						"hostname":       "downloads.example.com",
						"tls_secret_ref": "downloads",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			if profile.GetClusterType().HCP {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(MatchRegexp(`Can't update route selectors on[\s\S]?Hosted Control Plane cluster '%s'`, clusterID))
				return
			}
			Expect(err).ToNot(HaveOccurred())

			By("use ocm API to check if ingress config updated")
			ingress, err := cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(ingress.ComponentRoutes()["oauth"].Hostname()).To(Equal((*args.ComponentRoutes)["oauth"]["hostname"]))
			Expect(ingress.ComponentRoutes()["oauth"].TlsSecretRef()).To(Equal((*args.ComponentRoutes)["oauth"]["tls_secret_ref"]))
			Expect(ingress.ComponentRoutes()["console"].Hostname()).To(Equal((*args.ComponentRoutes)["console"]["hostname"]))
			Expect(ingress.ComponentRoutes()["console"].TlsSecretRef()).To(Equal((*args.ComponentRoutes)["console"]["tls_secret_ref"]))
			Expect(ingress.ComponentRoutes()["downloads"].Hostname()).To(Equal((*args.ComponentRoutes)["downloads"]["hostname"]))
			Expect(ingress.ComponentRoutes()["downloads"].TlsSecretRef()).To(Equal((*args.ComponentRoutes)["downloads"]["tls_secret_ref"]))

			By("update ingress component routes")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "oauth.test.example.com",
						"tls_secret_ref": "test-oauth",
					},
					"console": {
						"hostname":       "console.test.example.com",
						"tls_secret_ref": "test-console",
					},
					"downloads": {
						"hostname":       "downloads.test.example.com",
						"tls_secret_ref": "test-downloads",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).ToNot(HaveOccurred())

			By("use ocm API to check if ingress config updated")
			ingress, err = cms.RetrieveClusterIngress(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(ingress.ComponentRoutes()["oauth"].Hostname()).To(Equal((*args.ComponentRoutes)["oauth"]["hostname"]))
			Expect(ingress.ComponentRoutes()["oauth"].TlsSecretRef()).To(Equal((*args.ComponentRoutes)["oauth"]["tls_secret_ref"]))
			Expect(ingress.ComponentRoutes()["console"].Hostname()).To(Equal((*args.ComponentRoutes)["console"]["hostname"]))
			Expect(ingress.ComponentRoutes()["console"].TlsSecretRef()).To(Equal((*args.ComponentRoutes)["console"]["tls_secret_ref"]))
			Expect(ingress.ComponentRoutes()["downloads"].Hostname()).To(Equal((*args.ComponentRoutes)["downloads"]["hostname"]))
			Expect(ingress.ComponentRoutes()["downloads"].TlsSecretRef()).To(Equal((*args.ComponentRoutes)["downloads"]["tls_secret_ref"]))
		})

	It("validate ingress components - [id:75067]",
		ci.Day2,
		ci.Medium,
		func() {
			By("Try to update only some sections")
			args := exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "oauth.test.example.com",
						"tls_secret_ref": "test-oauth",
					},
				},
				Cluster: &clusterID,
			}
			_, err := ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "All component route kinds must be specified. Missing [console, downloads]")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"console": {
						"hostname":       "console.test.example.com",
						"tls_secret_ref": "test-console",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "All component route kinds must be specified. Missing [oauth, downloads]")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"downloads": {
						"hostname":       "downloads.test.example.com",
						"tls_secret_ref": "test-downloads",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "All component route kinds must be specified. Missing [oauth, console]")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"console": {
						"hostname":       "console.test.example.com",
						"tls_secret_ref": "test-console",
					},
					"downloads": {
						"hostname":       "downloads.test.example.com",
						"tls_secret_ref": "test-downloads",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "All component route kinds must be specified. Missing [oauth]")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "oauth.test.example.com",
						"tls_secret_ref": "test-oauth",
					},
					"downloads": {
						"hostname":       "downloads.test.example.com",
						"tls_secret_ref": "test-downloads",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "All component route kinds must be specified. Missing [console]")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "oauth.test.example.com",
						"tls_secret_ref": "test-oauth",
					},
					"console": {
						"hostname":       "console.test.example.com",
						"tls_secret_ref": "test-console",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "All component route kinds must be specified. Missing [downloads]")

			By("Try to remove only the hostname")
			out, err := ingressService.Output()
			Expect(err).ToNot(HaveOccurred())
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "",
						"tls_secret_ref": "test-oauth",
					},
					"console": {
						"hostname":       "console.test.example.com",
						"tls_secret_ref": "test-console",
					},
					"downloads": {
						"hostname":       "downloads.test.example.com",
						"tls_secret_ref": "test-downloads",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("Can't remove 'oauth' component route hostname for ingress '%s' in cluster '%s' without also removing TLS secret reference", out.ID, clusterID))

			By("Try to remove only the tls_secret_ref")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "oauth.test.example.com",
						"tls_secret_ref": "",
					},
					"console": {
						"hostname":       "console.test.example.com",
						"tls_secret_ref": "test-console",
					},
					"downloads": {
						"hostname":       "downloads.test.example.com",
						"tls_secret_ref": "test-downloads",
					},
				},
				Cluster: &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("Can't update 'oauth' component route hostname for ingress '%s' in cluster '%s' without also supplying a new TLS secret reference", out.ID, clusterID))

			By("Try with blank component routes")
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{},
				Cluster:         &clusterID,
			}
			_, err = ingressService.Apply(&args)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Error: Provider produced inconsistent result after apply")

			By("Try with invalid cluster ID")
			invalidID := "asdf"
			args = exec.IngressArgs{
				ComponentRoutes: &map[string]map[string]string{
					"oauth": {
						"hostname":       "oauth.test.example.com",
						"tls_secret_ref": "test-oauth",
					},
					"console": {
						"hostname":       "console.test.example.com",
						"tls_secret_ref": "test-console",
					},
					"downloads": {
						"hostname":       "downloads.test.example.com",
						"tls_secret_ref": "test-downloads",
					},
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
			if profile.GetClusterType().HCP {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(MatchRegexp(`Can't update route selectors on[\s\S]?Hosted Control Plane cluster '%s'`, clusterID))
				return
			}
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
