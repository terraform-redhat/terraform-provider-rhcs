package resource

import (
	"context"
	"errors"

	"github.com/openshift-online/ocm-common/pkg/cluster/validations"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type Cluster struct {
	clusterBuilder *cmv1.ClusterBuilder
}

func NewCluster() *Cluster {
	return &Cluster{
		clusterBuilder: cmv1.NewCluster(),
	}
}

func (c *Cluster) GetClusterBuilder() *cmv1.ClusterBuilder {
	return c.clusterBuilder
}

func (c *Cluster) Build() (object *cmv1.Cluster, err error) {
	return c.clusterBuilder.Build()
}

func (c *Cluster) CreateNodes(ctx context.Context, autoScalingEnabled bool, replicas *int64, minReplicas *int64,
	maxReplicas *int64, computeMachineType *string, labels map[string]string,
	availabilityZones []string, multiAZ bool) error {
	nodes := cmv1.NewClusterNodes()
	if computeMachineType != nil {
		nodes.ComputeMachineType(
			cmv1.NewMachineType().ID(*computeMachineType),
		)
	}

	if labels != nil {
		nodes.ComputeLabels(labels)
	}

	if availabilityZones != nil {
		if err := validations.ValidateAvailabilityZonesCount(multiAZ, len(availabilityZones)); err != nil {
			return err
		}
		nodes.AvailabilityZones(availabilityZones...)
	}

	if autoScalingEnabled {
		if replicas != nil {
			return errors.New("When autoscaling is enabled, replicas should not be configured")
		}

		autoscaling := cmv1.NewMachinePoolAutoscaling()
		minReplicasVal := 2
		if minReplicas != nil {
			minReplicasVal = int(*minReplicas)
		}
		if err := validations.MinReplicasValidator(minReplicasVal, multiAZ, false, 0); err != nil {
			return err
		}
		autoscaling.MinReplicas(minReplicasVal)
		maxReplicasVal := 2
		if maxReplicas != nil {
			maxReplicasVal = int(*maxReplicas)
		}
		if err := validations.MaxReplicasValidator(minReplicasVal, maxReplicasVal, multiAZ, false, 0); err != nil {
			return err
		}
		autoscaling.MaxReplicas(maxReplicasVal)
		if !autoscaling.Empty() {
			nodes.AutoscaleCompute(autoscaling)
		}
	} else {
		if minReplicas != nil || maxReplicas != nil {
			return errors.New("Autoscaling must be enabled in order to set min and max replicas")
		}

		replicasVal := 2
		if replicas != nil {
			replicasVal = int(*replicas)
		}
		if err := validations.MinReplicasValidator(replicasVal, multiAZ, false, 0); err != nil {
			return err
		}
		nodes.Compute(replicasVal)
	}

	if !nodes.Empty() {
		c.clusterBuilder.Nodes(nodes)
	}

	return nil
}
