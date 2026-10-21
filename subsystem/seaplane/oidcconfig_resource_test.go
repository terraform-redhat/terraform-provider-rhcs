/*
Copyright (c) 2026 Red Hat, Inc.

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
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_oidcconfig_seaplane", func() {
	It("creates an OIDC config with its client-selected name", func() {
		server := NewServer()
		DeferCleanup(server.Close)

		name := "test-cluster.oidc"
		response := fmt.Sprintf(`{
			"metadata": {
				"uid": "test-oidc-config-id",
				"name": %q,
				"creationTimestamp": "2024-01-01T00:00:00Z"
			},
			"spec": {
				"type": "managed",
				"issuerUrl": "https://issuer.example.com"
			},
			"status": {
				"phase": "Pending"
			}
		}`, name)
		header := http.Header{"Content-Type": []string{"application/json"}}
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/v0/oidc_configs"),
				verifyJQ(".metadata.name", name),
				verifyJQ(".spec.type", "managed"),
				RespondWith(http.StatusCreated, response, header),
			),
		)

		Terraform.Source(fmt.Sprintf(`
			provider "rhcs" {
				alias          = "sp"
							seaplane_url = "%s"
				aws_account_id = "123456789012"
				aws_region     = "us-east-1"
			}

			resource "rhcs_oidcconfig_seaplane" "test" {
				provider = rhcs.sp
				name     = %q
				type     = "managed"
			}
		`, server.URL(), name))

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})
})
