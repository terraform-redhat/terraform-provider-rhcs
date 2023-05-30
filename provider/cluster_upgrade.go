/*
Copyright (c) 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package provider

import (
	"context"
	"fmt"
	"time"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/rosa/pkg/ocm"
)

// clusterUpgrade bundles the description of the upgrade with its current state
type clusterUpgrade struct {
	policy      *cmv1.UpgradePolicy
	policyState *cmv1.UpgradePolicyState
}

func (cu *clusterUpgrade) State() cmv1.UpgradePolicyStateValue {
	return cu.policyState.Value()
}

func (cu *clusterUpgrade) Version() string {
	return cu.policy.Version()
}

func (cu *clusterUpgrade) NextRun() time.Time {
	return cu.policy.NextRun()
}

func (cu *clusterUpgrade) Delete(ctx context.Context, client *cmv1.ClustersClient) error {
	_, err := client.Cluster(cu.policy.ClusterID()).UpgradePolicies().UpgradePolicy(cu.policy.ID()).Delete().SendContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete upgrade policy: %v", err)
	}
	return nil
}

// Get the available upgrade versions that are reachable from a given starting
// version
func getAvailableUpgrades(ctx context.Context, client *cmv1.VersionsClient, fromVersionId string) ([]*cmv1.Version, error) {
	// Retrieve info about the current version
	resp, err := client.Version(fromVersionId).Get().SendContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get version information: %v", err)
	}
	version := resp.Body()

	// Cycle through the available upgrades and find the ones that are ROSA enabled
	availableUpgrades := []*cmv1.Version{}
	for _, v := range version.AvailableUpgrades() {
		id := ocm.CreateVersionID(v, version.ChannelGroup())
		resp, err := client.Version(id).
			Get().
			Send()
		if err != nil {
			return nil, fmt.Errorf("failed to get version information: %v", err)
		}
		availableVersion := resp.Body()
		if availableVersion.ROSAEnabled() {
			availableUpgrades = append(availableUpgrades, availableVersion)
		}
	}

	return availableUpgrades, nil
}

// Get the list of upgrade policies associated with a cluster
func getScheduledUpgrades(ctx context.Context, client *cmv1.ClustersClient, clusterId string) ([]clusterUpgrade, error) {
	upgrades := []clusterUpgrade{}

	// Get the upgrade policies for the cluster
	upgradePolicies := []*cmv1.UpgradePolicy{}
	upgradeClient := client.Cluster(clusterId).UpgradePolicies()
	page := 1
	size := 100
	for {
		resp, err := upgradeClient.List().
			Page(page).
			Size(size).
			SendContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list upgrade policies: %v", err)
		}
		upgradePolicies = append(upgradePolicies, resp.Items().Slice()...)
		if resp.Size() < size {
			break
		}
		page++
	}

	// For each upgrade policy, get its state
	for _, policy := range upgradePolicies {
		// We only care about OSD upgrades (i.e., not CVE upgrades)
		if policy.UpgradeType() != "OSD" {
			continue
		}
		resp, err := upgradeClient.UpgradePolicy(policy.ID()).
			State().
			Get().
			SendContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get upgrade policy state: %v", err)
		}
		upgrades = append(upgrades, clusterUpgrade{
			policy:      policy,
			policyState: resp.Body(),
		})
	}

	return upgrades, nil
}
