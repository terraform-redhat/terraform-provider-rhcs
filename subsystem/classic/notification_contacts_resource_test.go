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

package classic

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"                      // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_cluster_notification_contacts", func() {
	const clusterRoute = "/api/clusters_mgmt/v1/clusters/123"
	const ncRoute = "/api/accounts_mgmt/v1/subscriptions/sub-123/notification_contacts"

	clusterJSON := `{
		"kind": "Cluster",
		"id": "123",
		"href": "/api/clusters_mgmt/v1/clusters/123",
		"name": "my-cluster",
		"subscription": {
			"id": "sub-123",
			"href": "/api/accounts_mgmt/v1/subscriptions/sub-123"
		}
	}`

	ncListResponse := func(contacts ...string) string {
		items := ""
		for i, c := range contacts {
			if i > 0 {
				items += ","
			}
			items += `{"kind":"Account","id":"acc-` + c + `","username":"` + c + `"}`
		}
		return `{"kind":"AccountList","items":[` + items + `],"size":` + fmt.Sprintf("%d", len(contacts)) + `,"total":` + fmt.Sprintf("%d", len(contacts)) + `}`
	}

	It("Creates notification contacts for a cluster", func() {
		TestServer.AppendHandlers(
			// Create: GET cluster to find subscription ID
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			// Create: GET current contacts
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			// Create: POST user1
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				VerifyJQ(`.account_identifier`, "user1"),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-user1","username":"user1"}`),
			),
			// Create: GET to verify
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("user1")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["user1"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
		resource := Terraform.Resource("rhcs_cluster_notification_contacts", "nc")
		Expect(resource).To(MatchJQ(`.attributes.contacts`, []interface{}{"user1"}))
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
	})

	It("Fetches actual contacts on create partial failure", func() {
		TestServer.AppendHandlers(
			// Create: GET cluster
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			// Create: GET current contacts (empty)
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			// Create: POST alice succeeds
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				VerifyJQ(`.account_identifier`, "alice"),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-alice","username":"alice"}`),
			),
			// Create: POST bob fails
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusInternalServerError, `{"kind":"Error","id":"500","reason":"Internal error"}`),
			),
			// Create error path: fetch actual contacts to save partial state
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice", "bob"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("Can't set notification contacts")
		Expect(TestServer.ReceivedRequests()).To(HaveLen(5))
	})

	It("Retains resource with recovery contacts on create partial removal failure", func() {
		// Server already has [alice, bob, charlie]. User creates resource
		// with [alice]. Reconciliation keeps alice, deletes one stale
		// contact (success), then fails deleting the other.
		// State must retain the resource ID and recovery-fetch contacts.
		TestServer.AppendHandlers(
			// Create: GET cluster
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			// Create: UpdateNotificationContacts — GET current contacts
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob", "charlie")),
			),
			// alice is in desired set, no POST needed.
			// First DELETE (any stale contact) succeeds.
			CombineHandlers(
				VerifyRequest(http.MethodDelete, MatchRegexp(ncRoute+`/acc-`)),
				RespondWithJSON(http.StatusNoContent, ""),
			),
			// Second DELETE (other stale contact) fails.
			CombineHandlers(
				VerifyRequest(http.MethodDelete, MatchRegexp(ncRoute+`/acc-`)),
				RespondWithJSON(http.StatusForbidden, `{"kind":"Error","id":"403","reason":"forbidden"}`),
			),
			// Create error path: recovery fetch returns surviving contacts
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "charlie")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("Can't set notification contacts")
		resource := Terraform.Resource(
			"rhcs_cluster_notification_contacts", "nc",
		)
		Expect(resource).To(MatchJQ(`.attributes.id`, "123"))
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
		Expect(resource).To(MatchJQ(
			`.attributes.contacts`,
			[]interface{}{"alice", "charlie"},
		))
	})

	It("Retries when subscription ID is initially unavailable", func() {
		clusterNoSubJSON := `{
			"kind": "Cluster",
			"id": "123",
			"href": "/api/clusters_mgmt/v1/clusters/123",
			"name": "my-cluster"
		}`
		TestServer.AppendHandlers(
			// Create: first GET cluster — no subscription yet
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterNoSubJSON),
			),
			// Create: retry GET cluster — subscription available
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			// Normal Create flow
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-user1","username":"user1"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("user1")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["user1"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Updates notification contacts by adding a contact", func() {
		// Create with one contact
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-alice","username":"alice"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice"]
			}
		`)
		Expect(Terraform.Apply().ExitCode).To(BeZero())

		// Update: add bob
		TestServer.AppendHandlers(
			// Read for refresh
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
			// Update: cluster GET, then UpdateNotificationContacts (GET+POST), then FetchNotificationContacts (GET)
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				VerifyJQ(`.account_identifier`, "bob"),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-bob","username":"bob"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice", "bob"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
		resource := Terraform.Resource("rhcs_cluster_notification_contacts", "nc")
		Expect(resource).To(MatchJQ(`.attributes.contacts`, []interface{}{"alice", "bob"}))
	})

	It("Removes a notification contact", func() {
		// Create with two contacts
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				VerifyJQ(`.account_identifier`, "alice"),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-alice","username":"alice"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				VerifyJQ(`.account_identifier`, "bob"),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-bob","username":"bob"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice", "bob"]
			}
		`)
		Expect(Terraform.Apply().ExitCode).To(BeZero())

		// Update: remove bob
		TestServer.AppendHandlers(
			// Read for refresh
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob")),
			),
			// Update: cluster GET, then UpdateNotificationContacts (GET+DELETE), then FetchNotificationContacts (GET)
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob")),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, ncRoute+"/acc-bob"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
		resource := Terraform.Resource("rhcs_cluster_notification_contacts", "nc")
		Expect(resource).To(MatchJQ(`.attributes.contacts`, []interface{}{"alice"}))
	})

	It("Errors when update POST fails", func() {
		// Create with alice
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-alice","username":"alice"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice"]
			}
		`)
		Expect(Terraform.Apply().ExitCode).To(BeZero())

		// Update: add bob — POST fails
		post500 := CombineHandlers(
			VerifyRequest(http.MethodPost, ncRoute),
			RespondWithJSON(http.StatusInternalServerError, `{"kind":"Error","id":"500","reason":"Internal error"}`),
		)
		TestServer.AppendHandlers(
			// Read for refresh
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
			// Update: cluster GET, then UpdateNotificationContacts (GET + failed POST)
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
			post500,
			// Update error path: FetchNotificationContacts to sync state
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice", "bob"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("Can't update notification contacts")
	})

	It("Does not delete contacts when update POST fails", func() {
		// Create with alice and charlie
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				VerifyJQ(`.account_identifier`, "alice"),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-alice","username":"alice"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				VerifyJQ(`.account_identifier`, "charlie"),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-charlie","username":"charlie"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "charlie")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice", "charlie"]
			}
		`)
		Expect(Terraform.Apply().ExitCode).To(BeZero())

		// Update to [alice, bob]: add bob (fails), remove charlie (should be skipped)
		// No DELETE handler is registered — if code incorrectly tries to
		// delete charlie, ghttp will fail with an unhandled request.
		post500 := CombineHandlers(
			VerifyRequest(http.MethodPost, ncRoute),
			RespondWithJSON(http.StatusInternalServerError, `{"kind":"Error","id":"500","reason":"Internal error"}`),
		)
		TestServer.AppendHandlers(
			// Read for refresh
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "charlie")),
			),
			// Update: cluster GET, then UpdateNotificationContacts (GET + failed POST)
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "charlie")),
			),
			post500,
			// Update error path: FetchNotificationContacts to sync state
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "charlie")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice", "bob"]
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("Can't update notification contacts")
	})

	Context("Read error handling", func() {
		tfConfig := `
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["user1"]
			}
		`

		createHandlers := func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, clusterRoute),
					RespondWithJSON(http.StatusOK, clusterJSON),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, ncRoute),
					RespondWithJSON(http.StatusOK, ncListResponse()),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, ncRoute),
					RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-user1","username":"user1"}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, ncRoute),
					RespondWithJSON(http.StatusOK, ncListResponse("user1")),
				),
			)
		}

		It("Removes resource from state when cluster returns 404", func() {
			createHandlers()
			Terraform.Source(tfConfig)
			Expect(Terraform.Apply().ExitCode).To(BeZero())

			TestServer.AppendHandlers(
				// Read: cluster 404 -> RemoveResource
				CombineHandlers(
					VerifyRequest(http.MethodGet, clusterRoute),
					RespondWithJSON(http.StatusNotFound, `{"kind":"Error","id":"404","href":"/api/clusters_mgmt/v1/errors/404","code":"CLUSTERS-MGMT-404","reason":"Cluster '123' not found"}`),
				),
				// Re-create after RemoveResource (config still declares it)
				CombineHandlers(
					VerifyRequest(http.MethodGet, clusterRoute),
					RespondWithJSON(http.StatusOK, clusterJSON),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, ncRoute),
					RespondWithJSON(http.StatusOK, ncListResponse()),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, ncRoute),
					RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-user1","username":"user1"}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, ncRoute),
					RespondWithJSON(http.StatusOK, ncListResponse("user1")),
				),
			)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Errors when cluster returns 403", func() {
			createHandlers()
			Terraform.Source(tfConfig)
			Expect(Terraform.Apply().ExitCode).To(BeZero())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, clusterRoute),
					RespondWithJSON(http.StatusForbidden, `{"kind":"Error","id":"403","href":"/api/clusters_mgmt/v1/errors/403","code":"CLUSTERS-MGMT-403","reason":"Forbidden"}`),
				),
			)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).NotTo(BeZero())
			runOutput.VerifyErrorContainsSubstring("Can't find cluster")
		})

		It("Errors when cluster returns 500", func() {
			createHandlers()
			Terraform.Source(tfConfig)
			Expect(Terraform.Apply().ExitCode).To(BeZero())

			cluster500 := CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusInternalServerError, `{"kind":"Error","id":"500","href":"/api/clusters_mgmt/v1/errors/500","code":"CLUSTERS-MGMT-500","reason":"Internal error"}`),
			)
			// SDK retries GET 5xx up to 2 times (3 total attempts)
			TestServer.AppendHandlers(cluster500, cluster500, cluster500)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).NotTo(BeZero())
			runOutput.VerifyErrorContainsSubstring("Can't find cluster")
		})

		It("Errors when subscription ID is unavailable", func() {
			createHandlers()
			Terraform.Source(tfConfig)
			Expect(Terraform.Apply().ExitCode).To(BeZero())

			clusterNoSubJSON := `{
				"kind": "Cluster",
				"id": "123",
				"href": "/api/clusters_mgmt/v1/clusters/123",
				"name": "my-cluster"
			}`
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, clusterRoute),
					RespondWithJSON(http.StatusOK, clusterNoSubJSON),
				),
			)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).NotTo(BeZero())
			runOutput.VerifyErrorContainsSubstring("subscription ID is not available")
		})

		It("Errors when notification contacts fetch fails", func() {
			createHandlers()
			Terraform.Source(tfConfig)
			Expect(Terraform.Apply().ExitCode).To(BeZero())

			nc500 := CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusInternalServerError, `{"kind":"Error","id":"500","reason":"Internal error"}`),
			)
			// SDK retries GET 5xx up to 2 times (3 total attempts)
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, clusterRoute),
					RespondWithJSON(http.StatusOK, clusterJSON),
				),
				nc500, nc500, nc500,
			)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).NotTo(BeZero())
			runOutput.VerifyErrorContainsSubstring("Can't read notification contacts")
		})
	})

	It("Imports notification contacts", func() {
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["user1"]
			}
		`)
		// Import triggers ImportState (sets cluster_id + id) then Read
		TestServer.AppendHandlers(
			// Read: GET cluster
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			// Read: GET notification contacts
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("user1")),
			),
		)
		runOutput := Terraform.Import(
			"rhcs_cluster_notification_contacts.nc", "123",
		)
		Expect(runOutput.ExitCode).To(BeZero())
		resource := Terraform.Resource(
			"rhcs_cluster_notification_contacts", "nc",
		)
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
		Expect(resource).To(MatchJQ(`.attributes.id`, "123"))
		Expect(resource).To(MatchJQ(
			`.attributes.contacts`, []interface{}{"user1"},
		))
	})

	It("Destroys notification contacts", func() {
		// Create
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-user1","username":"user1"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("user1")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["user1"]
			}
		`)
		Expect(Terraform.Apply().ExitCode).To(BeZero())

		// Destroy
		TestServer.AppendHandlers(
			// Read for refresh
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("user1")),
			),
			// Delete
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("user1")),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, ncRoute+"/acc-user1"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)
		runOutput := Terraform.Destroy()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Retains resource in state on partial destroy failure", func() {
		// Create with two contacts
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-alice","username":"alice"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-bob","username":"bob"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice", "bob"]
			}
		`)
		Expect(Terraform.Apply().ExitCode).To(BeZero())

		// Destroy: first DELETE succeeds, second DELETE fails.
		// Handlers are consumed sequentially regardless of map iteration
		// order, so whichever contact is deleted first succeeds.
		TestServer.AppendHandlers(
			// Read for refresh
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob")),
			),
			// Delete: GET cluster, GET current contacts
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice", "bob")),
			),
			// First DELETE (any contact) succeeds
			CombineHandlers(
				VerifyRequest(http.MethodDelete, MatchRegexp(ncRoute+`/acc-`)),
				RespondWithJSON(http.StatusNoContent, ""),
			),
			// Second DELETE (other contact) fails
			CombineHandlers(
				VerifyRequest(http.MethodDelete, MatchRegexp(ncRoute+`/acc-`)),
				RespondWithJSON(http.StatusForbidden, `{"kind":"Error","id":"403","reason":"forbidden"}`),
			),
			// Error path: re-fetch returns the surviving contact
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("bob")),
			),
		)
		runOutput := Terraform.Destroy()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("Can't remove notification contacts")
		resource := Terraform.Resource(
			"rhcs_cluster_notification_contacts", "nc",
		)
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
		Expect(resource).To(MatchJQ(
			`.attributes.contacts`, []interface{}{"bob"},
		))
	})

	It("Errors on destroy when DELETE fails", func() {
		// Create with one contact
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse()),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, ncRoute),
				RespondWithJSON(http.StatusCreated, `{"kind":"Account","id":"acc-alice","username":"alice"}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
		)
		Terraform.Source(`
			resource "rhcs_cluster_notification_contacts" "nc" {
				cluster_id = "123"
				contacts   = ["alice"]
			}
		`)
		Expect(Terraform.Apply().ExitCode).To(BeZero())

		// Destroy: DELETE fails with 403
		TestServer.AppendHandlers(
			// Read for refresh
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
			// Delete: GET cluster, GET current contacts, DELETE fails
			CombineHandlers(
				VerifyRequest(http.MethodGet, clusterRoute),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, ncRoute+"/acc-alice"),
				RespondWithJSON(http.StatusForbidden, `{"kind":"Error","id":"403","reason":"forbidden"}`),
			),
			// Delete error path: fetch actual remaining contacts
			CombineHandlers(
				VerifyRequest(http.MethodGet, ncRoute),
				RespondWithJSON(http.StatusOK, ncListResponse("alice")),
			),
		)
		runOutput := Terraform.Destroy()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("Can't remove notification contacts")
		resource := Terraform.Resource(
			"rhcs_cluster_notification_contacts", "nc",
		)
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
		Expect(resource).To(MatchJQ(
			`.attributes.contacts`, []interface{}{"alice"},
		))
	})
})
