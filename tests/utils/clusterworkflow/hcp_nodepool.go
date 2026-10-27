// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"fmt"
	"maps"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
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
) (*hcpNodePoolReplica, error) {
	// Keep each nodepool's Terraform state separate so shared workflows can
	// create multiple pools before waiting for them to become ready.
	serviceProfile := profile.DuplicateRandom()
	service, err := serviceProfile.Services().GetMachinePoolsService()
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

func (n *hcpNodePoolReplica) Create(spec exec.NodePoolSpec) error { return n.apply(spec) }

func (n *hcpNodePoolReplica) WaitReady(ctx context.Context) error {
	return wait.PollUntilContextTimeout(
		ctx,
		30*time.Second,
		20*time.Minute,
		false,
		func(context.Context) (bool, error) {
			_, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, n.clusterID, n.name)
			if err != nil {
				log.Logger.Infof("[hcp] nodepool %s not ready yet: %v", n.name, err)
				return false, nil
			}
			log.Logger.Infof("[hcp] nodepool %s is available", n.name)
			return true, nil
		},
	)
}

func (n *hcpNodePoolReplica) Update(spec exec.NodePoolSpec) error { return n.apply(spec) }

func (n *hcpNodePoolReplica) apply(spec exec.NodePoolSpec) error {
	if spec.Replicas != nil && spec.Autoscaling != nil {
		return fmt.Errorf("nodepool replicas and autoscaling are mutually exclusive")
	}
	if spec.Replicas != nil {
		n.args.Replicas = spec.Replicas
		n.args.AutoscalingEnabled = new(false)
		n.args.MinReplicas = nil
		n.args.MaxReplicas = nil
	}
	if spec.Autoscaling != nil {
		n.args.Replicas = nil
		n.args.AutoscalingEnabled = spec.Autoscaling.Enabled
		n.args.MinReplicas = spec.Autoscaling.MinReplicas
		n.args.MaxReplicas = spec.Autoscaling.MaxReplicas
	}
	if spec.EC2InstanceMetadataHttpTokens != nil {
		n.args.Ec2MetadataHttpTokens = spec.EC2InstanceMetadataHttpTokens
	}
	if spec.Tags != nil {
		n.args.Tags = spec.Tags
	}
	if spec.Labels != nil {
		labels := maps.Clone(*spec.Labels)
		n.args.Labels = &labels
	}
	if spec.Taints != nil {
		taints := make([]map[string]string, len(*spec.Taints))
		for i, taint := range *spec.Taints {
			taints[i] = map[string]string{
				"key":           taint.Key,
				"value":         taint.Value,
				"schedule_type": taint.Effect,
			}
		}
		n.args.Taints = &taints
	}
	if spec.AutoRepair != nil {
		n.args.AutoRepair = spec.AutoRepair
	}
	if spec.DiskSize != nil {
		n.args.DiskSize = spec.DiskSize
	}
	_, err := n.service.Apply(n.args)
	return err
}

func (n *hcpNodePoolReplica) Read(_ context.Context) (exec.NodePoolState, error) {
	nodePool, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, n.clusterID, n.name)
	if err != nil {
		return exec.NodePoolState{}, err
	}
	replicas := nodePool.Replicas()
	state := exec.NodePoolState{
		Replicas:   &replicas,
		Labels:     nodePool.Labels(),
		AutoRepair: new(nodePool.AutoRepair()),
	}
	if autoscaling := nodePool.Autoscaling(); autoscaling != nil {
		state.Autoscaling = &exec.NodePoolAutoscalingState{
			Enabled:     true,
			MinReplicas: autoscaling.MinReplica(),
			MaxReplicas: autoscaling.MaxReplica(),
		}
	}
	if nodePool.AWSNodePool() != nil {
		if volume := nodePool.AWSNodePool().RootVolume(); volume != nil {
			diskSize := volume.Size()
			state.DiskSize = &diskSize
		}
		state.EC2InstanceMetadataHttpTokens = string(nodePool.AWSNodePool().Ec2MetadataHttpTokens())
		state.Tags = nodePool.AWSNodePool().Tags()
	}
	for _, taint := range nodePool.Taints() {
		state.Taints = append(state.Taints, exec.NodePoolTaint{
			Key:    taint.Key(),
			Value:  taint.Value(),
			Effect: taint.Effect(),
		})
	}
	return state, nil
}

var _ exec.NodePoolLifecycle = (*hcpNodePoolReplica)(nil)

func (n *hcpNodePoolReplica) WaitTagsApplied(ctx context.Context, expected map[string]string) error {
	return wait.PollUntilContextTimeout(
		ctx,
		15*time.Second,
		10*time.Minute,
		true,
		func(ctx context.Context) (bool, error) {
			state, err := n.Read(ctx)
			if err != nil {
				log.Logger.Infof("[hcp] nodepool %s tags are not available yet: %v", n.name, err)
				return false, nil
			}
			if !maps.Equal(state.Tags, expected) {
				log.Logger.Infof("[hcp] nodepool %s tags have not converged yet", n.name)
				return false, nil
			}
			return true, nil
		},
	)
}

func (n *hcpNodePoolReplica) Destroy() error {
	_, err := n.service.Destroy()
	return err
}

func (n *hcpNodePoolReplica) WaitDeleted(ctx context.Context) error {
	return wait.PollUntilContextTimeout(
		ctx,
		30*time.Second,
		45*time.Minute,
		false,
		func(context.Context) (bool, error) {
			_, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, n.clusterID, n.name)
			if err != nil {
				log.Logger.Infof("[hcp] nodepool %s deleted", n.name)
				return true, nil
			}
			log.Logger.Infof("[hcp] nodepool %s still exists while deleting", n.name)
			return false, nil
		},
	)
}
