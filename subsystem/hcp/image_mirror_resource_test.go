package hcp

import (
	"net/http"
	"strings"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	. "github.com/onsi/gomega/ghttp"       // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Image Mirror Resource", func() {
	Context("Image Mirror CRUD operations", func() {
		
		// Create a simple cluster template for testing
		var template string
		BeforeEach(func() {
			cluster, err := cmv1.NewCluster().
				ID("123").
				Name("test-cluster").
				State(cmv1.ClusterStateReady).
				Hypershift(cmv1.NewHypershift().Enabled(true)).
				Build()
			Expect(err).ToNot(HaveOccurred())
			
			var b strings.Builder
			err = cmv1.MarshalCluster(cluster, &b)
			Expect(err).ToNot(HaveOccurred())
			template = b.String()
		})
		
		It("Creates and manages image mirror for HCP cluster", func() {
			// Prepare the server for cluster validation:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/image_mirrors"),
					VerifyJQ(`.source`, "registry.example.com/team"),
					VerifyJQ(`.mirrors[0]`, "mirror.corp.com/team"),
					VerifyJQ(`.type`, "digest"),
					RespondWithJSON(http.StatusCreated, `{
						"id": "12345",
						"type": "digest",
						"source": "registry.example.com/team",
						"mirrors": ["mirror.corp.com/team", "backup.corp.com/team"],
						"creation_timestamp": "2024-01-01T00:00:00Z",
						"last_update_timestamp": "2024-01-01T00:00:00Z"
					}`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
			resource "rhcs_image_mirror" "corp_registry" {
				cluster_id = "123"
				type       = "digest"
				source     = "registry.example.com/team"
				mirrors    = [
					"mirror.corp.com/team",
					"backup.corp.com/team"
				]
			}`)
			
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			
			resource := Terraform.Resource("rhcs_image_mirror", "corp_registry")
			Expect(resource).To(MatchJQ(`.attributes.id`, "12345"))
			Expect(resource).To(MatchJQ(`.attributes.source`, "registry.example.com/team"))
			Expect(resource).To(MatchJQ(`.attributes.mirrors[0]`, "mirror.corp.com/team"))
		})


		It("Validates cluster is HCP enabled", func() {
			// Create non-HCP cluster template
			nonHcpCluster, err := cmv1.NewCluster().
				ID("456").
				Name("non-hcp-cluster").
				State(cmv1.ClusterStateReady).
				Hypershift(cmv1.NewHypershift().Enabled(false)).
				Build()
			Expect(err).ToNot(HaveOccurred())
			
			var b strings.Builder
			err = cmv1.MarshalCluster(nonHcpCluster, &b)
			Expect(err).ToNot(HaveOccurred())
			nonHcpTemplate := b.String()

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/456"),
					RespondWithJSON(http.StatusOK, nonHcpTemplate),
				),
			)

			Terraform.Source(`
			resource "rhcs_image_mirror" "non_hcp_cluster" {
				cluster_id = "456"
				source     = "registry.example.com/team"
				mirrors    = ["mirror.corp.com/team"]
			}`)
			
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Image mirrors are only supported on Hosted Control Plane clusters")
		})

		It("Validates cluster is in ready state", func() {
			// Create installing cluster template
			installingCluster, err := cmv1.NewCluster().
				ID("789").
				Name("installing-cluster").
				State(cmv1.ClusterStateInstalling).
				Hypershift(cmv1.NewHypershift().Enabled(true)).
				Build()
			Expect(err).ToNot(HaveOccurred())
			
			var b strings.Builder
			err = cmv1.MarshalCluster(installingCluster, &b)
			Expect(err).ToNot(HaveOccurred())
			installingTemplate := b.String()

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/789"),
					RespondWithJSON(http.StatusOK, installingTemplate),
				),
			)

			Terraform.Source(`
			resource "rhcs_image_mirror" "installing_cluster" {
				cluster_id = "789"
				source     = "registry.example.com/team"
				mirrors    = ["mirror.corp.com/team"]
			}`)
			
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("is not ready")
		})
	})
})