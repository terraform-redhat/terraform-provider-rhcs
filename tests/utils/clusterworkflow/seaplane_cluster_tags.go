// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"fmt"
	"maps"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

// SeaplaneClusterTagsUpdater updates root AWS tags on an existing
// profile-managed cluster without constructing or replacing the cluster.
type SeaplaneClusterTagsUpdater struct {
	client      *seaplane.Clientset
	service     exec.SeaplaneClusterService
	args        *exec.SeaplaneClusterArgs
	clusterName string
}

// NewSeaplaneClusterTagsUpdater loads the existing cluster's Terraform state
// and API client. It does not require a release image or create a cluster.
func NewSeaplaneClusterTagsUpdater(
	ctx context.Context,
	profile profilehandler.ProfileHandler,
	clusterID string,
) (*SeaplaneClusterTagsUpdater, error) {
	if profile.Profile().GetClusterType() != constants.SEAPLANE {
		return nil, fmt.Errorf("cluster tag updates require a Seaplane profile")
	}
	if clusterID == "" {
		return nil, fmt.Errorf("an existing Seaplane cluster ID is required")
	}
	fetcher, err := newSeaplaneFetcher(profile)
	if err != nil {
		return nil, err
	}
	fetchedID, err := fetcher.FetchClusterID(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch existing Seaplane cluster ID: %w", err)
	}
	if fetchedID != clusterID {
		return nil, fmt.Errorf("profile cluster ID %q does not match requested cluster ID %q", fetchedID, clusterID)
	}
	args, err := fetcher.service.ReadTFVars()
	if err != nil {
		return nil, fmt.Errorf("read existing Seaplane cluster Terraform variables: %w", err)
	}
	return &SeaplaneClusterTagsUpdater{
		client:      fetcher.client,
		service:     fetcher.service,
		args:        args,
		clusterName: fetcher.clusterName,
	}, nil
}

func (u *SeaplaneClusterTagsUpdater) ClusterID(ctx context.Context) (string, error) {
	cluster, err := u.client.SeaplaneV1alpha1().Clusters().Get(ctx, u.clusterName, spplatform.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(cluster.UID), nil
}

func (u *SeaplaneClusterTagsUpdater) Tags(ctx context.Context) (map[string]string, error) {
	cluster, err := u.client.SeaplaneV1alpha1().Clusters().Get(ctx, u.clusterName, spplatform.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cluster.Spec.Tags, nil
}

func (u *SeaplaneClusterTagsUpdater) UpdateTags(tags map[string]string) error {
	args := *u.args
	args.Tags = &tags
	if _, err := u.service.Apply(&args); err != nil {
		return fmt.Errorf("apply Seaplane cluster tags: %w", err)
	}
	u.args = &args
	return nil
}

func (u *SeaplaneClusterTagsUpdater) WaitTagsApplied(ctx context.Context, expected map[string]string) error {
	err := u.client.SeaplaneV1alpha1().Clusters().WaitUntil(
		ctx,
		u.clusterName,
		func(cluster *v1alpha1.Cluster) bool {
			if cluster == nil {
				log.Logger.Infof("[seaplane] cluster %s not found while waiting for tags", u.clusterName)
				return false
			}

			synced := false
			for _, condition := range cluster.Status.Conditions {
				if condition.Type == seaplaneConditionSynced {
					synced = condition.Status == metav1.ConditionTrue &&
						condition.ObservedGeneration == cluster.Generation
					break
				}
			}
			tagsMatch := maps.Equal(cluster.Spec.Tags, expected)
			statusCurrent := cluster.Status.ObservedGeneration == cluster.Generation
			log.Logger.Infof(
				"[seaplane] cluster %s tags applied: synced=%t statusCurrent=%t tagsMatch=%t",
				u.clusterName,
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
		return fmt.Errorf("wait for Seaplane Cluster %q tags to be applied: %w", u.clusterName, err)
	}
	return nil
}
