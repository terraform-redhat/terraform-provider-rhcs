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
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

const (
	ocmRoleCurrentAccountResponse = `{
		"id": "1234567890",
		"href": "/api/accounts_mgmt/v1/accounts/1234567890",
		"first_name": "Test",
		"last_name": "User",
		"username": "testuser",
		"email": "testuser@example.com",
		"organization": {
			"id": "org123",
			"external_id": "ext-org-456",
			"name": "Test Org"
		}
	}`

	ocmRoleEmptyLabelsResponse = `{
		"kind": "LabelList",
		"page": 1,
		"size": 0,
		"total": 0,
		"items": []
	}`

	ocmRoleLinkedLabelResponse = `{
		"kind": "Label",
		"id": "label-abc-123",
		"href": "/api/accounts_mgmt/v1/organizations/org123/labels/label-abc-123",
		"key": "sts_ocm_role",
		"value": "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
	}`

	ocmRoleExistingLabelListResponse = `{
		"kind": "LabelList",
		"page": 1,
		"size": 1,
		"total": 1,
		"items": [
			{
				"kind": "Label",
				"id": "label-abc-123",
				"href": "/api/accounts_mgmt/v1/organizations/org123/labels/label-abc-123",
				"key": "sts_ocm_role",
				"value": "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
			}
		]
	}`

	ocmRoleDifferentRoleSameAccountResponse = `{
		"kind": "LabelList",
		"page": 1,
		"size": 1,
		"total": 1,
		"items": [
			{
				"kind": "Label",
				"id": "label-other",
				"key": "sts_ocm_role",
				"value": "arn:aws:iam::123456789012:role/other-role"
			}
		]
	}`
)

