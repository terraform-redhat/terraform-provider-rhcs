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

package classic

import (
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	. "github.com/onsi/gomega/ghttp"       // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("Classic trust policy external ID", func() {
	const externalID = "rhcs-subsystem-classic-external-id"

	It("creates cluster with trust_policy_external_id and refreshes it from the API", func() {
		spec, err := cmv1.NewCluster().
			ID("123").
			ExternalID("123").
			Name("my-cluster").
			AWS(cmv1.NewAWS().
				AccountID("123456789012").
				BillingAccountID("123456789012").
				SubnetIDs("id1", "id2", "id3").
				STS(cmv1.NewSTS().
					RoleARN("arn:aws:iam::123456789012:role/installer").
					SupportRoleARN("arn:aws:iam::123456789012:role/support").
					OperatorRolePrefix("test").
					ExternalID(externalID).
					InstanceIAMRoles(cmv1.NewInstanceIAMRoles().
						MasterRoleARN("arn:aws:iam::123456789012:role/master").
						WorkerRoleARN("arn:aws:iam::123456789012:role/worker")))).
			State(cmv1.ClusterStateReady).
			Region(cmv1.NewCloudRegion().ID("us-west-1")).
			MultiAZ(true).
			Properties(map[string]string{
				"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
				"rosa_tf_version":   build.Version,
				"rosa_tf_commit":    build.Commit,
			}).
			Nodes(cmv1.NewClusterNodes().
				Compute(3).
				AvailabilityZones("us-west-1a").
				ComputeMachineType(cmv1.NewMachineType().ID("r5.xlarge"))).
			Network(cmv1.NewNetwork().
				MachineCIDR("10.0.0.0/16").
				ServiceCIDR("172.30.0.0/16").
				PodCIDR("10.128.0.0/14").
				HostPrefix(23)).
			Version(cmv1.NewVersion().ChannelGroup("stable").
				Enabled(true).
				ROSAEnabled(true).
				ID("openshift-v4.10.0").
				RawID("4.10.0")).
			Build()
		Expect(err).NotTo(HaveOccurred())

		b := new(strings.Builder)
		Expect(cmv1.MarshalCluster(spec, b)).To(Succeed())
		template := b.String()

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
				RespondWithJSON(http.StatusOK, versionListPage1),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJQ(`.aws.sts.external_id`, externalID),
				RespondWithPatchedJSON(http.StatusCreated, template, `[
				{
				  "op": "add",
				  "path": "/aws",
				  "value": {
				    "sts": {
				      "role_arn": "arn:aws:iam::123456789012:role/installer",
				      "support_role_arn": "arn:aws:iam::123456789012:role/support",
				      "operator_role_prefix": "test",
				      "external_id": "`+externalID+`",
				      "instance_iam_roles": {
				        "master_role_arn": "arn:aws:iam::123456789012:role/master",
				        "worker_role_arn": "arn:aws:iam::123456789012:role/worker"
				      }
				    }
				  }
				}]`),
			),
		)

		Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123456789012"
			  sts = {
			    operator_role_prefix     = "test"
			    role_arn                 = "arn:aws:iam::123456789012:role/installer"
			    support_role_arn         = "arn:aws:iam::123456789012:role/support"
			    trust_policy_external_id = "` + externalID + `"
			    instance_iam_roles = {
			      master_role_arn = "arn:aws:iam::123456789012:role/master"
			      worker_role_arn = "arn:aws:iam::123456789012:role/worker"
			    }
			  }
			  aws_subnet_ids = ["id1", "id2", "id3"]
			  availability_zones = ["us-west-1a"]
			  version = "4.10.0"
			  wait_for_create_complete = false
			}
		`)

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster").(map[string]interface{})
		attributes := resource["attributes"].(map[string]interface{})
		stsBlock := attributes["sts"].(map[string]interface{})
		Expect(stsBlock["trust_policy_external_id"]).To(Equal(externalID))
	})
})
