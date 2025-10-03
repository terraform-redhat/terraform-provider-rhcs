package hcp

import (
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	. "github.com/onsi/gomega/ghttp"       // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

const (
	testClusterRoute      = "/api/clusters_mgmt/v1/clusters/test-cluster"
	testExternalAuthRoute = "/api/clusters_mgmt/v1/clusters/test-cluster/external_auth_config/external_auths"
	testProviderRoute     = "/api/clusters_mgmt/v1/clusters/test-cluster/external_auth_config/external_auths/test-provider"
)

var _ = Describe("External Auth Provider", func() {

	// Create cluster template without external auth enabled
	clusterWithoutExternalAuthTemplate := func() string {
		cluster, err := cmv1.NewCluster().
			ID("test-cluster").
			Name("test-cluster").
			State(cmv1.ClusterStateReady).
			Build()
		Expect(err).ToNot(HaveOccurred())

		b := new(strings.Builder)
		err = cmv1.MarshalCluster(cluster, b)
		Expect(err).ToNot(HaveOccurred())
		return b.String()
	}

	// Create cluster template with external auth enabled
	clusterWithExternalAuthTemplate := func() string {
		cluster, err := cmv1.NewCluster().
			ID("test-cluster").
			Name("test-cluster").
			State(cmv1.ClusterStateReady).
			ExternalAuthConfig(cmv1.NewExternalAuthConfig().Enabled(true)).
			Build()
		Expect(err).ToNot(HaveOccurred())

		b := new(strings.Builder)
		err = cmv1.MarshalCluster(cluster, b)
		Expect(err).ToNot(HaveOccurred())
		return b.String()
	}

	// Create external auth provider template
	externalAuthProviderTemplate := func() string {
		provider, err := cmv1.NewExternalAuth().
			ID("test-provider").
			Issuer(cmv1.NewTokenIssuer().
				URL("https://example.com").
				Audiences("audience1")).
			Build()
		Expect(err).ToNot(HaveOccurred())

		b := new(strings.Builder)
		err = cmv1.MarshalExternalAuth(provider, b)
		Expect(err).ToNot(HaveOccurred())
		return b.String()
	}

	// Create updated external auth provider template
	updatedExternalAuthProviderTemplate := func() string {
		provider, err := cmv1.NewExternalAuth().
			ID("test-provider").
			Issuer(cmv1.NewTokenIssuer().
				URL("https://updated-example.com").
				Audiences("audience1", "audience2")).
			Build()
		Expect(err).ToNot(HaveOccurred())

		b := new(strings.Builder)
		err = cmv1.MarshalExternalAuth(provider, b)
		Expect(err).ToNot(HaveOccurred())
		return b.String()
	}

	// Create external auth provider template with console clients
	externalAuthProviderWithConsoleTemplate := func() string {
		provider, err := cmv1.NewExternalAuth().
			ID("test-provider").
			Issuer(cmv1.NewTokenIssuer().
				URL("https://example.com").
				Audiences("audience1")).
			Clients(cmv1.NewExternalAuthClientConfig().
				ID("console-client").
				Secret("console-secret")).
			Build()
		Expect(err).ToNot(HaveOccurred())

		b := new(strings.Builder)
		err = cmv1.MarshalExternalAuth(provider, b)
		Expect(err).ToNot(HaveOccurred())
		return b.String()
	}
	Context("static validation", func() {
		It("fails if cluster ID is empty", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = ""
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("cluster ID may not be empty/blank string")
		})

		It("fails if provider ID is empty", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = ""
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("provider ID may not be empty/blank string")
		})

		It("fails if issuer URL is not HTTPS", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "http://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("issuer URL must use HTTPS")
		})

		It("fails if client ID is provided but secret is missing", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						id = "client-123"
						# secret missing
					}
				]
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Client secret is required when client ID is provided")
		})

		It("succeeds when client ID and secret are both provided", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						id = "client-123"
						secret = "client-secret"
					}
				]
			}`)
			Expect(Terraform.Validate().ExitCode).To(BeZero())
		})

		It("succeeds when client has no ID and no secret", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						component = {
							name = "test-component"
							namespace = "test-namespace"
						}
					}
				]
			}`)
			Expect(Terraform.Validate().ExitCode).To(BeZero())
		})

		It("validates validation rules have both claim and required_value", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				claim = {
					validation_rules = [
						{
							claim = "test-claim"
							required_value = "test-value"
						}
					]
				}
			}`)
			Expect(Terraform.Validate().ExitCode).To(BeZero())
		})

		It("accepts minimal valid configuration", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			Expect(Terraform.Validate().ExitCode).To(BeZero())
		})

		It("accepts complex valid configuration", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://auth.example.com/oauth2"
					audiences = ["audience1", "audience2"]
					ca = "-----BEGIN CERTIFICATE-----\nMIIC..."
				}
				clients = [
					{
						component = {
							name = "console"
							namespace = "openshift-console"
						}
						id = "console-client"
						secret = "super-secret"
						extra_scopes = ["openid", "profile"]
					}
				]
				claim = {
					mappings = {
						username = {
							claim = "preferred_username"
							prefix = "ext:"
							prefix_policy = "NoPrefix"
						}
						groups = {
							claim = "groups"
							prefix = "ext-group:"
						}
					}
					validation_rules = [
						{
							claim = "aud"
							required_value = "my-app"
						},
						{
							claim = "iss"
							required_value = "https://auth.example.com"
						}
					]
				}
			}`)
			Expect(Terraform.Validate().ExitCode).To(BeZero())
		})

		It("validates multiple clients with different configurations", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						id = "confidential-client"
						secret = "client-secret"
					},
					{
						component = {
							name = "public-component"
							namespace = "default"
						}
					}
				]
			}`)
			Expect(Terraform.Validate().ExitCode).To(BeZero())
		})

		It("fails with multiple clients where one has ID but no secret", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						id = "confidential-client"
						secret = "client-secret"
					},
					{
						id = "invalid-client"
						# secret missing for this client
					}
				]
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Client secret is required when client ID is provided")
		})
	})

	Context("runtime validation", func() {
		It("fails when external auth is not enabled on cluster", func() {
			// Mock cluster calls for validation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testClusterRoute),
					RespondWithJSON(http.StatusOK, clusterWithoutExternalAuthTemplate()),
				),
			)

			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("External authentication configuration is not enabled")
		})

		It("fails with invalid cluster ID", func() {
			// Mock 404 for non-existent cluster
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/non-existent-cluster"),
					RespondWithJSON(http.StatusNotFound, `{"kind": "Error", "code": "CLUSTERS-MGMT-404", "reason": "Cluster 'non-existent-cluster' not found"}`),
				),
			)

			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "non-existent-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			// The error message may vary depending on cluster state check
		})
	})

	Context("import functionality", func() {
		It("fails with invalid import format", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Import("rhcs_external_auth_provider.test", "invalid-format")
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("External auth provider to import should be specified as <cluster_id>,<provider_id>")
		})

		It("fails with missing cluster ID", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Import("rhcs_external_auth_provider.test", ",provider-id")
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("External auth provider to import should be specified as <cluster_id>,<provider_id>")
		})

		It("fails with missing provider ID", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Import("rhcs_external_auth_provider.test", "cluster-id,")
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("External auth provider to import should be specified as <cluster_id>,<provider_id>")
		})

		It("succeeds with valid import format", func() {
			TestServer.AppendHandlers(
				// Mock external auth provider retrieval
				CombineHandlers(
					VerifyRequest(http.MethodGet, testProviderRoute),
					RespondWithJSON(http.StatusOK, externalAuthProviderTemplate()),
				),
			)

			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {}`)
			runOutput := Terraform.Import("rhcs_external_auth_provider.test", "test-cluster,test-provider")
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("configuration edge cases", func() {
		It("handles empty audiences list", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = []
				}
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("handles single audience", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["single-audience"]
				}
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("handles multiple audiences", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1", "audience2", "audience3"]
				}
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("validates complex claim mappings", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://auth.example.com"
					audiences = ["audience1"]
				}
				claim = {
					mappings = {
						username = {
							claim = "sub"
							prefix = "external:"
							prefix_policy = "Prefix"
						}
						groups = {
							claim = "groups"
							prefix = "external-group:"
						}
					}
					validation_rules = [
						{
							claim = "aud"
							required_value = "my-application"
						},
						{
							claim = "iss" 
							required_value = "https://auth.example.com"
						},
						{
							claim = "exp"
							required_value = "3600"
						}
					]
				}
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("validates empty claim configuration", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				claim = {}
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("validates client with component only", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						component = {
							name = "console"
							namespace = "openshift-console"
						}
					}
				]
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("validates client with extra scopes", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						id = "my-client"
						secret = "my-secret"
						extra_scopes = ["openid", "profile", "email", "groups"]
					}
				]
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("validates mixed client configurations", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [
					{
						id = "confidential-client"
						secret = "confidential-secret"
						extra_scopes = ["openid", "profile"]
					},
					{
						component = {
							name = "public-component"
							namespace = "openshift-auth"
						}
						extra_scopes = ["openid"]
					},
					{
						id = "another-client"
						secret = "another-secret"
					}
				]
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("URL validation", func() {
		It("fails with invalid protocol", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "ftp://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("issuer URL must use HTTPS")
		})

		It("fails with HTTP instead of HTTPS", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "http://insecure.example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("issuer URL must use HTTPS")
		})

		It("succeeds with HTTPS URL with path", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://auth.example.com/oauth2/default"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("succeeds with HTTPS URL with port", func() {
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "my_provider" {
				cluster = "test-cluster"
				id = "my-provider"
				issuer = {
					url = "https://auth.example.com:8443/oauth2"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Validate()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("CRUD operations", func() {
		It("creates external auth provider successfully", func() {
			// Mock cluster with external auth enabled (for validation)
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testClusterRoute),
					RespondWithJSON(http.StatusOK, clusterWithExternalAuthTemplate()),
				),
			)
			// Mock successful provider creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, testExternalAuthRoute),
					VerifyJQ(".id", "test-provider"),
					VerifyJQ(".issuer.url", "https://example.com"),
					VerifyJQ(".issuer.audiences[0]", "audience1"),
					RespondWithJSON(http.StatusCreated, externalAuthProviderTemplate()),
				),
			)
			// Mock provider read after creation (for state refresh)
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testProviderRoute),
					RespondWithJSON(http.StatusOK, externalAuthProviderTemplate()),
				),
			)

			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Verify resource was created with correct attributes
			resource := Terraform.Resource("rhcs_external_auth_provider", "test")
			Expect(resource).To(MatchJQ(".attributes.id", "test-provider"))
			Expect(resource).To(MatchJQ(".attributes.cluster", "test-cluster"))
			Expect(resource).To(MatchJQ(".attributes.issuer.url", "https://example.com"))
			Expect(resource).To(MatchJQ(".attributes.issuer.audiences[0]", "audience1"))
		})

		It("reads external auth provider state correctly", func() {
			// Mock cluster with external auth enabled
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testClusterRoute),
					RespondWithJSON(http.StatusOK, clusterWithExternalAuthTemplate()),
				),
			)
			// Mock successful provider creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, testExternalAuthRoute),
					RespondWithJSON(http.StatusCreated, externalAuthProviderTemplate()),
				),
			)
			// Mock provider read after creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testProviderRoute),
					RespondWithJSON(http.StatusOK, externalAuthProviderTemplate()),
				),
			)

			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)

			// Apply to create
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Verify resource state is populated correctly
			resource := Terraform.Resource("rhcs_external_auth_provider", "test")
			Expect(resource).To(MatchJQ(".attributes.id", "test-provider"))
			Expect(resource).To(MatchJQ(".attributes.cluster", "test-cluster"))
		})

		It("updates external auth provider successfully", func() {
			// Mock cluster with external auth enabled
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testClusterRoute),
					RespondWithJSON(http.StatusOK, clusterWithExternalAuthTemplate()),
				),
			)
			// Mock successful provider creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, testExternalAuthRoute),
					RespondWithJSON(http.StatusCreated, externalAuthProviderTemplate()),
				),
			)
			// Mock provider read after creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testProviderRoute),
					RespondWithJSON(http.StatusOK, externalAuthProviderTemplate()),
				),
			)
			// Mock successful provider update
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPatch, testProviderRoute),
					VerifyJQ(".issuer.url", "https://updated-example.com"),
					VerifyJQ(".issuer.audiences[0]", "audience1"),
					VerifyJQ(".issuer.audiences[1]", "audience2"),
					RespondWithJSON(http.StatusOK, updatedExternalAuthProviderTemplate()),
				),
			)
			// Mock provider read after update
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testProviderRoute),
					RespondWithJSON(http.StatusOK, updatedExternalAuthProviderTemplate()),
				),
			)

			// Apply initial config
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Apply updated config
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://updated-example.com"
					audiences = ["audience1", "audience2"]
				}
			}`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Verify updates were applied
			resource := Terraform.Resource("rhcs_external_auth_provider", "test")
			Expect(resource).To(MatchJQ(".attributes.issuer.url", "https://updated-example.com"))
			Expect(resource).To(MatchJQ(".attributes.issuer.audiences | length", 2))
		})

		It("deletes external auth provider successfully", func() {
			// Mock cluster with external auth enabled
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testClusterRoute),
					RespondWithJSON(http.StatusOK, clusterWithExternalAuthTemplate()),
				),
			)
			// Mock successful provider creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, testExternalAuthRoute),
					RespondWithJSON(http.StatusCreated, externalAuthProviderTemplate()),
				),
			)
			// Mock provider read after creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testProviderRoute),
					RespondWithJSON(http.StatusOK, externalAuthProviderTemplate()),
				),
			)
			// Mock successful provider deletion
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodDelete, testProviderRoute),
					RespondWithJSON(http.StatusNoContent, ""),
				),
			)

			// Apply config to create provider
			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
			}`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Destroy the provider
			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("creates external auth provider with console client", func() {
			// Mock cluster with external auth enabled
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testClusterRoute),
					RespondWithJSON(http.StatusOK, clusterWithExternalAuthTemplate()),
				),
			)
			// Mock successful provider creation with console client
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, testExternalAuthRoute),
					VerifyJQ(".id", "test-provider"),
					VerifyJQ(".issuer.url", "https://example.com"),
					VerifyJQ(".clients[0].id", "console-client"),
					VerifyJQ(".clients[0].secret", "console-secret"),
					RespondWithJSON(http.StatusCreated, externalAuthProviderWithConsoleTemplate()),
				),
			)
			// Mock provider read after creation
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, testProviderRoute),
					RespondWithJSON(http.StatusOK, externalAuthProviderWithConsoleTemplate()),
				),
			)

			Terraform.Source(`
			resource "rhcs_external_auth_provider" "test" {
				cluster = "test-cluster"
				id = "test-provider"
				issuer = {
					url = "https://example.com"
					audiences = ["audience1"]
				}
				clients = [{
					id = "console-client"
					secret = "console-secret"
				}]
			}`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Verify resource was created with console client
			resource := Terraform.Resource("rhcs_external_auth_provider", "test")
			Expect(resource).To(MatchJQ(".attributes.clients[0].id", "console-client"))
			Expect(resource).To(MatchJQ(".attributes.clients[0].secret", "console-secret"))
		})
	})
})
