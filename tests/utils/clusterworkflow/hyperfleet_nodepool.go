// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfplatform "github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

type hyperFleetNodePoolReplica struct {
	service exec.HyperfleetNodePoolService
	client  *hyperfleet.Clientset
	args    *exec.HyperfleetNodePoolArgs
	name    string
}

// NewHyperFleetNodePoolReplica creates the fixed-replica nodepool strategy for
// a cluster created by the HyperFleet cluster strategy.
func NewHyperFleetNodePoolReplica(
	profile profilehandler.ProfileHandler, clusterName, nodePoolName, clusterID, workspace string,
) (exec.NodePoolReplicaLifecycle, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID is required to create a HyperFleet nodepool")
	}
	return newHyperFleetNodePoolReplica(profile, clusterName, nodePoolName, clusterID, workspace)
}

func newHyperFleetNodePoolReplica(
	profile profilehandler.ProfileHandler, clusterName, nodePoolName, clusterID, workspace string,
) (*hyperFleetNodePoolReplica, error) {
	hyperfleetURL, err := requiredEnv("HYPERFLEET_URL")
	if err != nil {
		return nil, err
	}
	region := config.GetRegion()
	if region == "" {
		region = hyperFleetAWSRegionRE.FindString(hyperfleetURL)
	}
	if region == "" {
		return nil, fmt.Errorf(
			"AWS region is not set in REGION and cannot be derived from HYPERFLEET_URL %q",
			hyperfleetURL,
		)
	}
	if workspace == "" {
		workspace = clusterName + "-hyperfleet"
	}
	subnetIDs := config.GetSubnetIDList()
	subnetID := ""
	if len(subnetIDs) > 0 {
		subnetID = subnetIDs[0]
	} else {
		vpcService, vpcErr := exec.NewHyperfleetVPCService(workspace + "-vpc")
		if vpcErr != nil {
			return nil, vpcErr
		}
		output, outputErr := vpcService.Output()
		if outputErr != nil {
			return nil, outputErr
		}
		subnetID = output.PrivateSubnetID
	}
	if subnetID == "" {
		return nil, fmt.Errorf("no subnet available for HyperFleet nodepool")
	}
	service, err := exec.NewHyperfleetNodePoolService(workspace + "-nodepool-replicas")
	if err != nil {
		return nil, err
	}
	client, err := buildHyperFleetClient(hyperfleetURL, region)
	if err != nil {
		return nil, err
	}
	instanceType := config.GetComputeMachineType()
	if instanceType == "" {
		instanceType = "m5.xlarge"
	}
	name := nodePoolName
	if !strings.HasPrefix(name, clusterName+".") {
		name = clusterName + "." + name
	}
	args := &exec.HyperfleetNodePoolArgs{
		HyperfleetURL: &hyperfleetURL,
		AWSRegion:     &region,
		ClusterID:     &clusterID,
		ClusterName:   &clusterName,
		Name:          &name,
		SubnetID:      &subnetID,
		InstanceType:  &instanceType,
	}
	// The image family is baked into every instance the pool launches, so it
	// can only be set on create; changing it would replace the pool.
	if profile != nil && profile.Profile().IsWindows() {
		imageType := string(hypershiftv1beta1.ImageTypeWindows)
		args.ImageType = &imageType
	}
	return &hyperFleetNodePoolReplica{
		service: service,
		client:  client,
		name:    name,
		args:    args,
	}, nil
}

func (n *hyperFleetNodePoolReplica) Create(replicas int) error {
	n.args.Replicas = &replicas
	_, err := n.service.Apply(n.args)
	return err
}

func (n *hyperFleetNodePoolReplica) WaitReady(ctx context.Context) error {
	return n.client.HyperfleetV1alpha1().NodePools().WaitUntil(
		ctx,
		n.name,
		func(nodePool *v1alpha1.NodePool) bool {
			if nodePool == nil {
				log.Logger.Infof("[hyperfleet] nodepool %s not found while waiting for Ready", n.name)
				return false
			}
			log.Logger.Infof("[hyperfleet] nodepool %s phase: %s", n.name, nodePool.Status.Phase)
			return nodePool.Status.Phase == v1alpha1.NodePoolPhaseReady
		},
		30*time.Second,
		45*time.Minute,
	)
}

func (n *hyperFleetNodePoolReplica) UpdateReplicas(replicas int) error {
	n.args.Replicas = &replicas
	_, err := n.service.Apply(n.args)
	return err
}

func (n *hyperFleetNodePoolReplica) Replicas(ctx context.Context) (int, error) {
	nodePool, err := n.client.HyperfleetV1alpha1().NodePools().Get(ctx, n.name, hfplatform.GetOptions{})
	if err != nil {
		return 0, err
	}
	if nodePool.Spec.NodePool.Replicas == nil {
		return 0, fmt.Errorf("nodepool %q returned no replicas", n.name)
	}
	return int(*nodePool.Spec.NodePool.Replicas), nil
}

func (n *hyperFleetNodePoolReplica) Destroy() error {
	_, err := n.service.Destroy()
	return err
}

func (n *hyperFleetNodePoolReplica) WaitDeleted(ctx context.Context) error {
	return n.client.HyperfleetV1alpha1().NodePools().WaitUntil(
		ctx,
		n.name,
		func(nodePool *v1alpha1.NodePool) bool {
			if nodePool == nil {
				log.Logger.Infof("[hyperfleet] nodepool %s deleted", n.name)
				return true
			}
			log.Logger.Infof("[hyperfleet] nodepool %s phase while deleting: %s", n.name, nodePool.Status.Phase)
			return false
		},
		30*time.Second,
		45*time.Minute,
	)
}

var _ exec.NodePoolReplicaLifecycle = (*hyperFleetNodePoolReplica)(nil)
