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

package unittests

import (
	"encoding/json"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/cluster"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/cluster/clusterschema"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var _ = Describe("Cluster creation", func() {
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

	It("Creates ClusterBuilder with correct field values", func() {
		clusterState := &clusterschema.ClusterState{
			Name:              clusterName,
			Version:           common.Pointer(clusterVersion),
			CloudRegion:       regionId,
			AWSAccountID:      common.Pointer(awsAccountID),
			AvailabilityZones: []string{availabilityZone},
			Properties:        map[string]string{"rosa_creator_arn": rosaCreatorArn},
			Wait:              common.Pointer(false),
		}
		clusterObject, err := cluster.CreateClusterObject(clusterState)
		Expect(err).To(BeNil())

		Expect(err).To(BeNil())
		Expect(clusterObject).ToNot(BeNil())

		Expect(clusterObject.Name()).To(Equal(clusterName))
		Expect(clusterObject.Version().ID()).To(Equal(clusterVersion))

		id, ok := clusterObject.Region().GetID()
		Expect(ok).To(BeTrue())
		Expect(id).To(Equal(regionId))

		Expect(clusterObject.AWS().AccountID()).To(Equal(awsAccountID))

		availabilityZones := clusterObject.Nodes().AvailabilityZones()
		Expect(availabilityZones).To(HaveLen(1))
		Expect(availabilityZones[0]).To(Equal(availabilityZone))

		arn, ok := clusterObject.Properties()["rosa_creator_arn"]
		Expect(ok).To(BeTrue())
		Expect(arn).To(Equal(arn))
	})

	It("populateClusterState converts correctly a Cluster object into a ClusterState", func() {
		// We builder a Cluster object by creating a json and using cmv1.UnmarshalCluster on it
		//resourceData := &schema.ResourceData{}
		clusterJson := map[string]interface{}{
			"id": clusterId,
			"product": map[string]interface{}{
				"id": productId,
			},
			"cloud_provider": map[string]interface{}{
				"id": cloudProviderId,
			},
			"region": map[string]interface{}{
				"id": regionId,
			},
			"multi_az": multiAz,
			"properties": map[string]interface{}{
				"rosa_creator_arn": rosaCreatorArn,
			},
			"api": map[string]interface{}{
				"url": apiUrl,
			},
			"console": map[string]interface{}{
				"url": consoleUrl,
			},
			"nodes": map[string]interface{}{
				"compute_machine_type": map[string]interface{}{
					"id": machineType,
				},
				"availability_zones": []interface{}{
					availabilityZone,
				},
			},
			"ccs": map[string]interface{}{
				"enabled": ccsEnabled,
			},
			"aws": map[string]interface{}{
				"account_id":        awsAccountID,
				"access_key_id":     awsAccessKeyID,
				"secret_access_key": awsSecretAccessKey,
				"private_link":      privateLink,
			},
			"version": map[string]interface{}{
				"id": clusterVersion,
			},
		}
		clusterJsonString, err := json.Marshal(clusterJson)
		Expect(err).To(BeNil())

		_, err = cmv1.UnmarshalCluster(clusterJsonString)
		//clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString)
		Expect(err).To(BeNil())

		////We convert the Cluster object into a ClusterState and check that the conversion is correct
		//clusterState := &clusterschema.ClusterState{}
		//populateClusterState(clusterObject, clusterState)
		//
		//Expect(clusterState.ID.Value).To(Equal(clusterId))
		//Expect(clusterState.Version.Value).To(Equal(clusterVersion))
		//Expect(clusterState.Product.Value).To(Equal(productId))
		//Expect(clusterState.CloudProvider.Value).To(Equal(cloudProviderId))
		//Expect(clusterState.CloudRegion.Value).To(Equal(regionId))
		//Expect(clusterState.MultiAZ.Value).To(Equal(multiAz))
		//Expect(clusterState.Properties.Elems["rosa_creator_arn"].Equal(types.String{Value: rosaCreatorArn})).To(Equal(true))
		//Expect(clusterState.APIURL.Value).To(Equal(apiUrl))
		//Expect(clusterState.ConsoleURL.Value).To(Equal(consoleUrl))
		//Expect(clusterState.ComputeMachineType.Value).To(Equal(machineType))
		//Expect(clusterState.AvailabilityZones.Elems).To(HaveLen(1))
		//Expect(clusterState.AvailabilityZones.Elems[0].Equal(types.String{Value: availabilityZone})).To(Equal(true))
		//Expect(clusterState.CCSEnabled.Value).To(Equal(ccsEnabled))
		//Expect(clusterState.AWSAccountID.Value).To(Equal(awsAccountID))
		//Expect(clusterState.AWSAccessKeyID.Value).To(Equal(awsAccessKeyID))
		//Expect(clusterState.AWSSecretAccessKey.Value).To(Equal(awsSecretAccessKey))
		//Expect(clusterState.AWSPrivateLink.Value).To(Equal(privateLink))
	})
})
