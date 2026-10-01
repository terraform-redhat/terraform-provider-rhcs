// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

type hcpNodePoolReplica struct {
	service   exec.MachinePoolService
	clusterID string
	name      string
	args      *exec.MachinePoolArgs
}

func newHCPNodePoolReplica(
	profile profilehandler.ProfileHandler, clusterID, nodePoolName string,
) (exec.NodePoolReplicaLifecycle, error) {
	service, err := profile.Services().GetMachinePoolsService()
	if err != nil {
		return nil, err
	}
	vpcService, err := profile.Services().GetVPCService()
	if err != nil {
		return nil, err
	}
	vpcOutput, err := vpcService.Output()
	if err != nil {
		return nil, err
	}
	if len(vpcOutput.PrivateSubnets) == 0 {
		return nil, fmt.Errorf("HCP VPC has no private subnets")
	}
	name := nodePoolName
	instanceType := profile.Profile().GetComputeMachineType()
	if instanceType == "" {
		instanceType = "m5.xlarge"
	}
	return &hcpNodePoolReplica{
		service:   service,
		clusterID: clusterID,
		name:      name,
		args: &exec.MachinePoolArgs{
			Cluster:            &clusterID,
			Name:               &name,
			SubnetID:           &vpcOutput.PrivateSubnets[0],
			MachineType:        &instanceType,
			AutoscalingEnabled: new(false),
			AutoRepair:         new(true),
		},
	}, nil
}

func (n *hcpNodePoolReplica) Create(replicas int) error {
	n.args.Replicas = &replicas
	_, err := n.service.Apply(n.args)
	return err
}

func (n *hcpNodePoolReplica) WaitReady(ctx context.Context) error {
	return wait.PollUntilContextTimeout(ctx, 30*time.Second, 20*time.Minute, false, func(context.Context) (bool, error) {
		_, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, n.clusterID, n.name)
		if err != nil {
			Logger.Infof("[hcp] nodepool %s not ready yet: %v", n.name, err)
			return false, nil
		}
		Logger.Infof("[hcp] nodepool %s is available", n.name)
		return true, nil
	})
}

func (n *hcpNodePoolReplica) UpdateReplicas(replicas int) error {
	n.args.Replicas = &replicas
	_, err := n.service.Apply(n.args)
	return err
}

func (n *hcpNodePoolReplica) Replicas(_ context.Context) (int, error) {
	nodePool, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, n.clusterID, n.name)
	if err != nil {
		return 0, err
	}
	return nodePool.Replicas(), nil
}

func (n *hcpNodePoolReplica) Destroy() error {
	_, err := n.service.Destroy()
	return err
}

func (n *hcpNodePoolReplica) WaitDeleted(ctx context.Context) error {
	return wait.PollUntilContextTimeout(ctx, 30*time.Second, 45*time.Minute, false, func(context.Context) (bool, error) {
		_, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, n.clusterID, n.name)
		if err != nil {
			Logger.Infof("[hcp] nodepool %s deleted", n.name)
			return true, nil
		}
		Logger.Infof("[hcp] nodepool %s still exists while deleting", n.name)
		return false, nil
	})
}
