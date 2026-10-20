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

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type hyperFleetNodePoolReplica struct {
	service exec.HyperfleetNodePoolService
	client  *hyperfleet.Clientset
	args    *exec.HyperfleetNodePoolArgs
	name    string
}

func newHyperFleetNodePoolReplica(
	clusterName, nodePoolName, clusterID, workspace string,
) (*hyperFleetNodePoolReplica, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID is required to create a HyperFleet nodepool")
	}
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
	childName := strings.TrimPrefix(name, clusterName+".")
	service, err := exec.NewHyperfleetNodePoolService(workspace + "-nodepool-" + childName)
	if err != nil {
		return nil, err
	}
	return &hyperFleetNodePoolReplica{
		service: service,
		client:  client,
		name:    name,
		args: &exec.HyperfleetNodePoolArgs{
			HyperfleetURL: &hyperfleetURL,
			AWSRegion:     &region,
			ClusterID:     &clusterID,
			ClusterName:   &clusterName,
			Name:          &name,
			SubnetID:      &subnetID,
			InstanceType:  &instanceType,
		},
	}, nil
}

func (n *hyperFleetNodePoolReplica) Create(spec exec.NodePoolSpec) error { return n.apply(spec) }

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

func (n *hyperFleetNodePoolReplica) Update(spec exec.NodePoolSpec) error { return n.apply(spec) }

func (n *hyperFleetNodePoolReplica) apply(spec exec.NodePoolSpec) error {
	if spec.Replicas != nil && spec.Autoscaling != nil {
		return fmt.Errorf("nodepool replicas and autoscaling are mutually exclusive")
	}
	if spec.Replicas != nil {
		n.args.Replicas = spec.Replicas
		n.args.Autoscaling = nil
	}
	if spec.Autoscaling != nil {
		n.args.Replicas = nil
		n.args.Autoscaling = spec.Autoscaling
	}
	if spec.EC2InstanceMetadataHttpTokens != nil {
		n.args.EC2InstanceMetadataHttpTokens = spec.EC2InstanceMetadataHttpTokens
	}
	_, err := n.service.Apply(n.args)
	return err
}

func (n *hyperFleetNodePoolReplica) Read(ctx context.Context) (exec.NodePoolState, error) {
	nodePool, err := n.client.HyperfleetV1alpha1().NodePools().Get(ctx, n.name, hfplatform.GetOptions{})
	if err != nil {
		return exec.NodePoolState{}, err
	}
	state := exec.NodePoolState{EC2InstanceMetadataHttpTokens: string(nodePool.Spec.EC2InstanceMetadataHttpTokens)}
	if nodePool.Spec.NodePool.Replicas != nil {
		state.Replicas = new(int(*nodePool.Spec.NodePool.Replicas))
	}
	if autoscaling := nodePool.Spec.NodePool.AutoScaling; autoscaling != nil && autoscaling.Min != nil {
		state.Autoscaling = &exec.NodePoolAutoscalingState{
			Enabled:     true,
			MinReplicas: int(*autoscaling.Min),
			MaxReplicas: int(autoscaling.Max),
		}
	}
	return state, nil
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

var _ exec.NodePoolLifecycle = (*hyperFleetNodePoolReplica)(nil)
