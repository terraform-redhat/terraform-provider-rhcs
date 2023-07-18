package resource

***REMOVED***
	"context"
	"errors"

	"github.com/openshift-online/ocm-common/pkg/cluster/validations"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type Cluster struct {
	clusterBuilder *cmv1.ClusterBuilder
}

func NewCluster(***REMOVED*** *Cluster {
	return &Cluster{
		clusterBuilder: cmv1.NewCluster(***REMOVED***,
	}
}

func (c *Cluster***REMOVED*** GetClusterBuilder(***REMOVED*** *cmv1.ClusterBuilder {
	return c.clusterBuilder
}

func (c *Cluster***REMOVED*** Build(***REMOVED*** (object *cmv1.Cluster, err error***REMOVED*** {
	return c.clusterBuilder.Build(***REMOVED***
}

func (c *Cluster***REMOVED*** CreateNodes(ctx context.Context, autoScalingEnabled bool, replicas *int64, minReplicas *int64,
	maxReplicas *int64, computeMachineType *string, labels map[string]string,
	availabilityZones []string, multiAZ bool***REMOVED*** error {
	nodes := cmv1.NewClusterNodes(***REMOVED***
	if computeMachineType != nil {
		nodes.ComputeMachineType(
			cmv1.NewMachineType(***REMOVED***.ID(*computeMachineType***REMOVED***,
		***REMOVED***
	}

	if labels != nil {
		nodes.ComputeLabels(labels***REMOVED***
	}

	if availabilityZones != nil {
		if err := validations.ValidateAvailabilityZonesCount(multiAZ, len(availabilityZones***REMOVED******REMOVED***; err != nil {
			return err
***REMOVED***
		nodes.AvailabilityZones(availabilityZones...***REMOVED***
	}

	if autoScalingEnabled {
		if replicas != nil {
			return errors.New("When autoscaling is enabled, replicas should not be configured"***REMOVED***
***REMOVED***

		autoscaling := cmv1.NewMachinePoolAutoscaling(***REMOVED***
		minReplicasVal := 2
		if minReplicas != nil {
			minReplicasVal = int(*minReplicas***REMOVED***
***REMOVED***
		if err := validations.MinReplicasValidator(minReplicasVal, multiAZ, false, 0***REMOVED***; err != nil {
			return err
***REMOVED***
		autoscaling.MinReplicas(minReplicasVal***REMOVED***
		maxReplicasVal := 2
		if maxReplicas != nil {
			maxReplicasVal = int(*maxReplicas***REMOVED***
***REMOVED***
		if err := validations.MaxReplicasValidator(minReplicasVal, maxReplicasVal, multiAZ, false, 0***REMOVED***; err != nil {
			return err
***REMOVED***
		autoscaling.MaxReplicas(maxReplicasVal***REMOVED***
		if !autoscaling.Empty(***REMOVED*** {
			nodes.AutoscaleCompute(autoscaling***REMOVED***
***REMOVED***
	} else {
		if minReplicas != nil || maxReplicas != nil {
			return errors.New("Autoscaling must be enabled in order to set min and max replicas"***REMOVED***
***REMOVED***

		replicasVal := 2
		if replicas != nil {
			replicasVal = int(*replicas***REMOVED***
***REMOVED***
		if err := validations.MinReplicasValidator(replicasVal, multiAZ, false, 0***REMOVED***; err != nil {
			return err
***REMOVED***
		nodes.Compute(replicasVal***REMOVED***
	}

	if !nodes.Empty(***REMOVED*** {
		c.clusterBuilder.Nodes(nodes***REMOVED***
	}

	return nil
}
