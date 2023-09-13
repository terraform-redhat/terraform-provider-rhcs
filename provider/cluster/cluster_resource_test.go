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

package cluster

***REMOVED***
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
***REMOVED***             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

var _ = Describe("Cluster creation", func(***REMOVED*** {
	clusterId := "1n2j3k4l5m6n7o8p9q0r"
	clusterName := "my-cluster"
	clusterVersion := "openshift-v4.11.12"
	productId := "rosa"
	cloudProviderId := "aws"
	regionId := "us-east-1"
	multiAz := true
	rosaCreatorArn := "arn:aws:iam::123456789012:dummy/dummy"
	apiUrl := "https://api.my-cluster.com:6443"
	consoleUrl := "https://console.my-cluster.com"
	machineType := "m5.xlarge"
	availabilityZone := "us-east-1a"
	ccsEnabled := true
	awsAccountID := "123456789012"
	awsAccessKeyID := "AKIAIOSFODNN7EXAMPLE"
	awsSecretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	privateLink := false

	It("Creates ClusterBuilder with correct field values", func(***REMOVED*** {
		clusterState := &ClusterState{
			Name: types.String{
				Value: clusterName,
	***REMOVED***,
			Version: types.String{
				Value: clusterVersion,
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
						Value: availabilityZone,
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
			Wait: types.Bool{
				Value: false,
	***REMOVED***,
***REMOVED***
		clusterObject, err := createClusterObject(context.Background(***REMOVED***, clusterState, diag.Diagnostics{}***REMOVED***
		Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

		Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***
		Expect(clusterObject***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***

		Expect(clusterObject.Name(***REMOVED******REMOVED***.To(Equal(clusterName***REMOVED******REMOVED***
		Expect(clusterObject.Version(***REMOVED***.ID(***REMOVED******REMOVED***.To(Equal(clusterVersion***REMOVED******REMOVED***

		id, ok := clusterObject.Region(***REMOVED***.GetID(***REMOVED***
		Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
		Expect(id***REMOVED***.To(Equal(regionId***REMOVED******REMOVED***

		Expect(clusterObject.AWS(***REMOVED***.AccountID(***REMOVED******REMOVED***.To(Equal(awsAccountID***REMOVED******REMOVED***

		availabilityZones := clusterObject.Nodes(***REMOVED***.AvailabilityZones(***REMOVED***
		Expect(availabilityZones***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		Expect(availabilityZones[0]***REMOVED***.To(Equal(availabilityZone***REMOVED******REMOVED***

		arn, ok := clusterObject.Properties(***REMOVED***["rosa_creator_arn"]
		Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
		Expect(arn***REMOVED***.To(Equal(arn***REMOVED******REMOVED***
	}***REMOVED***

	It("populateClusterState converts correctly a Cluster object into a ClusterState", func(***REMOVED*** {
		// We builder a Cluster object by creating a json and using cmv1.UnmarshalCluster on it
		clusterJson := map[string]interface{}{
			"id": clusterId,
			"product": map[string]interface{}{
				"id": productId,
	***REMOVED***,
			"cloud_provider": map[string]interface{}{
				"id": cloudProviderId,
	***REMOVED***,
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
					availabilityZone,
		***REMOVED***,
	***REMOVED***,
			"ccs": map[string]interface{}{
				"enabled": ccsEnabled,
	***REMOVED***,
			"aws": map[string]interface{}{
				"account_id":        awsAccountID,
				"access_key_id":     awsAccessKeyID,
				"secret_access_key": awsSecretAccessKey,
				"private_link":      privateLink,
	***REMOVED***,
			"version": map[string]interface{}{
				"id": clusterVersion,
	***REMOVED***,
***REMOVED***
		clusterJsonString, err := json.Marshal(clusterJson***REMOVED***
		Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

		clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString***REMOVED***
		Expect(err***REMOVED***.To(BeNil(***REMOVED******REMOVED***

		//We convert the Cluster object into a ClusterState and check that the conversion is correct
		clusterState := &ClusterState{}
		populateClusterState(clusterObject, clusterState***REMOVED***

		Expect(clusterState.ID.Value***REMOVED***.To(Equal(clusterId***REMOVED******REMOVED***
		Expect(clusterState.Version.Value***REMOVED***.To(Equal(clusterVersion***REMOVED******REMOVED***
		Expect(clusterState.Product.Value***REMOVED***.To(Equal(productId***REMOVED******REMOVED***
		Expect(clusterState.CloudProvider.Value***REMOVED***.To(Equal(cloudProviderId***REMOVED******REMOVED***
		Expect(clusterState.CloudRegion.Value***REMOVED***.To(Equal(regionId***REMOVED******REMOVED***
		Expect(clusterState.MultiAZ.Value***REMOVED***.To(Equal(multiAz***REMOVED******REMOVED***
		Expect(clusterState.Properties.Elems["rosa_creator_arn"].Equal(types.String{Value: rosaCreatorArn}***REMOVED******REMOVED***.To(Equal(true***REMOVED******REMOVED***
		Expect(clusterState.APIURL.Value***REMOVED***.To(Equal(apiUrl***REMOVED******REMOVED***
		Expect(clusterState.ConsoleURL.Value***REMOVED***.To(Equal(consoleUrl***REMOVED******REMOVED***
		Expect(clusterState.ComputeMachineType.Value***REMOVED***.To(Equal(machineType***REMOVED******REMOVED***
		Expect(clusterState.AvailabilityZones.Elems***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		Expect(clusterState.AvailabilityZones.Elems[0].Equal(types.String{Value: availabilityZone}***REMOVED******REMOVED***.To(Equal(true***REMOVED******REMOVED***
		Expect(clusterState.CCSEnabled.Value***REMOVED***.To(Equal(ccsEnabled***REMOVED******REMOVED***
		Expect(clusterState.AWSAccountID.Value***REMOVED***.To(Equal(awsAccountID***REMOVED******REMOVED***
		Expect(clusterState.AWSAccessKeyID.Value***REMOVED***.To(Equal(awsAccessKeyID***REMOVED******REMOVED***
		Expect(clusterState.AWSSecretAccessKey.Value***REMOVED***.To(Equal(awsSecretAccessKey***REMOVED******REMOVED***
		Expect(clusterState.AWSPrivateLink.Value***REMOVED***.To(Equal(privateLink***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
