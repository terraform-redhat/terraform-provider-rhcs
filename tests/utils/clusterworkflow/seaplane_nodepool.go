// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type seaplaneNodePoolReplica struct {
	service exec.SeaplaneNodePoolService
	client  *seaplane.Clientset
	args    *exec.SeaplaneNodePoolArgs
	name    string
}

// NewSeaplaneNodePoolReplica creates the fixed-replica nodepool strategy for
// a cluster created by the Seaplane cluster strategy.
func NewSeaplaneNodePoolReplica(
	clusterName, nodePoolName, clusterID, workspace string,
) (exec.NodePoolReplicaLifecycle, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID is required to create a Seaplane nodepool")
	}
	return newSeaplaneNodePoolReplica(clusterName, nodePoolName, clusterID, workspace)
}

func newSeaplaneNodePoolReplica(
	clusterName, nodePoolName, clusterID, workspace string,
) (*seaplaneNodePoolReplica, error) {
	seaplaneURL, err := requiredEnv("SEAPLANE_URL")
	if err != nil {
		return nil, err
	}
	region := config.GetRegion()
	if region == "" {
		region = seaplaneAWSRegionRE.FindString(seaplaneURL)
	}
	if region == "" {
		return nil, fmt.Errorf(
			"AWS region is not set in REGION and cannot be derived from SEAPLANE_URL %q",
			seaplaneURL,
		)
	}
	if workspace == "" {
		workspace = clusterName + "-seaplane"
	}
	subnetIDs := config.GetSubnetIDList()
	subnetID := ""
	if len(subnetIDs) > 0 {
		subnetID = subnetIDs[0]
	} else {
		vpcService, vpcErr := exec.NewSeaplaneVPCService(workspace + "-vpc")
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
		return nil, fmt.Errorf("no subnet available for Seaplane nodepool")
	}
	client, err := buildSeaplaneClient(seaplaneURL, region)
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
	service, err := exec.NewSeaplaneNodePoolService(workspace + "-nodepool-" + childName)
	if err != nil {
		return nil, err
	}
	return &seaplaneNodePoolReplica{
		service: service,
		client:  client,
		name:    name,
		args: &exec.SeaplaneNodePoolArgs{
			SeaplaneURL:  &seaplaneURL,
			AWSRegion:    &region,
			ClusterID:    &clusterID,
			ClusterName:  &clusterName,
			Name:         &name,
			SubnetID:     &subnetID,
			InstanceType: &instanceType,
		},
	}, nil
}

func (n *seaplaneNodePoolReplica) Create(replicas int) error {
	n.args.Replicas = &replicas
	_, err := n.service.Apply(n.args)
	return err
}

func (n *seaplaneNodePoolReplica) WaitReady(ctx context.Context) error {
	return n.client.SeaplaneV1alpha1().NodePools().WaitUntil(
		ctx,
		n.name,
		func(nodePool *v1alpha1.NodePool) bool {
			if nodePool == nil {
				log.Logger.Infof("[seaplane] nodepool %s not found while waiting for Ready", n.name)
				return false
			}
			log.Logger.Infof("[seaplane] nodepool %s phase: %s", n.name, nodePool.Status.Phase)
			return nodePool.Status.Phase == v1alpha1.NodePoolPhaseReady
		},
		30*time.Second,
		45*time.Minute,
	)
}

func (n *seaplaneNodePoolReplica) UpdateReplicas(replicas int) error {
	n.args.Replicas = &replicas
	_, err := n.service.Apply(n.args)
	return err
}

func (n *seaplaneNodePoolReplica) Replicas(ctx context.Context) (int, error) {
	nodePool, err := n.client.SeaplaneV1alpha1().NodePools().Get(ctx, n.name, spplatform.GetOptions{})
	if err != nil {
		return 0, err
	}
	if nodePool.Spec.NodePool.Replicas == nil {
		return 0, fmt.Errorf("nodepool %q returned no replicas", n.name)
	}
	return int(*nodePool.Spec.NodePool.Replicas), nil
}

func (n *seaplaneNodePoolReplica) SetEC2InstanceMetadataHttpTokens(value string) {
	n.args.EC2InstanceMetadataHttpTokens = &value
}

func (n *seaplaneNodePoolReplica) EC2InstanceMetadataHttpTokens(ctx context.Context) (string, error) {
	nodePool, err := n.client.SeaplaneV1alpha1().NodePools().Get(ctx, n.name, spplatform.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(nodePool.Spec.EC2InstanceMetadataHttpTokens), nil
}

func (n *seaplaneNodePoolReplica) Destroy() error {
	_, err := n.service.Destroy()
	return err
}

func (n *seaplaneNodePoolReplica) WaitDeleted(ctx context.Context) error {
	return n.client.SeaplaneV1alpha1().NodePools().WaitUntil(
		ctx,
		n.name,
		func(nodePool *v1alpha1.NodePool) bool {
			if nodePool == nil {
				log.Logger.Infof("[seaplane] nodepool %s deleted", n.name)
				return true
			}
			log.Logger.Infof("[seaplane] nodepool %s phase while deleting: %s", n.name, nodePool.Status.Phase)
			return false
		},
		30*time.Second,
		45*time.Minute,
	)
}

var _ exec.NodePoolReplicaLifecycle = (*seaplaneNodePoolReplica)(nil)
