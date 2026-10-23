// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

type seaplaneNodePoolReplica struct {
	service exec.SeaplaneNodePoolService
	client  *seaplane.Clientset
	args    *exec.SeaplaneNodePoolArgs
	name    string
}

// NewSeaplaneNodePoolReplica creates the fixed-replica nodepool strategy for
// a cluster created by the Seaplane cluster strategy. The optional profile
// parameter supplies platform-specific defaults (like Windows image type).
func NewSeaplaneNodePoolReplica(
	profile profilehandler.ProfileHandler, clusterName, nodePoolName, clusterID, workspace string,
) (*seaplaneNodePoolReplica, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID is required to create a Seaplane nodepool")
	}
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

	args := &exec.SeaplaneNodePoolArgs{
		SeaplaneURL:  &seaplaneURL,
		AWSRegion:    &region,
		ClusterID:    &clusterID,
		ClusterName:  &clusterName,
		Name:         &name,
		SubnetID:     &subnetID,
		InstanceType: &instanceType,
	}

	// The image family is baked into every instance the pool launches, so it
	// can only be set on create; changing it would replace the pool.
	if profile != nil && profile.Profile().IsWindows() {
		imageType := string(hypershiftv1beta1.ImageTypeWindows)
		args.ImageType = &imageType
	}

	return &seaplaneNodePoolReplica{
		service: service,
		client:  client,
		name:    name,
		args:    args,
	}, nil
}

func (n *seaplaneNodePoolReplica) Create(spec exec.NodePoolSpec) error { return n.apply(spec) }

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

func (n *seaplaneNodePoolReplica) Update(spec exec.NodePoolSpec) error { return n.apply(spec) }

func (n *seaplaneNodePoolReplica) apply(spec exec.NodePoolSpec) error {
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
	if spec.Tags != nil {
		n.args.Tags = *spec.Tags
	}
	_, err := n.service.Apply(n.args)
	return err
}

func (n *seaplaneNodePoolReplica) Read(ctx context.Context) (exec.NodePoolState, error) {
	nodePool, err := n.client.SeaplaneV1alpha1().NodePools().Get(ctx, n.name, spplatform.GetOptions{})
	if err != nil {
		return exec.NodePoolState{}, err
	}
	state := exec.NodePoolState{
		EC2InstanceMetadataHttpTokens: string(nodePool.Spec.EC2InstanceMetadataHttpTokens),
		Tags:                          nodePool.Spec.Tags,
	}
	if nodePool.Spec.NodePool.Replicas != nil {
		replicas := int(*nodePool.Spec.NodePool.Replicas)
		state.Replicas = &replicas
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

func (n *seaplaneNodePoolReplica) WaitTagsApplied(ctx context.Context, expected map[string]string) error {
	err := n.client.SeaplaneV1alpha1().NodePools().WaitUntil(
		ctx,
		n.name,
		func(nodePool *v1alpha1.NodePool) bool {
			if nodePool == nil {
				log.Logger.Infof("[seaplane] nodepool %s not found while waiting for tags", n.name)
				return false
			}

			synced := false
			for _, condition := range nodePool.Status.Conditions {
				if condition.Type == "Synced" {
					synced = condition.Status == metav1.ConditionTrue &&
						condition.ObservedGeneration == nodePool.Generation
					break
				}
			}
			tagsMatch := maps.Equal(nodePool.Spec.Tags, expected)
			statusCurrent := nodePool.Status.ObservedGeneration == nodePool.Generation
			log.Logger.Infof(
				"[seaplane] nodepool %s tags applied: synced=%t statusCurrent=%t tagsMatch=%t",
				n.name,
				synced,
				statusCurrent,
				tagsMatch,
			)
			return synced && statusCurrent && tagsMatch
		},
		15*time.Second,
		10*time.Minute,
	)
	if err != nil {
		return fmt.Errorf("wait for Seaplane NodePool %q tags to be applied: %w", n.name, err)
	}
	return nil
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

var _ exec.NodePoolLifecycle = (*seaplaneNodePoolReplica)(nil)