var _ = Describe("OCM role link resource", func() {

	It("Can create and link an OCM role", func() {
		TestServer.AppendHandlers(
			// Create: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Create: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleEmptyLabelsResponse),
			),
			// Create: POST label
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				VerifyJQ(`.key`, "sts_ocm_role"),
				VerifyJQ(`.value`, "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"),
				RespondWithJSON(http.StatusCreated, ocmRoleLinkedLabelResponse),
			),
			// Destroy plan Read: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Destroy plan Read: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleExistingLabelListResponse),
			),
			// Delete: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Delete: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleExistingLabelListResponse),
			),
			// Delete: DELETE label
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/accounts_mgmt/v1/organizations/org123/labels/sts_ocm_role"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("rhcs_rosa_ocm_role_link", "ocm_role")
		Expect(resource).To(MatchJQ(".attributes.role_arn", "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"))
		Expect(resource).To(MatchJQ(".attributes.id", "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"))

		Expect(Terraform.Destroy().ExitCode).To(BeZero())
	})

	It("Resolves organization_id from current account", func() {
		anotherRoleLabelList := `{
			"kind": "LabelList",
			"page": 1, "size": 1, "total": 1,
			"items": [{
				"kind": "Label",
				"id": "label-another",
				"key": "sts_ocm_role",
				"value": "arn:aws:iam::999999999999:role/another-role"
			}]
		}`
		TestServer.AppendHandlers(
			// Create: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Create: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleEmptyLabelsResponse),
			),
			// Create: POST label
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				VerifyJQ(`.key`, "sts_ocm_role"),
				VerifyJQ(`.value`, "arn:aws:iam::999999999999:role/another-role"),
				RespondWithJSON(http.StatusCreated, `{
					"kind": "Label",
					"id": "label-another",
					"key": "sts_ocm_role",
					"value": "arn:aws:iam::999999999999:role/another-role"
				}`),
			),
			// Destroy plan Read: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Destroy plan Read: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, anotherRoleLabelList),
			),
			// Delete: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Delete: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, anotherRoleLabelList),
			),
			// Delete: DELETE label
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/accounts_mgmt/v1/organizations/org123/labels/sts_ocm_role"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::999999999999:role/another-role"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("rhcs_rosa_ocm_role_link", "ocm_role")
		Expect(resource).To(MatchJQ(".attributes.role_arn", "arn:aws:iam::999999999999:role/another-role"))

		Expect(Terraform.Destroy().ExitCode).To(BeZero())
	})

	It("Detects drift when the link is deleted externally", func() {
		TestServer.AppendHandlers(
			// Create: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Create: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleEmptyLabelsResponse),
			),
			// Create: POST label
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				VerifyJQ(`.key`, "sts_ocm_role"),
				VerifyJQ(`.value`, "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"),
				RespondWithJSON(http.StatusCreated, ocmRoleLinkedLabelResponse),
			),
			// Plan Read: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Plan Read: findLabel (label gone → drift)
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleEmptyLabelsResponse),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		runOutput = Terraform.Run("plan", "-detailed-exitcode")
		Expect(runOutput.ExitCode).To(Equal(2))
	})

	It("Fails create when the role is already linked", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleExistingLabelListResponse),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("already linked")
		runOutput.VerifyErrorContainsSubstring("terraform import")
	})

	It("Rejects invalid ARN format at plan time", func() {
		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "not-a-valid-arn"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("must be a valid AWS IAM role ARN")
	})

	It("Rejects ARN with commas at plan time", func() {
		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::123456789012:role/bad,name"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("must not contain commas")
	})

	It("Appends an OCM role to an existing list", func() {
		appendedLabelList := `{
			"kind": "LabelList",
			"page": 1, "size": 1, "total": 1,
			"items": [{
				"kind": "Label",
				"id": "label-existing",
				"key": "sts_ocm_role",
				"value": "arn:aws:iam::111111111111:role/existing-role,arn:aws:iam::222222222222:role/new-role"
			}]
		}`
		TestServer.AppendHandlers(
			// Create: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Create: findLabel (label exists with different account's role)
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, `{
					"kind": "LabelList",
					"page": 1, "size": 1, "total": 1,
					"items": [{
						"kind": "Label",
						"id": "label-existing",
						"key": "sts_ocm_role",
						"value": "arn:aws:iam::111111111111:role/existing-role"
					}]
				}`),
			),
			// Create: PATCH label (append new ARN)
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/accounts_mgmt/v1/organizations/org123/labels/sts_ocm_role"),
				VerifyJQ(`.value`, "arn:aws:iam::111111111111:role/existing-role,arn:aws:iam::222222222222:role/new-role"),
				RespondWithJSON(http.StatusOK, `{
					"kind": "Label",
					"id": "label-existing",
					"key": "sts_ocm_role",
					"value": "arn:aws:iam::111111111111:role/existing-role,arn:aws:iam::222222222222:role/new-role"
				}`),
			),
			// Destroy plan Read: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Destroy plan Read: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, appendedLabelList),
			),
			// Delete: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Delete: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, appendedLabelList),
			),
			// Delete: PATCH label (remove our ARN, keep existing)
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/accounts_mgmt/v1/organizations/org123/labels/sts_ocm_role"),
				VerifyJQ(`.value`, "arn:aws:iam::111111111111:role/existing-role"),
				RespondWithJSON(http.StatusOK, `{
					"kind": "Label",
					"id": "label-existing",
					"key": "sts_ocm_role",
					"value": "arn:aws:iam::111111111111:role/existing-role"
				}`),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::222222222222:role/new-role"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		Expect(Terraform.Destroy().ExitCode).To(BeZero())
	})

	It("Succeeds unlink when role has already been removed", func() {
		otherRoleLabelList := `{
			"kind": "LabelList",
			"page": 1, "size": 1, "total": 1,
			"items": [{
				"kind": "Label",
				"id": "label-other",
				"key": "sts_ocm_role",
				"value": "arn:aws:iam::999999999999:role/other-role"
			}]
		}`
		TestServer.AppendHandlers(
			// Create: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Create: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleEmptyLabelsResponse),
			),
			// Create: POST label
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				VerifyJQ(`.key`, "sts_ocm_role"),
				RespondWithJSON(http.StatusCreated, ocmRoleLinkedLabelResponse),
			),
			// Plan Read: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Plan Read: findLabel (our ARN is gone → drift)
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, otherRoleLabelList),
			),
			// Destroy Read: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Destroy Read: findLabel (still gone → state already removed)
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, otherRoleLabelList),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		runOutput = Terraform.Run("plan", "-detailed-exitcode")
		Expect(runOutput.ExitCode).To(Equal(2))

		Expect(Terraform.Destroy().ExitCode).To(BeZero())
	})

	It("Rejects create when same AWS account already has a linked role", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleDifferentRoleSameAccountResponse),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::123456789012:role/new-role"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("already has a linked role")
		runOutput.VerifyErrorContainsSubstring("Only one role per AWS account")
	})

	It("Triggers replace when role_arn changes", func() {
		TestServer.AppendHandlers(
			// Create: resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Create: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleEmptyLabelsResponse),
			),
			// Create: POST label
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				VerifyJQ(`.key`, "sts_ocm_role"),
				VerifyJQ(`.value`, "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"),
				RespondWithJSON(http.StatusCreated, ocmRoleLinkedLabelResponse),
			),
			// Plan Read (for the changed source): resolveOrgID
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
			// Plan Read: findLabel
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/org123/labels"),
				RespondWithJSON(http.StatusOK, ocmRoleExistingLabelListResponse),
			),
		)

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
		}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		Terraform.Source(`
		resource "rhcs_rosa_ocm_role_link" "ocm_role" {
			role_arn = "arn:aws:iam::999999999999:role/different-role"
		}
		`)
		runOutput = Terraform.Run("plan", "-detailed-exitcode")
		Expect(runOutput.ExitCode).To(Equal(2))
		runOutput.VerifyOutputContainsSubstring("forces replacement")
	})
})
