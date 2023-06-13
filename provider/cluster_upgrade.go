/*
Copyright (c***REMOVED*** 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"context"
***REMOVED***
	"time"

	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/rosa/pkg/ocm"
***REMOVED***

// clusterUpgrade bundles the description of the upgrade with its current state
type clusterUpgrade struct {
	policy      *cmv1.UpgradePolicy
	policyState *cmv1.UpgradePolicyState
}

func (cu *clusterUpgrade***REMOVED*** State(***REMOVED*** cmv1.UpgradePolicyStateValue {
	return cu.policyState.Value(***REMOVED***
}

func (cu *clusterUpgrade***REMOVED*** Version(***REMOVED*** string {
	return cu.policy.Version(***REMOVED***
}

func (cu *clusterUpgrade***REMOVED*** NextRun(***REMOVED*** time.Time {
	return cu.policy.NextRun(***REMOVED***
}

func (cu *clusterUpgrade***REMOVED*** Delete(ctx context.Context, client *cmv1.ClustersClient***REMOVED*** error {
	_, err := client.Cluster(cu.policy.ClusterID(***REMOVED******REMOVED***.UpgradePolicies(***REMOVED***.UpgradePolicy(cu.policy.ID(***REMOVED******REMOVED***.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to delete upgrade policy: %v", err***REMOVED***
	}
	return nil
}

// Get the available upgrade versions that are reachable from a given starting
// version
func getAvailableUpgradeVersions(ctx context.Context, client *cmv1.VersionsClient, fromVersionId string***REMOVED*** ([]*cmv1.Version, error***REMOVED*** {
	// Retrieve info about the current version
	resp, err := client.Version(fromVersionId***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		return nil, fmt.Errorf("failed to get version information: %v", err***REMOVED***
	}
	version := resp.Body(***REMOVED***

	// Cycle through the available upgrades and find the ones that are ROSA enabled
	availableUpgradeVersions := []*cmv1.Version{}
	for _, v := range version.AvailableUpgrades(***REMOVED*** {
		id := ocm.CreateVersionID(v, version.ChannelGroup(***REMOVED******REMOVED***
		resp, err := client.Version(id***REMOVED***.
			Get(***REMOVED***.
			Send(***REMOVED***
		if err != nil {
			return nil, fmt.Errorf("failed to get version information: %v", err***REMOVED***
***REMOVED***
		availableVersion := resp.Body(***REMOVED***
		if availableVersion.ROSAEnabled(***REMOVED*** {
			availableUpgradeVersions = append(availableUpgradeVersions, availableVersion***REMOVED***
***REMOVED***
	}

	return availableUpgradeVersions, nil
}

// Get the list of upgrade policies associated with a cluster
func getScheduledUpgrades(ctx context.Context, client *cmv1.ClustersClient, clusterId string***REMOVED*** ([]clusterUpgrade, error***REMOVED*** {
	upgrades := []clusterUpgrade{}

	// Get the upgrade policies for the cluster
	upgradePolicies := []*cmv1.UpgradePolicy{}
	upgradeClient := client.Cluster(clusterId***REMOVED***.UpgradePolicies(***REMOVED***
	page := 1
	size := 100
	for {
		resp, err := upgradeClient.List(***REMOVED***.
			Page(page***REMOVED***.
			Size(size***REMOVED***.
			SendContext(ctx***REMOVED***
		if err != nil {
			return nil, fmt.Errorf("failed to list upgrade policies: %v", err***REMOVED***
***REMOVED***
		upgradePolicies = append(upgradePolicies, resp.Items(***REMOVED***.Slice(***REMOVED***...***REMOVED***
		if resp.Size(***REMOVED*** < size {
			break
***REMOVED***
		page++
	}

	// For each upgrade policy, get its state
	for _, policy := range upgradePolicies {
		// We only care about OSD upgrades (i.e., not CVE upgrades***REMOVED***
		if policy.UpgradeType(***REMOVED*** != "OSD" {
			continue
***REMOVED***
		resp, err := upgradeClient.UpgradePolicy(policy.ID(***REMOVED******REMOVED***.
			State(***REMOVED***.
			Get(***REMOVED***.
			SendContext(ctx***REMOVED***
		if err != nil {
			return nil, fmt.Errorf("failed to get upgrade policy state: %v", err***REMOVED***
***REMOVED***
		upgrades = append(upgrades, clusterUpgrade{
			policy:      policy,
			policyState: resp.Body(***REMOVED***,
***REMOVED******REMOVED***
	}

	return upgrades, nil
}

// Check the provided list of upgrades, canceling pending upgrades that are not
// for the correct version, and returning an error if there is already an
// upgrade in progress that is not for the desired version
func checkAndCancelUpgrades(ctx context.Context, client *cmv1.ClustersClient, upgrades []clusterUpgrade, desiredVersion *semver.Version***REMOVED*** (bool, error***REMOVED*** {
	correctUpgradePending := false
	tenMinFromNow := time.Now(***REMOVED***.UTC(***REMOVED***.Add(10 * time.Minute***REMOVED***

	for _, upgrade := range upgrades {
		tflog.Debug(ctx, "Found existing upgrade policy to %s in state %s", upgrade.Version(***REMOVED***, upgrade.State(***REMOVED******REMOVED***
		toVersion, err := semver.NewVersion(upgrade.Version(***REMOVED******REMOVED***
		if err != nil {
			return false, fmt.Errorf("failed to parse upgrade version: %v", err***REMOVED***
***REMOVED***
		switch upgrade.State(***REMOVED*** {
		case cmv1.UpgradePolicyStateValueDelayed, cmv1.UpgradePolicyStateValueStarted:
			if desiredVersion.Equal(toVersion***REMOVED*** {
				correctUpgradePending = true
	***REMOVED*** else {
				return false, fmt.Errorf("a cluster upgrade is already in progress"***REMOVED***
	***REMOVED***
		case cmv1.UpgradePolicyStateValuePending, cmv1.UpgradePolicyStateValueScheduled:
			if desiredVersion.Equal(toVersion***REMOVED*** && upgrade.NextRun(***REMOVED***.Before(tenMinFromNow***REMOVED*** {
				correctUpgradePending = true
	***REMOVED*** else {
				// The upgrade is not one we want, so cancel it
				if err := upgrade.Delete(ctx, client***REMOVED***; err != nil {
					return false, fmt.Errorf("failed to delete upgrade policy: %v", err***REMOVED***
		***REMOVED***
	***REMOVED***
***REMOVED***
	}
	return correctUpgradePending, nil
}
