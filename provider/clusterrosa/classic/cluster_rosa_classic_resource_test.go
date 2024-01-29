/*
Copyright (c) 2021 Red Hat, Inc.

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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/sts"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"
)

type MockHttpClient struct {
	response *http.Response
}

func (c MockHttpClient) Get(url string) (resp *http.Response, err error) {
	return c.response, nil
}

const (
	clusterId         = "1n2j3k4l5m6n7o8p9q0r"
	clusterName       = "my-cluster"
	regionId          = "us-east-1"
	multiAz           = true
	rosaCreatorArn    = "arn:aws:iam::123456789012:dummy/dummy"
	apiUrl            = "https://api.my-cluster.com:6443"
	consoleUrl        = "https://console.my-cluster.com"
	baseDomain        = "alias.p1.openshiftapps.com"
	machineType       = "m5.xlarge"
	availabilityZone1 = "us-east-1a"
	availabilityZone2 = "us-east-1b"
	ccsEnabled        = true
	awsAccountID      = "123456789012"
	privateLink       = false
	oidcEndpointUrl   = "example.com"
	roleArn           = "arn:aws:iam::123456789012:role/role-name"
	httpProxy         = "http://proxy.com"
	httpsProxy        = "https://proxy.com"
	httpTokens        = "required"
)

var (
	mockHttpClient = MockHttpClient{
		response: &http.Response{
			TLS: &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{
						Raw: []byte("nonce"),
					},
				},
			},
		},
	}
)

func generateBasicRosaClassicClusterJson() map[string]interface{} {
	return map[string]interface{}{
		"id":   clusterId,
		"name": clusterName,
		"region": map[string]interface{}{
			"id": regionId,
		},
		"multi_az": multiAz,
		"properties": map[string]interface{}{
			"rosa_creator_arn": rosaCreatorArn,
			"rosa_tf_version":  build.Version,
			"rosa_tf_commit":   build.Commit,
		},
		"api": map[string]interface{}{
			"url": apiUrl,
		},
		"console": map[string]interface{}{
			"url": consoleUrl,
		},
		"dns": map[string]interface{}{
			"base_domain": baseDomain,
		},
		"nodes": map[string]interface{}{
			"compute_machine_type": map[string]interface{}{
				"id": machineType,
			},
			"availability_zones": []interface{}{
				availabilityZone1,
			},
		},
		"ccs": map[string]interface{}{
			"enabled": ccsEnabled,
		},
		"aws": map[string]interface{}{
			"account_id":               awsAccountID,
			"private_link":             privateLink,
			"ec2_metadata_http_tokens": httpTokens,
			"sts": map[string]interface{}{
				"oidc_endpoint_url": oidcEndpointUrl,
				"role_arn":          roleArn,
			},
		},
	}
}

func generateBasicRosaClassicClusterState() *ClusterRosaClassicState {
	azs, err := common.StringArrayToList([]string{availabilityZone1})
	if err != nil {
		return nil
	}
	properties, err := common.ConvertStringMapToMapType(map[string]string{"rosa_creator_arn": rosaCreatorArn})
	if err != nil {
		return nil
	}
	return &ClusterRosaClassicState{
		Name:              types.StringValue(clusterName),
		CloudRegion:       types.StringValue(regionId),
		AWSAccountID:      types.StringValue(awsAccountID),
		AvailabilityZones: azs,
		Properties:        properties,
		ChannelGroup:      types.StringValue("stable"),
		Version:           types.StringValue("4.10"),
		Proxy: &proxy.Proxy{
			HttpProxy:  types.StringValue(httpProxy),
			HttpsProxy: types.StringValue(httpsProxy),
		},
		Sts:         &sts.ClassicSts{},
		Replicas:    types.Int64Value(2),
		MinReplicas: types.Int64Null(),
		MaxReplicas: types.Int64Null(),
		KMSKeyArn:   types.StringNull(),
	}
}

func TestResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Rosa Resource Suite")
}

var _ = Describe("Rosa Classic Sts cluster", func() {
	Context("createClassicClusterObject", func() {
		It("Creates a cluster with correct field values", func() {
			clusterState := generateBasicRosaClassicClusterState()
			rosaClusterObject, err := createClassicClusterObject(context.Background(), clusterState, diag.Diagnostics{})
			Expect(err).To(BeNil())

			Expect(rosaClusterObject.Name()).To(Equal(clusterName))

			id, ok := rosaClusterObject.Region().GetID()
			Expect(ok).To(BeTrue())
			Expect(id).To(Equal(regionId))

			Expect(rosaClusterObject.AWS().AccountID()).To(Equal(awsAccountID))

			availabilityZones := rosaClusterObject.Nodes().AvailabilityZones()
			Expect(availabilityZones).To(HaveLen(1))
			Expect(availabilityZones[0]).To(Equal(availabilityZone1))

			Expect(rosaClusterObject.Proxy().HTTPProxy()).To(Equal(httpProxy))
			Expect(rosaClusterObject.Proxy().HTTPSProxy()).To(Equal(httpsProxy))

			arn, ok := rosaClusterObject.Properties()["rosa_creator_arn"]
			Expect(ok).To(BeTrue())
			Expect(arn).To(Equal(rosaCreatorArn))

			version, ok := rosaClusterObject.Version().GetID()
			Expect(ok).To(BeTrue())
			Expect(version).To(Equal("openshift-v4.10"))
			channel, ok := rosaClusterObject.Version().GetChannelGroup()
			Expect(ok).To(BeTrue())
			Expect(channel).To(Equal("stable"))
		})
	})
	It("Throws an error when version format is invalid", func() {
		clusterState := generateBasicRosaClassicClusterState()
		clusterState.Version = types.StringValue("a.4.1")
		_, err := createClassicClusterObject(context.Background(), clusterState, diag.Diagnostics{})
		Expect(err).ToNot(BeNil())
	})

	It("Throws an error when version is unsupported", func() {
		clusterState := generateBasicRosaClassicClusterState()
		clusterState.Version = types.StringValue("4.1.0")
		_, err := createClassicClusterObject(context.Background(), clusterState, diag.Diagnostics{})
		Expect(err).ToNot(BeNil())
	})

	It("appends the non-default channel name to the requested version", func() {
		clusterState := generateBasicRosaClassicClusterState()
		clusterState.ChannelGroup = types.StringValue("somechannel")
		rosaClusterObject, err := createClassicClusterObject(context.Background(), clusterState, diag.Diagnostics{})
		Expect(err).To(BeNil())

		version, ok := rosaClusterObject.Version().GetID()
		Expect(ok).To(BeTrue())
		Expect(version).To(Equal("openshift-v4.10-somechannel"))
		channel, ok := rosaClusterObject.Version().GetChannelGroup()
		Expect(ok).To(BeTrue())
		Expect(channel).To(Equal("somechannel"))
	})

	Context("populateRosaClassicClusterState", func() {
		It("Converts correctly a Cluster object into a ClusterRosaClassicState", func() {
			clusterState := &ClusterRosaClassicState{}
			clusterJson := generateBasicRosaClassicClusterJson()
			clusterJsonString, err := json.Marshal(clusterJson)
			Expect(err).To(BeNil())

			clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString)
			Expect(err).To(BeNil())
			Expect(populateRosaClassicClusterState(context.Background(), clusterObject, clusterState, mockHttpClient)).To(Succeed())

			Expect(clusterState.ID.ValueString()).To(Equal(clusterId))
			Expect(clusterState.CloudRegion.ValueString()).To(Equal(regionId))
			Expect(clusterState.MultiAZ.ValueBool()).To(Equal(multiAz))

			properties, err := common.OptionalMap(context.Background(), clusterState.Properties)
			Expect(err).To(BeNil())
			Expect(properties["rosa_creator_arn"]).To(Equal(rosaCreatorArn))

			ocmProperties, err := common.OptionalMap(context.Background(), clusterState.OCMProperties)
			Expect(err).To(BeNil())
			Expect(ocmProperties["rosa_tf_version"]).To(Equal(build.Version))
			Expect(ocmProperties["rosa_tf_commit"]).To(Equal(build.Commit))

			Expect(clusterState.APIURL.ValueString()).To(Equal(apiUrl))
			Expect(clusterState.ConsoleURL.ValueString()).To(Equal(consoleUrl))
			Expect(clusterState.Domain.ValueString()).To(Equal(fmt.Sprintf("%s.%s", clusterName, baseDomain)))

			Expect(clusterState.AvailabilityZones.Elements()).To(HaveLen(1))
			azs, err := common.StringListToArray(context.Background(), clusterState.AvailabilityZones)
			Expect(err).To(BeNil())
			Expect(azs[0]).To(Equal(availabilityZone1))

			Expect(clusterState.CCSEnabled.ValueBool()).To(Equal(ccsEnabled))
			Expect(clusterState.AWSAccountID.ValueString()).To(Equal(awsAccountID))
			Expect(clusterState.AWSPrivateLink.ValueBool()).To(Equal(privateLink))
			Expect(clusterState.Sts.OIDCEndpointURL.ValueString()).To(Equal(oidcEndpointUrl))
			Expect(clusterState.Sts.RoleARN.ValueString()).To(Equal(roleArn))
			Expect(clusterState.Ec2MetadataHttpTokens.ValueString()).To(Equal(httpTokens))
		})
		It("Check trimming of oidc url with https perfix", func() {
			clusterState := &ClusterRosaClassicState{}
			clusterJson := generateBasicRosaClassicClusterJson()
			clusterJson["aws"].(map[string]interface{})["sts"].(map[string]interface{})["oidc_endpoint_url"] = "https://nonce.com"
			clusterJson["aws"].(map[string]interface{})["sts"].(map[string]interface{})["operator_role_prefix"] = "terraform-operator"

			clusterJsonString, err := json.Marshal(clusterJson)
			Expect(err).To(BeNil())
			print(string(clusterJsonString))

			clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString)
			Expect(err).To(BeNil())

			err = populateRosaClassicClusterState(context.Background(), clusterObject, clusterState, mockHttpClient)
			Expect(err).To(BeNil())
			Expect(clusterState.Sts.OIDCEndpointURL.ValueString()).To(Equal("nonce.com"))
		})

		It("Throws an error when oidc_endpoint_url is an invalid url", func() {
			clusterState := &ClusterRosaClassicState{}
			clusterJson := generateBasicRosaClassicClusterJson()
			clusterJson["aws"].(map[string]interface{})["sts"].(map[string]interface{})["oidc_endpoint_url"] = "invalid$url"
			clusterJsonString, err := json.Marshal(clusterJson)
			Expect(err).To(BeNil())
			print(string(clusterJsonString))

			clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString)
			Expect(err).To(BeNil())

			err = populateRosaClassicClusterState(context.Background(), clusterObject, clusterState, mockHttpClient)
			Expect(err).To(BeNil())
			Expect(clusterState.Sts.Thumbprint.ValueString()).To(Equal(""))
		})
	})

	Context("http tokens state validation", func() {
		It("Fail validation with lower version than allowed", func() {
			clusterState := generateBasicRosaClassicClusterState()
			clusterState.Ec2MetadataHttpTokens = types.StringValue(string(cmv1.Ec2MetadataHttpTokensRequired))
			err := validateHttpTokensVersion(context.Background(), clusterState, "openshift-v4.10.0")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("is not supported with ec2_metadata_http_tokens"))
		})
		It("Pass validation with http_tokens_state and supported version", func() {
			clusterState := generateBasicRosaClassicClusterState()
			err := validateHttpTokensVersion(context.Background(), clusterState, "openshift-v4.11.0")
			Expect(err).To(BeNil())
		})
	})

})
