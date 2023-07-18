package resource

***REMOVED***
	"context"

***REMOVED***
***REMOVED***
***REMOVED***

var _ = Describe("Cluster - create nodes validation", func(***REMOVED*** {
	var cluster *Cluster
	ctx := context.TODO(***REMOVED***
	BeforeEach(func(***REMOVED*** {
		cluster = NewCluster(***REMOVED***
	}***REMOVED***
	It("Autoscaling disabled minReplicas set - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Autoscaling must be enabled in order to set min and max replicas"***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled maxReplicas set - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, nil, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Autoscaling must be enabled in order to set min and max replicas"***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled replicas smaller than 2 - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, pointer(int64(1***REMOVED******REMOVED***, nil, nil, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Cluster requires at least 2 compute nodes"***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled default replicas - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, nil, nil, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled 3 replicas - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, pointer(int64(3***REMOVED******REMOVED***, nil, nil, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(3***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling enabled replicas set - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, true, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("When autoscaling is enabled, replicas should not be configured"***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling enabled default minReplicas & maxReplicas - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, true, nil, nil, nil, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(0***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(autoscaleCompute.MinReplicas(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
		Expect(autoscaleCompute.MaxReplicas(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling enabled default maxReplicas smaller than minReplicas - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, true, nil, pointer(int64(4***REMOVED******REMOVED***, pointer(int64(3***REMOVED******REMOVED***, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("max-replicas must be greater or equal to min-replicas"***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling enabled set minReplicas & maxReplicas - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, true, nil, pointer(int64(2***REMOVED******REMOVED***, pointer(int64(4***REMOVED******REMOVED***, nil, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(0***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(autoscaleCompute.MinReplicas(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
		Expect(autoscaleCompute.MaxReplicas(***REMOVED******REMOVED***.To(Equal(4***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled set ComputeMachineType - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, nil, nil, pointer("asdf"***REMOVED***, nil, nil, false***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		machineType := ocmClusterNode.ComputeMachineType(***REMOVED***
		Expect(machineType***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(machineType.ID(***REMOVED******REMOVED***.To(Equal("asdf"***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled set compute labels - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, nil, nil, nil, map[string]string{"key1": "val1"}, nil, false***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
		computeLabels := ocmClusterNode.ComputeLabels(***REMOVED***
		Expect(computeLabels***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		Expect(computeLabels["key1"]***REMOVED***.To(Equal("val1"***REMOVED******REMOVED***
		Expect(ocmClusterNode.AvailabilityZones(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled multiAZ false set one availability zone - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, nil, nil, nil, nil, []string{"us-east-1a"}, false***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		azs := ocmClusterNode.AvailabilityZones(***REMOVED***
		Expect(azs***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(2***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled multiAZ false set three availability zones - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, false***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("The number of availability zones for a single AZ cluster should be 1, instead received: 3"***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled multiAZ true set three availability zones and two replicas - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, pointer(int64(2***REMOVED******REMOVED***, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Multi AZ cluster requires at least 3 compute nodes"***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled multiAZ true set three availability zones and three replicas - success", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, pointer(int64(3***REMOVED******REMOVED***, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmCluster, err := cluster.Build(***REMOVED***
		Expect(err***REMOVED***.NotTo(HaveOccurred(***REMOVED******REMOVED***
		ocmClusterNode := ocmCluster.Nodes(***REMOVED***
		Expect(ocmClusterNode***REMOVED***.NotTo(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeMachineType(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***
		Expect(ocmClusterNode.ComputeLabels(***REMOVED******REMOVED***.To(BeEmpty(***REMOVED******REMOVED***
		azs := ocmClusterNode.AvailabilityZones(***REMOVED***
		Expect(azs***REMOVED***.To(HaveLen(3***REMOVED******REMOVED***
		Expect(ocmClusterNode.Compute(***REMOVED******REMOVED***.To(Equal(3***REMOVED******REMOVED***
		autoscaleCompute := ocmClusterNode.AutoscaleCompute(***REMOVED***
		Expect(autoscaleCompute***REMOVED***.To(BeNil(***REMOVED******REMOVED***
	}***REMOVED***
	It("Autoscaling disabled multiAZ true set one zone - failure", func(***REMOVED*** {
		err := cluster.CreateNodes(ctx, false, nil, nil, nil, nil, nil, []string{"us-east-1a", "us-east-1b", "us-east-1c"}, true***REMOVED***
		Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
		Expect(err.Error(***REMOVED******REMOVED***.To(Equal("Multi AZ cluster requires at least 3 compute nodes"***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

func pointer[T any](src T***REMOVED*** *T {
	return &src
}
