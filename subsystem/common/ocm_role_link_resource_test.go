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

package common

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing"
)

const (
	orgID        = "1a2b3c4d5e6f"
	roleArn1     = "arn:aws:iam::123456789012:role/ManagedOpenshift-OCM-Role"
	roleArn2     = "arn:aws:iam::987654321098:role/ManagedOpenshift-OCM-Role"
	duplicateArn = "arn:aws:iam::123456789012:role/ManagedOpenshift-OCM-Role-Different"
)

var _ = Describe("OCM Role Link", func() {
	Context("Create", func() {
		It("successfully links a role to organization", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			label, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{
						"kind": "Error",
						"id": "404",
						"href": "/api/accounts_mgmt/v1/errors/404",
						"code": "ACCOUNT-MGMT-404",
						"reason": "Label not found"
					}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels"),
					VerifyJQ(".key", "sts_ocm_role"),
					VerifyJQ(".value", roleArn1),
					RespondWithOcmObjectMarshal(http.StatusCreated, label, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := testCtx.Terraform.Resource("rhcs_ocm_role_link", "ocm_role")
			Expect(resource).To(MatchJQ(".attributes.role_arn", roleArn1))
			Expect(resource).To(MatchJQ(".attributes.organization_id", orgID))
		})

		It("successfully links when role is already linked (idempotent)", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			existingLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, existingLabel, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("successfully links additional role from different AWS account", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			existingLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			updatedLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1 + "," + roleArn2).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, existingLabel, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(
						http.MethodPatch,
						"/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role",
					),
					VerifyJQ(".value", roleArn1+","+roleArn2),
					RespondWithOcmObjectMarshal(http.StatusOK, updatedLabel, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn2 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("fails when linking second role from same AWS account", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			existingLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, existingLabel, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + duplicateArn + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("already has role from AWS account")
			runOutput.VerifyErrorContainsSubstring("123456789012")
		})

		It("fails with invalid ARN format", func() {
			// No handlers needed - validation happens client-side before API calls

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "invalid-arn"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("role_arn must be a valid AWS IAM role ARN")
		})

		It("fails when unable to get current account", func() {
			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithJSON(http.StatusUnauthorized, `{
						"kind": "Error",
						"id": "401",
						"href": "/api/accounts_mgmt/v1/errors/401",
						"code": "ACCOUNT-MGMT-401",
						"reason": "Unauthorized"
					}`),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Failed to get current account")
		})
	})

	Context("Read", func() {
		It("reads existing linked role", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			label, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels"),
					RespondWithOcmObjectMarshal(http.StatusCreated, label, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource := testCtx.Terraform.Resource("rhcs_ocm_role_link", "ocm_role")
			Expect(resource).To(MatchJQ(".attributes.role_arn", roleArn1))
		})

		It("removes from state when role is unlinked externally", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			label, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels"),
					RespondWithOcmObjectMarshal(http.StatusCreated, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels"),
					RespondWithOcmObjectMarshal(http.StatusCreated, label, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("removes from state when role is removed from label but others remain", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			multipleRolesLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1 + "," + roleArn2).
				Build()
			Expect(err).ToNot(HaveOccurred())

			roleRemovedLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn2).
				Build()
			Expect(err).ToNot(HaveOccurred())

			updatedLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn2 + "," + roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, multipleRolesLabel, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, roleRemovedLabel, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, roleRemovedLabel, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(
						http.MethodPatch,
						"/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role",
					),
					RespondWithOcmObjectMarshal(http.StatusOK, updatedLabel, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("Delete", func() {
		It("successfully unlinks role when it's the only one", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			label, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels"),
					RespondWithOcmObjectMarshal(http.StatusCreated, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(
						http.MethodDelete,
						"/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role",
					),
					RespondWithJSON(http.StatusNoContent, `{}`),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = testCtx.Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("successfully unlinks role when multiple roles exist", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			multipleRolesLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1 + "," + roleArn2).
				Build()
			Expect(err).ToNot(HaveOccurred())

			updatedLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn2).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, multipleRolesLabel, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, multipleRolesLabel, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, multipleRolesLabel, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(
						http.MethodPatch,
						"/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role",
					),
					VerifyJQ(".value", roleArn2),
					RespondWithOcmObjectMarshal(http.StatusOK, updatedLabel, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = testCtx.Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("succeeds when role is already unlinked", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			label, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels"),
					RespondWithOcmObjectMarshal(http.StatusCreated, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = testCtx.Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("fails with permission error when user is not org admin", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			label, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels"),
					VerifyJQ(".key", "sts_ocm_role"),
					VerifyJQ(".value", roleArn1),
					RespondWithOcmObjectMarshal(http.StatusCreated, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusForbidden, `{
						"kind": "Error",
						"id": "403",
						"href": "/api/accounts_mgmt/v1/errors/403",
						"code": "ACCOUNT-MGMT-403",
						"reason": "Forbidden"
					}`),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = testCtx.Terraform.Destroy()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Insufficient permissions")
			runOutput.VerifyErrorContainsSubstring("organization admin")
		})
	})

	Context("Import", func() {
		It("successfully imports existing link", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			label, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, label, amsv1.MarshalLabel),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, label, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Import("rhcs_ocm_role_link.ocm_role", roleArn1)
			Expect(runOutput.ExitCode).To(BeZero())

			resource := testCtx.Terraform.Resource("rhcs_ocm_role_link", "ocm_role")
			Expect(resource).To(MatchJQ(".attributes.role_arn", roleArn1))
			Expect(resource).To(MatchJQ(".attributes.organization_id", orgID))
		})

		It("fails with invalid ARN", func() {
			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Import("rhcs_ocm_role_link.ocm_role", "invalid-arn")
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Invalid role ARN")
		})

		It("fails when role is not linked", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithJSON(http.StatusNotFound, `{
						"kind": "Error",
						"id": "404",
						"href": "/api/accounts_mgmt/v1/errors/404",
						"code": "ACCOUNT-MGMT-404",
						"reason": "Label not found"
					}`),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Import("rhcs_ocm_role_link.ocm_role", roleArn1)
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Role not linked")
		})

		It("fails when role is not in the label", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			otherLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn2).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, otherLabel, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + roleArn1 + `"
				}
			`)

			runOutput := testCtx.Terraform.Import("rhcs_ocm_role_link.ocm_role", roleArn1)
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Role not linked")
		})

		It("fails when AWS account already has a different role linked", func() {
			account, err := amsv1.NewAccount().
				ID("account-123").
				Organization(
					amsv1.NewOrganization().
						ID(orgID).
						Name("Test Org"),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())

			existingLabel, err := amsv1.NewLabel().
				Key("sts_ocm_role").
				Value(roleArn1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			testCtx.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
					RespondWithOcmObjectMarshal(http.StatusOK, account, amsv1.MarshalAccount),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/organizations/"+orgID+"/labels/sts_ocm_role"),
					RespondWithOcmObjectMarshal(http.StatusOK, existingLabel, amsv1.MarshalLabel),
				),
			)

			testCtx.Terraform.Source(`
				resource "rhcs_ocm_role_link" "ocm_role" {
					role_arn = "` + duplicateArn + `"
				}
			`)

			runOutput := testCtx.Terraform.Import("rhcs_ocm_role_link.ocm_role", duplicateArn)
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("AWS account conflict detected")
			runOutput.VerifyErrorContainsSubstring("123456789012")
		})
	})
})
