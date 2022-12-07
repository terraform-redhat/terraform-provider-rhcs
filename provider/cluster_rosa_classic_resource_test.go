/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

***REMOVED***
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
***REMOVED***             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

type MockHttpClient struct {
	response *http.Response
}

func (c MockHttpClient***REMOVED*** Get(url string***REMOVED*** (resp *http.Response, err error***REMOVED*** {
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
***REMOVED***

var (
	mockHttpClient = MockHttpClient{
		response: &http.Response{
			TLS: &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					&x509.Certificate{
						Raw: []byte("nonce"***REMOVED***,
			***REMOVED***,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
	}
***REMOVED***

func generateBasicRosaClassicClusterJson(***REMOVED*** map[string]interface{} {
	return map[string]interface{}{
		"id": clusterId,
		"region": map[string]interface{}{
			"id": regionId,
***REMOVED***,
		"multi_az": multiAz,
		"properties": map[string]interface{}{
			"rosa_creator_arn": rosaCreatorArn,
***REMOVED***,
		"api": map[string]interface{}{
			"url": apiUrl,
***REMOVED***,
		"console": map[string]interface{}{
			"url": consoleUrl,
***REMOVED***,
		"nodes": map[string]interface{}{
			"compute_machine_type": map[string]interface{}{
				"id": machineType,
	***REMOVED***,
			"availability_zones": []interface{}{
				availabilityZone1,
	***REMOVED***,
***REMOVED***,
		"ccs": map[string]interface{}{
			"enabled": ccsEnabled,
***REMOVED***,
		"aws": map[string]interface{}{
			"account_id":   awsAccountID,
			"private_link": privateLink,
			"sts": map[string]interface{}{
				"oidc_endpoint_url": oidcEndpointUrl,
				"role_arn":          roleArn,
	***REMOVED***,
***REMOVED***,
	}
}

func generateBasicRosaClassicClusterState(***REMOVED*** *ClusterRosaClassicState {
	return &ClusterRosaClassicState{
		Name: types.String{
			Value: clusterName,
***REMOVED***,
		CloudRegion: types.String{
			Value: regionId,
***REMOVED***,
		AWSAccountID: types.String{
			Value: awsAccountID,
***REMOVED***,
		AvailabilityZones: types.List{
			Elems: []attr.Value{
				types.String{
					Value: availabilityZone1,
		***REMOVED***,
				types.String{
					Value: availabilityZone2,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
		Properties: types.Map{
			Elems: map[string]attr.Value{
				"rosa_creator_arn": types.String{
					Value: rosaCreatorArn,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
		Version: types.String{
			Value: "4.10",
***REMOVED***,
		Proxy: &Proxy{
			HttpProxy: types.String{
				Value: httpProxy,
	***REMOVED***,
			HttpsProxy: types.String{
				Value: httpsProxy,
	***REMOVED***,
***REMOVED***,
		Sts: &Sts{

***REMOVED***,
	}
}

var _ = Describe("Rosa Classic Sts cluster", func(***REMOVED*** {
	Context("createClassicClusterObject", func(***REMOVED*** {
		It("Creates a cluster with correct field values", func(***REMOVED*** {
			clusterState := generateBasicRosaClassicClusterState(***REMOVED***
			rosaClusterObject, err := createClassicClusterObject(context.Background(***REMOVED***, clusterState, &logging.StdLogger{}, diag.Diagnostics{}***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

			Expect(rosaClusterObject.Name(***REMOVED******REMOVED***.To(Equal(clusterName***REMOVED******REMOVED***

			id, ok := rosaClusterObject.Region(***REMOVED***.GetID(***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
			Expect(id***REMOVED***.To(Equal(regionId***REMOVED******REMOVED***

			Expect(rosaClusterObject.AWS(***REMOVED***.AccountID(***REMOVED******REMOVED***.To(Equal(awsAccountID***REMOVED******REMOVED***

			availabilityZones := rosaClusterObject.Nodes(***REMOVED***.AvailabilityZones(***REMOVED***
			Expect(availabilityZones***REMOVED***.To(HaveLen(2***REMOVED******REMOVED***
			Expect(availabilityZones[0]***REMOVED***.To(Equal(availabilityZone1***REMOVED******REMOVED***
			Expect(availabilityZones[1]***REMOVED***.To(Equal(availabilityZone2***REMOVED******REMOVED***

			Expect(rosaClusterObject.Proxy(***REMOVED***.HTTPProxy(***REMOVED******REMOVED***.To(Equal(httpProxy***REMOVED******REMOVED***
			Expect(rosaClusterObject.Proxy(***REMOVED***.HTTPSProxy(***REMOVED******REMOVED***.To(Equal(httpsProxy***REMOVED******REMOVED***

			arn, ok := rosaClusterObject.Properties(***REMOVED***["rosa_creator_arn"]
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
			Expect(arn***REMOVED***.To(Equal(rosaCreatorArn***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	It("Throws an error when version format is invalid", func(***REMOVED*** {
		clusterState := generateBasicRosaClassicClusterState(***REMOVED***
		clusterState.Version.Value = "a.4.1"
		_, err := createClassicClusterObject(context.Background(***REMOVED***, clusterState, &logging.StdLogger{}, diag.Diagnostics{}***REMOVED***
		Expect(err***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	}***REMOVED***

	It("Throws an error when version is unsupported", func(***REMOVED*** {
		clusterState := generateBasicRosaClassicClusterState(***REMOVED***
		clusterState.Version.Value = "4.1.0"
		_, err := createClassicClusterObject(context.Background(***REMOVED***, clusterState, &logging.StdLogger{}, diag.Diagnostics{}***REMOVED***
		Expect(err***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	}***REMOVED***

	Context("populateRosaClassicClusterState", func(***REMOVED*** {
		It("Converts correctly a Cluster object into a ClusterRosaClassicState", func(***REMOVED*** {
			clusterState := &ClusterRosaClassicState{}
			clusterJson := generateBasicRosaClassicClusterJson(***REMOVED***
			clusterJsonString, err := json.Marshal(clusterJson***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

			clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

			populateRosaClassicClusterState(context.Background(***REMOVED***, clusterObject, clusterState, &logging.StdLogger{}, mockHttpClient***REMOVED***

			Expect(clusterState.ID.Value***REMOVED***.To(Equal(clusterId***REMOVED******REMOVED***
			Expect(clusterState.CloudRegion.Value***REMOVED***.To(Equal(regionId***REMOVED******REMOVED***
			Expect(clusterState.MultiAZ.Value***REMOVED***.To(Equal(multiAz***REMOVED******REMOVED***
			Expect(clusterState.Properties.Elems["rosa_creator_arn"].Equal(types.String{Value: rosaCreatorArn}***REMOVED******REMOVED***.To(Equal(true***REMOVED******REMOVED***
			Expect(clusterState.APIURL.Value***REMOVED***.To(Equal(apiUrl***REMOVED******REMOVED***
			Expect(clusterState.ConsoleURL.Value***REMOVED***.To(Equal(consoleUrl***REMOVED******REMOVED***
			Expect(clusterState.ComputeMachineType.Value***REMOVED***.To(Equal(machineType***REMOVED******REMOVED***
			Expect(clusterState.AvailabilityZones.Elems***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
			Expect(clusterState.AvailabilityZones.Elems[0].Equal(types.String{Value: availabilityZone1}***REMOVED******REMOVED***.To(Equal(true***REMOVED******REMOVED***
			Expect(clusterState.CCSEnabled.Value***REMOVED***.To(Equal(ccsEnabled***REMOVED******REMOVED***
			Expect(clusterState.AWSAccountID.Value***REMOVED***.To(Equal(awsAccountID***REMOVED******REMOVED***
			Expect(clusterState.AWSPrivateLink.Value***REMOVED***.To(Equal(privateLink***REMOVED******REMOVED***
			Expect(clusterState.Sts.OIDCEndpointURL.Value***REMOVED***.To(Equal(oidcEndpointUrl***REMOVED******REMOVED***
			Expect(clusterState.Sts.RoleARN.Value***REMOVED***.To(Equal(roleArn***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Check trimming of oidc url with https perfix", func(***REMOVED*** {
			clusterState := &ClusterRosaClassicState{}
			clusterJson := generateBasicRosaClassicClusterJson(***REMOVED***
			clusterJson["aws"].(map[string]interface{}***REMOVED***["sts"].(map[string]interface{}***REMOVED***["oidc_endpoint_url"] = "https://nonce.com"
			clusterJson["aws"].(map[string]interface{}***REMOVED***["sts"].(map[string]interface{}***REMOVED***["operator_role_prefix"] = "terraform-operator"

			clusterJsonString, err := json.Marshal(clusterJson***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***
			print(string(clusterJsonString***REMOVED******REMOVED***

			clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

			err = populateRosaClassicClusterState(context.Background(***REMOVED***, clusterObject, clusterState, &logging.StdLogger{}, mockHttpClient***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(clusterState.Sts.OIDCEndpointURL.Value***REMOVED***.To(Equal("nonce.com"***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Throws an error when oidc_endpoint_url is an invalid url", func(***REMOVED*** {
			clusterState := &ClusterRosaClassicState{}
			clusterJson := generateBasicRosaClassicClusterJson(***REMOVED***
			clusterJson["aws"].(map[string]interface{}***REMOVED***["sts"].(map[string]interface{}***REMOVED***["oidc_endpoint_url"] = "invalid$url"
			clusterJsonString, err := json.Marshal(clusterJson***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***
			print(string(clusterJsonString***REMOVED******REMOVED***

			clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

			err = populateRosaClassicClusterState(context.Background(***REMOVED***, clusterObject, clusterState, &logging.StdLogger{}, mockHttpClient***REMOVED***
			Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***
			Expect(clusterState.Sts.Thumbprint.Value***REMOVED***.To(Equal(""***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
