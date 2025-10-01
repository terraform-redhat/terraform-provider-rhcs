package hcp

import (
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("External Auth Provider", func() {
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
		It("will fail at runtime since CRUD operations are not implemented", func() {
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
			runOutput.VerifyErrorContainsSubstring("Create operation will be implemented in OCM-18658")
		})
	})
})
