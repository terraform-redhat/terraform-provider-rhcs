/*
Copyright (c) 2023 Red Hat, Inc.

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

package kubeletconfig

import (
	"context"
	"fmt"

	tfResources "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-online/ocm-common/pkg/ocm/client/test"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"go.uber.org/mock/gomock"
)

const (
	clusterId             = "myClusterId"
	name                  = "myKubenetConfigName"
	id                    = "myKubenetConfigId"
	createPidsLimit int64 = 5000
	updatePidsLimit int64 = 10000
)

var _ = Describe("KubeletConfig Resource", func() {

	var resource KubeletConfigResource
	var ctx context.Context
	var configsClient *test.MockKubeletConfigsClient
	var clusterWait *common.MockClusterWait
	var clusterClient *common.MockClusterClient
	var kubeletConfig, returnedKubeletConfig *v1.KubeletConfig
	var err error

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl := gomock.NewController(GinkgoT())
		clusterClient = common.NewMockClusterClient(ctrl)
		configsClient = test.NewMockKubeletConfigsClient(ctrl)
		clusterWait = common.NewMockClusterWait(ctrl)
		resource = KubeletConfigResource{
			configsClient: configsClient,
			clusterWait:   clusterWait,
			clusterClient: clusterClient,
		}
		kubeletConfig, err = createKubeletConfig(createPidsLimit)
		Expect(err).NotTo(HaveOccurred())
		returnedKubeletConfig, err = createReturnedKubeletConfig(createPidsLimit, name, id)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Schema", func() {
		It("Returns the correct schema", func() {
			request := tfResources.SchemaRequest{}
			response := &tfResources.SchemaResponse{}

			resource.Schema(ctx, request, response)

			schema := response.Schema
			Expect(schema.Attributes).To(HaveKey("cluster"))
			Expect(schema.Attributes).To(HaveKey("pod_pids_limit"))

			cluster := schema.Attributes["cluster"]
			Expect(cluster.IsRequired()).To(BeTrue())
			Expect(cluster.IsOptional()).To(BeFalse())
			Expect(cluster.IsComputed()).To(BeFalse())
			Expect(cluster.GetType().String()).To(Equal("basetypes.StringType"))

			podPidsLimit := schema.Attributes["pod_pids_limit"]
			Expect(podPidsLimit.IsRequired()).To(BeTrue())
			Expect(podPidsLimit.IsOptional()).To(BeFalse())
			Expect(podPidsLimit.IsComputed()).To(BeFalse())
			Expect(podPidsLimit.GetType().String()).To(Equal("basetypes.Int64Type"))
		})
	})

	Context("Metadata", func() {
		It("Returns the correct metatdata", func() {
			request := tfResources.MetadataRequest{
				ProviderTypeName: "rhcs",
			}
			response := &tfResources.MetadataResponse{}

			resource.Metadata(ctx, request, response)
			Expect(response.TypeName).To(Equal("rhcs_kubeletconfig"))
		})
	})

	Context("Create", func() {

		waitTimeoutInMinutes := int64(60)

		var plan tfsdk.Plan
		var state tfsdk.State
		var classicCluster, hcpCluster *cmv1.Cluster
		var err error
		var request tfResources.CreateRequest
		var response *tfResources.CreateResponse

		BeforeEach(func() {
			plan = createPlan(ctx, resource, createPidsLimit)
			state = createState(ctx, resource)
			classicCluster, err = createCluster(false)
			Expect(err).NotTo(HaveOccurred())
			hcpCluster, err = createCluster(true)
			Expect(err).NotTo(HaveOccurred())
			request = tfResources.CreateRequest{
				Plan: plan,
			}
			response = &tfResources.CreateResponse{
				State: state,
			}
		})

		It("Creates KubeletConfig", func() {
			clusterWait.EXPECT().WaitForClusterToBeReady(gomock.Eq(ctx), gomock.Eq(clusterId), waitTimeoutInMinutes).Return(&cmv1.Cluster{}, nil)
			clusterClient.EXPECT().FetchCluster(gomock.Any(), gomock.Eq(clusterId)).Return(classicCluster, nil)
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).Return([]*v1.KubeletConfig{}, false, nil)
			configsClient.EXPECT().Create(
				gomock.Eq(ctx), gomock.Eq(clusterId), test.MatchKubeletConfig(kubeletConfig)).Return(returnedKubeletConfig, nil)

			resource.Create(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(0))
		})

		It("Does not create KubeletConfig if the cluster is not ready", func() {
			clusterWait.EXPECT().WaitForClusterToBeReady(gomock.Eq(ctx), gomock.Eq(clusterId), waitTimeoutInMinutes).Return(
				nil, fmt.Errorf("cluster is not ready"))
			clusterClient.EXPECT().FetchCluster(gomock.Any(), gomock.Eq(clusterId)).Return(hcpCluster, nil)

			resource.Create(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(1))
		})

		It("Does not create KubeletConfig if it already exists in classic cluster", func() {
			clusterWait.EXPECT().WaitForClusterToBeReady(gomock.Eq(ctx), gomock.Eq(clusterId), waitTimeoutInMinutes).Return(&cmv1.Cluster{}, nil)
			clusterClient.EXPECT().FetchCluster(gomock.Any(), gomock.Eq(clusterId)).Return(classicCluster, nil)
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).
				Return([]*v1.KubeletConfig{returnedKubeletConfig}, true, nil)

			resource.Create(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(1))
		})

		It("Creates the second KubeletConfig for HCP cluster", func() {
			clusterWait.EXPECT().WaitForClusterToBeReady(gomock.Eq(ctx), gomock.Eq(clusterId), waitTimeoutInMinutes).Return(&cmv1.Cluster{}, nil)
			clusterClient.EXPECT().FetchCluster(gomock.Any(), gomock.Eq(clusterId)).Return(hcpCluster, nil)
			configsClient.EXPECT().Create(
				gomock.Eq(ctx), gomock.Eq(clusterId), test.MatchKubeletConfig(kubeletConfig)).Return(returnedKubeletConfig, nil)

			resource.Create(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(0))
		})

		It("Fails the plan if cannot create KubeletConfig", func() {
			clusterWait.EXPECT().WaitForClusterToBeReady(gomock.Eq(ctx), gomock.Eq(clusterId), waitTimeoutInMinutes).Return(&cmv1.Cluster{}, nil)
			clusterClient.EXPECT().FetchCluster(gomock.Any(), gomock.Eq(clusterId)).Return(classicCluster, nil)
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).
				Return([]*v1.KubeletConfig{returnedKubeletConfig}, true, nil)

			resource.Create(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(1))
		})
	})

	Context("Read", func() {

		var request tfResources.ReadRequest
		var response *tfResources.ReadResponse
		var state tfsdk.State
		var kubeletConfig *v1.KubeletConfig
		var err error

		BeforeEach(func() {

			state = createState(ctx, resource)
			kubeletConfig, err = createKubeletConfig(createPidsLimit)
			Expect(err).NotTo(HaveOccurred())

			request = tfResources.ReadRequest{
				State: state,
			}

			response = &tfResources.ReadResponse{
				State: state,
			}
		})

		It("Reads the existing KubeletConfig", func() {
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).
				Return([]*v1.KubeletConfig{returnedKubeletConfig}, true, nil)
			configsClient.EXPECT().Exists(ctx, clusterId, id).Return(true, kubeletConfig, nil)

			resource.Read(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(0))
		})

		It("Fails to read the KubeletConfig if it does not exist", func() {
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).Return([]*v1.KubeletConfig{}, true, nil)

			resource.Read(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(1))
		})
	})

	Context("Update", func() {

		var request tfResources.UpdateRequest
		var response *tfResources.UpdateResponse
		var plan tfsdk.Plan
		var state tfsdk.State
		var kubeletConfig *v1.KubeletConfig
		var err error

		BeforeEach(func() {
			plan = createPlan(ctx, resource, updatePidsLimit)
			state = createState(ctx, resource)
			request = tfResources.UpdateRequest{
				Plan:  plan,
				State: state,
			}
			response = &tfResources.UpdateResponse{
				State: state,
			}
			kubeletConfig, err = createKubeletConfig(updatePidsLimit)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Successfully updates KubeletConfig", func() {
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).
				Return([]*v1.KubeletConfig{returnedKubeletConfig}, true, nil)
			configsClient.EXPECT().Get(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Eq(id)).Return(
				kubeletConfig, nil)
			configsClient.EXPECT().Update(
				gomock.Eq(ctx), gomock.Eq(clusterId), test.MatchKubeletConfig(kubeletConfig)).Return(kubeletConfig, nil)

			resource.Update(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(0))
		})

		It("Fails to update a KubeletConfig that does not exist", func() {
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).Return([]*v1.KubeletConfig{}, true, nil)

			resource.Update(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(1))
		})

		It("Fails to update a KubeletConfig", func() {
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).
				Return([]*v1.KubeletConfig{returnedKubeletConfig}, true, nil)
			configsClient.EXPECT().Get(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Eq(id)).Return(
				kubeletConfig, nil)
			configsClient.EXPECT().Update(
				gomock.Eq(ctx), gomock.Eq(clusterId), test.MatchKubeletConfig(kubeletConfig)).Return(
				nil, fmt.Errorf("failed to update kubeletconfig"))

			resource.Update(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(1))
		})
	})

	Context("Delete", func() {

		var request tfResources.DeleteRequest
		var response *tfResources.DeleteResponse
		var state tfsdk.State

		BeforeEach(func() {
			state = createState(ctx, resource)
			request = tfResources.DeleteRequest{
				State: state,
			}

			response = &tfResources.DeleteResponse{
				State: state,
			}
		})

		It("Deletes the existing KubeletConfig", func() {
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).
				Return([]*v1.KubeletConfig{returnedKubeletConfig}, true, nil)
			configsClient.EXPECT().Delete(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Eq(id)).Return(nil)

			resource.Delete(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(0))
		})

		It("Fails to delete a KubeletConfig", func() {
			configsClient.EXPECT().List(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Any()).
				Return([]*v1.KubeletConfig{returnedKubeletConfig}, true, nil)
			configsClient.EXPECT().Delete(gomock.Eq(ctx), gomock.Eq(clusterId), gomock.Eq(id)).Return(
				fmt.Errorf("failed to delete KubeletConfig"))

			resource.Delete(ctx, request, response)
			Expect(response.Diagnostics.ErrorsCount()).To(Equal(1))
		})
	})
})

func createKubeletConfig(requiredPidsLimit int64) (*v1.KubeletConfig, error) {
	builder := v1.KubeletConfigBuilder{}
	return builder.PodPidsLimit(int(requiredPidsLimit)).Build()
}

func createReturnedKubeletConfig(requiredPidsLimit int64, name string, id string) (*v1.KubeletConfig, error) {
	builder := v1.KubeletConfigBuilder{}
	return builder.PodPidsLimit(int(requiredPidsLimit)).ID(id).Name(name).Build()
}

func createState(ctx context.Context, resource KubeletConfigResource) tfsdk.State {
	request := tfResources.SchemaRequest{}
	response := &tfResources.SchemaResponse{}
	resource.Schema(ctx, request, response)

	cluster, err := types.StringValue(clusterId).ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())
	pids, err := types.Int64Unknown().ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())
	configId, err := types.StringUnknown().ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())
	configName, err := types.StringUnknown().ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())

	state := map[string]tftypes.Value{
		"cluster":        cluster,
		"pod_pids_limit": pids,
		"id":             configId,
		"name":           configName,
	}

	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.Object{}, state),
		Schema: response.Schema,
	}
}

func createPlan(ctx context.Context, resource KubeletConfigResource, requiredPidsLimit int64) tfsdk.Plan {
	request := tfResources.SchemaRequest{}
	response := &tfResources.SchemaResponse{}
	resource.Schema(ctx, request, response)

	cluster, err := types.StringValue(clusterId).ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())
	pids, err := types.Int64Value(requiredPidsLimit).ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())
	configId, err := types.StringUnknown().ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())
	configName, err := types.StringUnknown().ToTerraformValue(ctx)
	Expect(err).NotTo(HaveOccurred())

	state := map[string]tftypes.Value{
		"cluster":        cluster,
		"pod_pids_limit": pids,
		"id":             configId,
		"name":           configName,
	}

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.Object{}, state),
		Schema: response.Schema,
	}
}

func createCluster(isHCP bool) (*cmv1.Cluster, error) {
	return cmv1.NewCluster().ID(clusterId).Hypershift(cmv1.NewHypershift().Enabled(isHCP)).Build()
}
