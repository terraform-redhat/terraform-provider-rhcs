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

package hcp

import (
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"    // nolint
	. "github.com/onsi/gomega"       // nolint
	. "github.com/onsi/gomega/ghttp" // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Tuning Configs", func() {
	clusterSpecBuilder := cmv1.NewCluster().
		ID("123").
		Name("cluster").
		HREF("/api/clusters_mgmt/v1/clusters/123")
	spec, err := clusterSpecBuilder.Build()
	Expect(err).ToNot(HaveOccurred())
	b := new(strings.Builder)
	err = cmv1.MarshalCluster(spec, b)
	Expect(err).ToNot(HaveOccurred())
	clusterTemplate := b.String()

	clusterReadySpecBuilder := clusterSpecBuilder.State(cmv1.ClusterStateReady)
	spec, err = clusterReadySpecBuilder.Build()
	Expect(err).ToNot(HaveOccurred())
	b.Reset()
	err = cmv1.MarshalCluster(spec, b)
	Expect(err).ToNot(HaveOccurred())
	clusterReadyTemplate := b.String()

	tuningConfigBuilder := cmv1.NewTuningConfig().
		ID("456").
		HREF("/api/clusters_mgmt/v1/clusters/123/tuning_configs/456").
		Name("my_config").
		Spec(map[string]interface{}{})
	tuningConfig, err := tuningConfigBuilder.Build()
	Expect(err).ToNot(HaveOccurred())
	b.Reset()
	err = cmv1.MarshalTuningConfig(tuningConfig, b)
	Expect(err).ToNot(HaveOccurred())
	tuningConfigTemplate := b.String()

	tuningConfigSpecBuilder := cmv1.NewTuningConfig().
		ID("456").
		HREF("/api/clusters_mgmt/v1/clusters/123/tuning_configs/456").
		Name("my_config").
		Spec(map[string]interface{}{"key": "value"})
	tuningConfigSpec, err := tuningConfigSpecBuilder.Build()
	Expect(err).ToNot(HaveOccurred())
	b.Reset()
	err = cmv1.MarshalTuningConfig(tuningConfigSpec, b)
	Expect(err).ToNot(HaveOccurred())
	tuningConfigSpecTemplate := b.String()

	Context("tuning configs creation", func() {
		It("fails to find a matching cluster object", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "Cluster '123' not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
					cluster = "123"
					name = "my_config"
					spec = "{}"
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("fails if OCM backend fails to create the object", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReadyTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/tuning_configs"),
					RespondWithJSON(http.StatusInternalServerError, "Internal Server Error"),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
					cluster = "123"
					name = "my_config"
					spec = "{}"
				}
	    	`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("successfully creates a tuning_config object", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReadyTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/tuning_configs"),
					VerifyJQ(".name", "my_config"),
					VerifyJQ(".spec", map[string]interface{}{}),
					RespondWithJSON(http.StatusCreated, tuningConfigTemplate),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
					cluster = "123"
					name = "my_config"
					spec = "{}"
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})
	})

	Context("tuning configs importing", func() {
		It("fails if resource does not exist in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/tuning_configs/456"),
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "TuningConfig for cluster ID '123' is not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
				}
	    	`)
			Expect(terraform.Import("rhcs_tuning_config.tuning_config", "123,456")).ToNot(BeZero())
		})

		It("succeeds if resource exists in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/tuning_configs/456"),
					RespondWithJSON(http.StatusOK, tuningConfigTemplate),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
				}
	    	`)
			Expect(terraform.Import("rhcs_tuning_config.tuning_config", "123,456")).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_tuning_config", "tuning_config").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster": "123",
					"id":      "456",
					"name":    "my_config",
					"spec":    `{}`,
				},
			))
		})
	})

	Context("tuning configs updating", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReadyTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/tuning_configs"),
					VerifyJQ(".name", "my_config"),
					VerifyJQ(".spec", map[string]interface{}{}),
					RespondWithJSON(http.StatusCreated, tuningConfigTemplate),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
					cluster = "123"
					name = "my_config"
					spec = "{}"
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("successfully applies the changes in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/tuning_configs/456"),
					RespondWithJSON(http.StatusOK, tuningConfigTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/tuning_configs/456"),
					VerifyJQ(".spec", map[string]interface{}{"key": "value"}),
					RespondWithJSON(http.StatusOK, tuningConfigSpecTemplate),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
					cluster = "123"
					name = "my_config"
					spec = "{\"key\":\"value\"}"
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_tuning_config", "tuning_config").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster": "123",
					"id":      "456",
					"name":    "my_config",
					"spec":    `{"key":"value"}`,
				},
			))
		})
	})

	Context("tuning configs deletion", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReadyTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/tuning_configs"),
					VerifyJQ(".name", "my_config"),
					VerifyJQ(".spec", map[string]interface{}{}),
					RespondWithJSON(http.StatusCreated, tuningConfigTemplate),
				),
			)

			terraform.Source(`
				resource "rhcs_tuning_config" "tuning_config" {
					cluster = "123"
					name = "my_config"
					spec = "{}"
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("successfully applies the deletion in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/tuning_configs/456"),
					RespondWithJSON(http.StatusOK, tuningConfigTemplate),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/tuning_configs/456"),
					RespondWithJSON(http.StatusNoContent, "{}"),
				),
			)

			Expect(terraform.Destroy()).To(BeZero())
		})

	})
})
