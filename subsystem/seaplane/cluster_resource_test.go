/*
Copyright (c) 2024 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package seaplane

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/openshift-online/ocm-sdk-go/testing"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

func verifySeaplaneProxyRequest(httpProxy, httpsProxy, noProxy string) http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		var body struct {
			Proxy map[string]string `json:"proxy"`
			Spec  struct {
				HostedCluster struct {
					Configuration struct {
						Proxy map[string]string `json:"proxy"`
					} `json:"configuration"`
				} `json:"hostedCluster"`
			} `json:"spec"`
		}
		Expect(json.NewDecoder(req.Body).Decode(&body)).To(Succeed())
		Expect(body.Proxy).To(BeNil(), "response-only proxy projection must not be sent back")
		want := map[string]string{}
		if httpProxy != "" {
			want["httpProxy"] = httpProxy
		}
		if httpsProxy != "" {
			want["httpsProxy"] = httpsProxy
		}
		if noProxy != "" {
			want["noProxy"] = noProxy
		}
		proxy := body.Spec.HostedCluster.Configuration.Proxy
		if proxy == nil {
			proxy = map[string]string{}
		}
		Expect(proxy).To(Equal(want))
	}
}

func verifySeaplaneClusterUpdate(expected string) http.HandlerFunc {
	return func(_ http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		Expect(err).NotTo(HaveOccurred())
		var actualObject, expectedObject map[string]any
		Expect(json.Unmarshal(body, &actualObject)).To(Succeed())
		Expect(json.Unmarshal([]byte(expected), &expectedObject)).To(Succeed())

		Expect(actualObject["kind"]).To(Equal("Cluster"))
		Expect(actualObject["apiVersion"]).To(Equal("seaplane.io/v1alpha1"))
		metadata, ok := actualObject["metadata"].(map[string]any)
		Expect(ok).To(BeTrue(), "PUT body should contain resource metadata")
		Expect(metadata["name"]).To(Equal("my-cluster"))
		Expect(metadata["uid"]).To(Equal("test-cluster-id"))
		status, ok := actualObject["status"].(map[string]any)
		Expect(ok).To(BeTrue(), "PUT body should preserve the fetched status")
		Expect(status["phase"]).To(Equal("Ready"))

		expectSeaplaneJSONSubset(actualObject, expectedObject, "")
	}
}

func expectSeaplaneJSONSubset(actual, expected any, path string) {
	expectedObject, isObject := expected.(map[string]any)
	if !isObject {
		Expect(actual).To(Equal(expected), "unexpected value at %s", path)
		return
	}
	actualObject, isObject := actual.(map[string]any)
	Expect(isObject).To(BeTrue(), "expected object at %s", path)
	for key, expectedValue := range expectedObject {
		actualValue, exists := actualObject[key]
		fieldPath := key
		if path != "" {
			fieldPath = path + "." + key
		}
		// Optional nil API fields use omitempty and are serialized as absent.
		if expectedValue == nil && (!exists || actualValue == nil) {
			continue
		}
		Expect(exists).To(BeTrue(), "missing field %s", fieldPath)
		expectSeaplaneJSONSubset(actualValue, expectedValue, fieldPath)
	}
}

func seaplaneClusterProxyResponse(httpProxy, httpsProxy, noProxy string) string {
	proxyConfig := map[string]string{}
	if httpProxy != "" {
		proxyConfig["httpProxy"] = httpProxy
	}
	if httpsProxy != "" {
		proxyConfig["httpsProxy"] = httpsProxy
	}
	if noProxy != "" {
		proxyConfig["noProxy"] = noProxy
	}
	configuration := ""
	if len(proxyConfig) > 0 {
		configJSON, err := json.Marshal(proxyConfig)
		Expect(err).NotTo(HaveOccurred())
		configuration = fmt.Sprintf(`"configuration":{"proxy":%s},`, configJSON)
	}

	return fmt.Sprintf(`{
		"metadata": {
			"uid": "test-cluster-id",
			"name": "my-cluster",
			"creationTimestamp": "2024-01-01T00:00:00Z"
		},
		"spec": {
			"hostedCluster": {
				%s
				"platform": {
					"type": "AWS",
					"aws": {
						"region": "us-east-1",
						"cloudProviderConfig": {
							"vpc": "vpc-0def456",
							"zone": "us-east-1a",
							"subnet": {
								"id": "subnet-0abc123"
							}
						}
					}
				}
			}
		},
		"status": {
			"phase": "Ready",
			"controlPlaneEndpoint": {
				"host": "api.my-cluster.example.com",
				"port": 6443
			}
		}
	}`, configuration)
}

func seaplaneClusterAllowedPrincipalsResponse(principals []string) string {
	principalsJSON, err := json.Marshal(principals)
	Expect(err).NotTo(HaveOccurred())

	return fmt.Sprintf(`{
		"metadata": {
			"uid": "test-cluster-id",
			"name": "my-cluster",
			"creationTimestamp": "2024-01-01T00:00:00Z"
		},
		"spec": {
			"hostedCluster": {
				"platform": {
					"type": "AWS",
					"aws": {
						"region": "us-east-1",
						"additionalAllowedPrincipals": %s,
						"cloudProviderConfig": {
							"vpc": "vpc-0def456",
							"zone": "us-east-1a",
							"subnet": {
								"id": "subnet-0abc123"
							}
						}
					}
				}
			}
		},
		"status": {
			"phase": "Ready",
			"controlPlaneEndpoint": {
				"host": "api.my-cluster.example.com",
				"port": 6443
			}
		}
	}`, string(principalsJSON))
}

func verifySeaplaneAllowedPrincipalsRequest(principals []string) http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		var body struct {
			Spec struct {
				HostedCluster struct {
					Platform struct {
						AWS struct {
							AdditionalAllowedPrincipals []string `json:"additionalAllowedPrincipals"`
						} `json:"aws"`
					} `json:"platform"`
				} `json:"hostedCluster"`
			} `json:"spec"`
		}
		Expect(json.NewDecoder(req.Body).Decode(&body)).To(Succeed())
		Expect(body.Spec.HostedCluster.Platform.AWS.AdditionalAllowedPrincipals).To(Equal(principals))
	}
}

func seaplaneClusterAllowedPrincipalsTerraformConfig(serverURL string, principals []string) string {
	principalsJSON, err := json.Marshal(principals)
	Expect(err).NotTo(HaveOccurred())

	return fmt.Sprintf(`
		provider "rhcs" {
			alias          = "sp"
			seaplane_url   = %q
			aws_account_id = "123456789012"
			aws_region     = "us-east-1"
		}

		resource "rhcs_cluster_seaplane" "test" {
			provider              = rhcs.sp
			name                  = "my-cluster"
			operator_roles_prefix = "my-cluster"
			aws = {
				aws_subnet_ids                = ["subnet-0abc123"]
				vpc_id                        = "vpc-0def456"
				availability_zones            = ["us-east-1a"]
				additional_allowed_principals = %s
			}
		}
	`, serverURL, string(principalsJSON))
}

func seaplaneClusterTerraformConfig(serverURL, httpProxy, httpsProxy, noProxy string, includeProxy bool) string {
	proxyAttributes := ""
	if includeProxy {
		proxyAttributes = fmt.Sprintf(`
				http_proxy  = %q
				https_proxy = %q
				no_proxy    = %q`, httpProxy, httpsProxy, noProxy)
	}
	return fmt.Sprintf(`
		provider "rhcs" {
			alias          = "sp"
			seaplane_url = %q
			aws_account_id = "123456789012"
			aws_region     = "us-east-1"
		}

		resource "rhcs_cluster_seaplane" "test" {
			provider              = rhcs.sp
			name                  = "my-cluster"
			operator_roles_prefix = "my-cluster"
			%s
			aws = {
				aws_subnet_ids     = ["subnet-0abc123"]
				vpc_id             = "vpc-0def456"
				availability_zones = ["us-east-1a"]
			}
		}
	`, serverURL, proxyAttributes)
}

var _ = Describe("rhcs_cluster_seaplane", func() {
	Context("Provider configuration validation", func() {
		It("Reports a descriptive error when seaplane_url is absent from the provider", func() {
			Terraform.Source(`
				resource "rhcs_cluster_seaplane" "test" {
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("seaplane_url")
		})

		It("Reports a descriptive error when aws_account_id is absent", func() {
			Terraform.Source(`
				provider "rhcs" {
						alias          = "sp"
					seaplane_url = "https://abc123.execute-api.us-east-1.amazonaws.com/prod"
				}

				resource "rhcs_cluster_seaplane" "test" {
						provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("aws_account_id")
		})

		It("Reports a descriptive error when aws_region contradicts the region in seaplane_url", func() {
			Terraform.Source(`
				provider "rhcs" {
						alias          = "sp"
					seaplane_url = "https://abc123.execute-api.us-east-1.amazonaws.com/prod"
					aws_account_id = "123456789012"
					aws_region     = "us-west-2"
				}

				resource "rhcs_cluster_seaplane" "test" {
						provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("AWS region mismatch")
		})
	})

	Context("IPv4 networking", func() {
		It("requires replacement for CIDR changes and permits updates with unchanged networking", func() {
			server := NewServer()
			defer server.Close()

			response := `{
				"metadata": {"uid": "test-cluster-id", "name": "my-cluster"},
				"spec": {"hostedCluster": {
					"networking": {
						"machineNetwork": [{"cidr": "10.0.0.0/16"}],
						"clusterNetwork": [{"cidr": "10.128.0.0/14", "hostPrefix": 23}],
						"serviceNetwork": [{"cidr": "172.30.0.0/16"}],
						"networkType": "OVNKubernetes"
					},
					"platform": {"type": "AWS", "aws": {
						"region": "us-east-1",
						"cloudProviderConfig": {"vpc": "vpc-0def456", "zone": "us-east-1a", "subnet": {"id": "subnet-0abc123"}}
					}}
				}},
				"status": {"phase": "Ready"}
			}`
			header := http.Header{"Content-Type": []string{"application/json"}}
			server.RouteToHandler(http.MethodGet, "/api/v0/clusters/my-cluster", RespondWith(http.StatusOK, response, header))
			server.AppendHandlers(CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/v0/clusters"),
				verifyJQ(`.spec.hostedCluster.networking.machineNetwork[0].cidr`, "10.0.0.0/16"),
				verifyJQ(`.spec.hostedCluster.networking.clusterNetwork[0].cidr`, "10.128.0.0/14"),
				verifyJQ(`.spec.hostedCluster.networking.clusterNetwork[0].hostPrefix`, float64(23)),
				verifyJQ(`.spec.hostedCluster.networking.serviceNetwork[0].cidr`, "172.30.0.0/16"),
				RespondWith(http.StatusCreated, response, header),
			))

			config := fmt.Sprintf(`
				provider "rhcs" {
					alias          = "sp"
					seaplane_url   = %q
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}
				resource "rhcs_cluster_seaplane" "test" {
					provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
					network = {
						machine_cidr = jsonencode([{cidr = "10.0.0.0/16"}])
						pod_cidr     = jsonencode([{cidr = "10.128.0.0/14", hostPrefix = 23}])
						service_cidr = jsonencode([{cidr = "172.30.0.0/16"}])
						network_type = "OVNKubernetes"
					}
					%%s
				}
			`, server.URL())
			Terraform.Source(fmt.Sprintf(config, ""))
			Expect(Terraform.Apply().ExitCode).To(BeZero())
			Expect(Terraform.Run("plan", "-detailed-exitcode").ExitCode).To(BeZero())

			for _, cidr := range []string{"10.0.0.0/16", "10.128.0.0/14", "172.30.0.0/16"} {
				By("planning replacement when changing " + cidr)
				Terraform.Source(strings.Replace(fmt.Sprintf(config, ""), cidr, "192.168.0.0/16", 1))
				plan := Terraform.Run("plan", "-detailed-exitcode")
				Expect(plan.ExitCode).To(Equal(2))
				plan.VerifyOutputContainsSubstring("1 to add, 0 to change, 1 to destroy")
			}

			By("planning an in-place expiration update with unchanged networking")
			Terraform.Source(fmt.Sprintf(config, `expiration_timestamp = "2030-01-01T00:00:00Z"`))
			plan := Terraform.Run("plan", "-detailed-exitcode")
			Expect(plan.ExitCode).To(Equal(2))
			plan.VerifyOutputContainsSubstring("0 to add, 1 to change, 0 to destroy")
		})
	})

	Context("Create operation", func() {
		var seaplaneServer *Server

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
		})

		It("creates a cluster successfully", func() {
			clusterResponse := `{
			"metadata": {
				"uid": "test-cluster-id",
				"name": "my-cluster",
				"creationTimestamp": "2024-01-01T00:00:00Z"
			},
			"spec": {
				"hostedCluster": {
					"platform": {
						"type": "AWS",
						"aws": {
							"region": "us-east-1",
							"rolesRef": {
								"ingressARN": "arn:aws:iam::123456789012:role/my-cluster-ingress",
								"kubeCloudControllerARN": "arn:aws:iam::123456789012:role/my-cluster-cloud-controller-manager",
								"storageARN": "arn:aws:iam::123456789012:role/my-cluster-ebs-csi",
								"imageRegistryARN": "arn:aws:iam::123456789012:role/my-cluster-image-registry",
								"networkARN": "arn:aws:iam::123456789012:role/my-cluster-network-config",
								"controlPlaneOperatorARN": "arn:aws:iam::123456789012:role/my-cluster-control-plane-operator",
								"nodePoolManagementARN": "arn:aws:iam::123456789012:role/my-cluster-node-pool-management"
							},
							"cloudProviderConfig": {
								"vpc": "vpc-0def456",
								"zone": "us-east-1a",
								"subnet": {
									"id": "subnet-0abc123"
								}
							}
						}
					}
				}
			},
			"status": {
				"phase": "Provisioning",
				"controlPlaneEndpoint": {
					"host": "api.my-cluster.example.com",
					"port": 6443
				}
			}
		}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, clusterResponse, header),
				),
			)

			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
						alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_seaplane" "test" {
						provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("creates a cluster with proxy settings", func() {
			httpProxy := "http://proxy.example.com:8080"
			httpsProxy := "https://proxy.example.com:8443"
			noProxy := "localhost,127.0.0.1,.example.com"
			clusterResponse := seaplaneClusterProxyResponse(httpProxy, httpsProxy, noProxy)
			header := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					verifySeaplaneProxyRequest(httpProxy, httpsProxy, noProxy),
					RespondWith(http.StatusCreated, clusterResponse, header),
				),
			)

			Terraform.Source(seaplaneClusterTerraformConfig(seaplaneServer.URL(), httpProxy, httpsProxy, noProxy, true))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource := Terraform.Resource("rhcs_cluster_seaplane", "test")
			Expect(resource).To(matchJQ(".attributes.http_proxy", httpProxy))
			Expect(resource).To(matchJQ(".attributes.https_proxy", httpsProxy))
			Expect(resource).To(matchJQ(".attributes.no_proxy", noProxy))
		})

		It("creates a cluster with additional allowed principals", func() {
			principals := []string{
				"arn:aws:iam::123456789012:role/example-consumer",
				"arn:aws:iam::123456789012:role/another-consumer",
			}
			header := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					verifySeaplaneAllowedPrincipalsRequest(principals),
					RespondWith(
						http.StatusCreated,
						seaplaneClusterAllowedPrincipalsResponse(principals),
						header,
					),
				),
			)

			Terraform.Source(seaplaneClusterAllowedPrincipalsTerraformConfig(
				seaplaneServer.URL(), principals,
			))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource := Terraform.Resource("rhcs_cluster_seaplane", "test")
			Expect(resource).To(matchJQ(".attributes.aws.additional_allowed_principals[0]", principals[0]))
			Expect(resource).To(matchJQ(".attributes.aws.additional_allowed_principals[1]", principals[1]))
		})

		It("reports a Seaplane proxy validation failure on create", func() {
			// Seaplane limits ProxyConfiguration.httpProxy to 2048 characters.
			// Verify a Platform API rejection reaches Terraform on create.
			invalidHTTPProxy := "http://proxy.example.com/" + strings.Repeat("x", 2049)
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					verifySeaplaneProxyRequest(invalidHTTPProxy, "", ""),
					RespondWith(http.StatusUnprocessableEntity,
						`{"message":"spec.hostedCluster.configuration.proxy.httpProxy exceeds 2048 characters"}`,
						http.Header{"Content-Type": []string{"application/json"}}),
				),
			)

			Terraform.Source(seaplaneClusterTerraformConfig(
				seaplaneServer.URL(), invalidHTTPProxy, "", "", true,
			))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).NotTo(BeZero())
		})
	})

	Context("Customer AWS tags", func() {
		var seaplaneServer *Server

		// clusterWithTags is the Platform API representation of a cluster created
		// with the two tags used throughout this context.
		clusterWithTags := `{
			"metadata": {
				"uid": "test-cluster-id",
				"name": "my-cluster",
				"creationTimestamp": "2024-01-01T00:00:00Z"
			},
			"spec": {
				"tags": {
					"cost-center": "cc-1234",
					"environment": "production"
				},
				"hostedCluster": {
					"platform": {
						"type": "AWS",
						"aws": {
							"region": "us-east-1",
							"cloudProviderConfig": {
								"vpc": "vpc-0def456",
								"zone": "us-east-1a",
								"subnet": {
									"id": "subnet-0abc123"
								}
							}
						}
					}
				}
			},
			"status": {
				"phase": "Ready",
				"controlPlaneEndpoint": {
					"host": "api.my-cluster.example.com",
					"port": 6443
				}
			}
		}`

		header := http.Header{"Content-Type": []string{"application/json"}}

		// clusterSource renders the cluster configuration with the given tags
		// block (HCL, may be empty).
		clusterSource := func(url, tags string) string {
			return fmt.Sprintf(`
				provider "rhcs" {
						alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_seaplane" "test" {
						provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
					%s
				}
			`, url, tags)
		}

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
		})

		It("sends the tags as spec.tags on create", func() {
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					verifyJQ(`.spec.tags."cost-center"`, "cc-1234"),
					verifyJQ(`.spec.tags.environment`, "production"),
					RespondWith(http.StatusCreated, clusterWithTags, header),
				),
			)

			Terraform.Source(clusterSource(seaplaneServer.URL(), `
					tags = {
						"cost-center" = "cc-1234"
						"environment" = "production"
					}`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("updates cluster tags in place through the Platform API", func() {
			clusterWithUpdatedTags := strings.Replace(clusterWithTags, "cc-1234", "cc-9999", 1)
			seaplaneServer.RouteToHandler(
				http.MethodGet, "/api/v0/clusters/my-cluster",
				RespondWith(http.StatusOK, clusterWithTags, header),
			)
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, clusterWithTags, header),
				),
			)

			Terraform.Source(clusterSource(seaplaneServer.URL(), `
					tags = {
						"cost-center" = "cc-1234"
						"environment" = "production"
					}`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/my-cluster"),
					verifyJQ(`.spec.tags."cost-center"`, "cc-9999"),
					verifyJQ(`.spec.tags.environment`, "production"),
					RespondWith(http.StatusOK, clusterWithUpdatedTags, header),
				),
			)
			Terraform.Source(clusterSource(seaplaneServer.URL(), `
					tags = {
						"cost-center" = "cc-9999"
						"environment" = "production"
					}`))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("rejects more tags than the Platform API accepts", func() {
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusUnprocessableEntity,
						`{"message":"rendered AWS resource tags would contain 26 entries including reserved system tags; maximum is 25"}`,
						header),
				),
			)

			tags := make([]string, 0, 24)
			for i := range 24 {
				tags = append(tags, fmt.Sprintf("    \"tag%d\" = \"value\"", i))
			}

			Terraform.Source(clusterSource(seaplaneServer.URL(), fmt.Sprintf(`
					tags = {
%s
					}`, strings.Join(tags, "\n"))))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Failed to create Cluster")
		})

		It("rejects a tag key reserved for AWS", func() {
			Terraform.Source(clusterSource(seaplaneServer.URL(), `
					tags = {
						"aws:cloudformation:stack-name" = "mine"
					}`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("reserved for AWS use")
		})
	})

	Context("Cluster capabilities", func() {
		var seaplaneServer *Server

		// clusterWithCapabilities is the Platform API representation of a cluster
		// created with the ImageRegistry capability disabled.
		clusterWithCapabilities := `{"metadata":{"uid":"test-cluster-id","name":"my-cluster","creationTimestamp":"2024-01-01T00:00:00Z"},"spec":{"hostedCluster":{"capabilities":{"disabled":["ImageRegistry"]},"platform":{"type":"AWS","aws":{"region":"us-east-1","cloudProviderConfig":{"vpc":"vpc-0def456","zone":"us-east-1a","subnet":{"id":"subnet-0abc123"}}}}}},"status":{"phase":"Ready","controlPlaneEndpoint":{"host":"api.my-cluster.example.com","port":6443}}}`

		header := http.Header{"Content-Type": []string{"application/json"}}

		// clusterCapabilitiesSource renders the cluster configuration with the
		// given capabilities block (HCL, may be empty).
		clusterCapabilitiesSource := func(url, capabilities string) string {
			return fmt.Sprintf(`
				provider "rhcs" {
					alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_seaplane" "test" {
					provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
					%s
				}
			`, url, capabilities)
		}

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
		})

		It("sends disabled capabilities on create", func() {
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					VerifyJQ(`.spec.hostedCluster.capabilities.disabled[0]`, "ImageRegistry"),
					VerifyJQ(`.spec.hostedCluster.capabilities.disabled | length`, 1),
					RespondWith(http.StatusCreated, clusterWithCapabilities, header),
				),
			)

			Terraform.Source(clusterCapabilitiesSource(seaplaneServer.URL(), `
					disabled_capabilities = jsonencode(["ImageRegistry"])`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("replaces the cluster when the capabilities change", func() {
			// Capabilities can only be set at cluster creation, so changing them
			// must be planned as a replacement rather than an update.
			seaplaneServer.RouteToHandler(
				http.MethodGet, "/api/v0/clusters/my-cluster",
				RespondWith(http.StatusOK, clusterWithCapabilities, header),
			)
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, clusterWithCapabilities, header),
				),
			)

			Terraform.Source(clusterCapabilitiesSource(seaplaneServer.URL(), `
					disabled_capabilities = jsonencode(["ImageRegistry"])`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			Terraform.Source(clusterCapabilitiesSource(seaplaneServer.URL(), `
					disabled_capabilities = jsonencode([])`))

			runOutput = Terraform.Run("plan", "-no-color")
			Expect(runOutput.ExitCode).To(BeZero())
			runOutput.VerifyOutputContainsSubstring("must be replaced")
		})

		It("passes unknown capabilities through to the API", func() {
			// Capability names are validated server side, not by the provider,
			// so an unrecognized value must be sent in the create request
			// unchanged and the API response surfaced to the user.
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					VerifyJQ(`.spec.hostedCluster.capabilities.disabled[0]`, "NotACapability"),
					RespondWith(http.StatusUnprocessableEntity, `{"reason":"InvalidRequest","message":"invalid capability \"NotACapability\""}`, header),
				),
			)

			Terraform.Source(clusterCapabilitiesSource(seaplaneServer.URL(), `
					disabled_capabilities = jsonencode(["NotACapability"])`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("server rejected our request")
		})

		It("rejects a capability both enabled and disabled", func() {
			Terraform.Source(clusterCapabilitiesSource(seaplaneServer.URL(), `
					enabled_capabilities  = jsonencode(["ImageRegistry"])
					disabled_capabilities = jsonencode(["ImageRegistry"])`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("capability \"ImageRegistry\" cannot be both enabled and disabled at once")
		})
	})

	Context("Update operation", func() {
		var seaplaneServer *Server

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
		})

		It("updates cluster expiration_timestamp", func() {
			clusterWithoutExpiration := `{
				"metadata": {
					"uid": "test-cluster-id",
					"name": "my-cluster",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1",
								"cloudProviderConfig": {
									"vpc": "vpc-0def456",
									"zone": "us-east-1a",
									"subnet": {
										"id": "subnet-0abc123"
									}
								}
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.my-cluster.example.com",
						"port": 6443
					}
				}
			}`

			clusterWithExpiration := `{
				"metadata": {
					"uid": "test-cluster-id",
					"name": "my-cluster",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"expirationTimestamp": "2025-01-01T00:00:00Z",
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1",
								"cloudProviderConfig": {
									"vpc": "vpc-0def456",
									"zone": "us-east-1a",
									"subnet": {
										"id": "subnet-0abc123"
									}
								}
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.my-cluster.example.com",
						"port": 6443
					}
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.AppendHandlers(
				// First apply: create without expiration
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, clusterWithoutExpiration, header),
				),
				// Second apply: refresh (Read), then send only the changed field.
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, clusterWithoutExpiration, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, clusterWithoutExpiration, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/my-cluster"),
					verifySeaplaneClusterUpdate(`{"spec":{"expirationTimestamp":"2025-01-01T00:00:00Z"}}`),
					RespondWith(http.StatusOK, clusterWithExpiration, header),
				),
			)

			// First apply: create without expiration
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
						alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_seaplane" "test" {
						provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Second apply: update to add expiration
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
						alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_seaplane" "test" {
						provider                = rhcs.sp
					name                    = "my-cluster"
					operator_roles_prefix   = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
					expiration_timestamp   = "2025-01-01T00:00:00Z"
				}
			`, seaplaneServer.URL()))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("updates cluster proxy settings", func() {
			initialHTTPProxy := "http://proxy.example.com:8080"
			initialHTTPSProxy := "https://proxy.example.com:8443"
			initialNoProxy := "localhost,127.0.0.1"
			updatedHTTPProxy := "http://proxy2.example.com:8080"
			updatedHTTPSProxy := "https://proxy2.example.com:8443"
			updatedNoProxy := "localhost,127.0.0.1,.example.com"
			invalidHTTPProxy := "http://proxy.example.com/" + strings.Repeat("x", 2049)
			initialResponse := seaplaneClusterProxyResponse(initialHTTPProxy, initialHTTPSProxy, initialNoProxy)
			updatedResponse := seaplaneClusterProxyResponse(updatedHTTPProxy, updatedHTTPSProxy, updatedNoProxy)
			clearedResponse := seaplaneClusterProxyResponse("", "", "")
			header := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					verifySeaplaneProxyRequest(initialHTTPProxy, initialHTTPSProxy, initialNoProxy),
					RespondWith(http.StatusCreated, initialResponse, header),
				),
				// The second apply attempts an invalid edit. Terraform refreshes the
				// existing state before the rejected partial PUT.
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, initialResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, initialResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/my-cluster"),
					verifySeaplaneClusterUpdate(fmt.Sprintf(`{
						"spec": {
							"hostedCluster": {
								"configuration": {
									"proxy": {
										"httpProxy": %q
									}
								}
							}
						}
					}`, invalidHTTPProxy)),
					RespondWith(http.StatusUnprocessableEntity,
						`{"message":"spec.hostedCluster.configuration.proxy.httpProxy exceeds 2048 characters"}`,
						header),
				),
				// The valid update below follows the rejected update. The GET
				// confirms that the previous configuration is still present.
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, initialResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, initialResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/my-cluster"),
					verifySeaplaneClusterUpdate(fmt.Sprintf(`{
						"spec": {
							"hostedCluster": {
								"configuration": {
									"proxy": {
										"httpProxy": %q,
										"httpsProxy": %q,
										"noProxy": %q
									}
								}
							}
						}
					}`, updatedHTTPProxy, updatedHTTPSProxy, updatedNoProxy)),
					RespondWith(http.StatusOK, updatedResponse, header),
				),
				// The third apply removes the proxy attributes from Terraform.
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, updatedResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, updatedResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/my-cluster"),
					verifySeaplaneClusterUpdate(`{
						"spec": {
							"hostedCluster": {
								"configuration": {
									"proxy": null
								}
							}
						}
					}`),
					RespondWith(http.StatusOK, clearedResponse, header),
				),
			)

			Terraform.Source(seaplaneClusterTerraformConfig(
				seaplaneServer.URL(), initialHTTPProxy, initialHTTPSProxy, initialNoProxy, true,
			))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			Terraform.Source(seaplaneClusterTerraformConfig(
				seaplaneServer.URL(), invalidHTTPProxy, initialHTTPSProxy, initialNoProxy, true,
			))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).NotTo(BeZero())

			Terraform.Source(seaplaneClusterTerraformConfig(
				seaplaneServer.URL(), updatedHTTPProxy, updatedHTTPSProxy, updatedNoProxy, true,
			))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource := Terraform.Resource("rhcs_cluster_seaplane", "test")
			Expect(resource).To(matchJQ(".attributes.http_proxy", updatedHTTPProxy))
			Expect(resource).To(matchJQ(".attributes.https_proxy", updatedHTTPSProxy))
			Expect(resource).To(matchJQ(".attributes.no_proxy", updatedNoProxy))

			Terraform.Source(seaplaneClusterTerraformConfig(seaplaneServer.URL(), "", "", "", false))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource = Terraform.Resource("rhcs_cluster_seaplane", "test")
			Expect(resource).To(matchJQ(".attributes.http_proxy", nil))
			Expect(resource).To(matchJQ(".attributes.https_proxy", nil))
			Expect(resource).To(matchJQ(".attributes.no_proxy", nil))
		})

		It("updates additional allowed principals", func() {
			initialPrincipals := []string{
				"arn:aws:iam::123456789012:role/example-consumer",
			}
			updatedPrincipals := []string{
				"arn:aws:iam::123456789012:role/example-consumer",
				"arn:aws:iam::123456789012:role/another-consumer",
			}
			clearedPrincipals := []string{}
			initialResponse := seaplaneClusterAllowedPrincipalsResponse(initialPrincipals)
			updatedResponse := seaplaneClusterAllowedPrincipalsResponse(updatedPrincipals)
			clearedResponse := seaplaneClusterAllowedPrincipalsResponse(clearedPrincipals)
			header := http.Header{"Content-Type": []string{"application/json"}}
			updatedPrincipalsJSON, err := json.Marshal(updatedPrincipals)
			Expect(err).NotTo(HaveOccurred())

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					verifySeaplaneAllowedPrincipalsRequest(initialPrincipals),
					RespondWith(http.StatusCreated, initialResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, initialResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, initialResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/my-cluster"),
					verifySeaplaneClusterUpdate(fmt.Sprintf(`{
						"spec": {
							"hostedCluster": {
								"platform": {
									"aws": {
										"additionalAllowedPrincipals": %s
									}
								}
							}
						}
					}`, string(updatedPrincipalsJSON))),
					RespondWith(http.StatusOK, updatedResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, updatedResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, updatedResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/my-cluster"),
					verifySeaplaneClusterUpdate(`{
						"spec": {
							"hostedCluster": {
								"platform": {
									"aws": {
										"additionalAllowedPrincipals": null
									}
								}
							}
						}
					}`),
					RespondWith(http.StatusOK, clearedResponse, header),
				),
			)

			Terraform.Source(seaplaneClusterAllowedPrincipalsTerraformConfig(
				seaplaneServer.URL(), initialPrincipals,
			))
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			Terraform.Source(seaplaneClusterAllowedPrincipalsTerraformConfig(
				seaplaneServer.URL(), updatedPrincipals,
			))
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource := Terraform.Resource("rhcs_cluster_seaplane", "test")
			Expect(resource).To(matchJQ(".attributes.aws.additional_allowed_principals[0]", updatedPrincipals[0]))
			Expect(resource).To(matchJQ(".attributes.aws.additional_allowed_principals[1]", updatedPrincipals[1]))

			Terraform.Source(seaplaneClusterAllowedPrincipalsTerraformConfig(
				seaplaneServer.URL(), nil,
			))
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource = Terraform.Resource("rhcs_cluster_seaplane", "test")
			Expect(resource).To(matchJQ(".attributes.aws.additional_allowed_principals", nil))
		})
	})

	Context("Delete operation", func() {
		var seaplaneServer *Server

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
		})

		It("deletes a cluster successfully", func() {
			createResponse := `{
				"metadata": {
					"uid": "test-cluster-id",
					"name": "my-cluster",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1",
								"cloudProviderConfig": {
									"vpc": "vpc-0def456",
									"zone": "us-east-1a",
									"subnet": {
										"id": "subnet-0abc123"
									}
								}
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.my-cluster.example.com",
						"port": 6443
					}
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/v0/clusters/my-cluster"),
					RespondWith(http.StatusNoContent, ""),
				),
			)

			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
						alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_seaplane" "test" {
						provider              = rhcs.sp
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws = {
						aws_subnet_ids     = ["subnet-0abc123"]
						vpc_id             = "vpc-0def456"
						availability_zones = ["us-east-1a"]
					}
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})
})

var _ = Describe("rhcs_oidcconfig_seaplane", func() {
	It("requires installer role ARN for unmanaged OIDC configurations", func() {
		Terraform.Source(`
			provider "rhcs" {
				alias          = "sp"
				seaplane_url = "https://abc123.execute-api.us-east-1.amazonaws.com/prod"
				aws_account_id = "123456789012"
				aws_region     = "us-east-1"
				aws_caller_arn = "arn:aws:iam::123456789012:user/test"
			}

			resource "rhcs_oidcconfig_seaplane" "test" {
				provider = rhcs.sp
				name     = "test-cluster.oidc"
				type     = "unmanaged"
			}
		`)

		runOutput := Terraform.Apply()
		runOutput.VerifyErrorContainsSubstring("Installer role ARN must be specified for unmanaged type")
	})
})
