/*
Copyright (c) 2024 Red Hat, Inc.

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
package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ocmUtils "github.com/openshift-online/ocm-common/pkg/ocm/utils"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/zgalor/weberr"
)

type MachinePoolUpgrade struct {
	Policy      *cmv1.NodePoolUpgradePolicy
	PolicyState *cmv1.UpgradePolicyState
}

// Get the available upgrade versions that are reachable from a given starting
// version
func GetAvailableUpgradeVersions(
	ctx context.Context,
	clustersClient *cmv1.ClustersClient,
	versionClient *cmv1.VersionsClient,
	clusterId string, nodePoolId string) ([]*cmv1.Version, error) {
	clusterResp, err := clustersClient.Cluster(clusterId).Get().SendContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster information: %v", err)
	}
	// Retrieve info about the current version
	resp, err := clustersClient.Cluster(clusterId).
		NodePools().NodePool(nodePoolId).
		Get().SendContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get version information: %v", err)
	}
	nodePool := resp.Body()
	cluster := clusterResp.Body()
	version := nodePool.Version()

	// Cycle through the available upgrades and find the ones that are HCP enabled
	availableUpgradeVersions := []*cmv1.Version{}
	for _, v := range version.AvailableUpgrades() {
		id := ocmUtils.CreateVersionId(v, cluster.Version().ChannelGroup())
		resp, err := versionClient.Version(id).
			Get().
			Send()
		if err != nil {
			return nil, fmt.Errorf("failed to get version information: %v", err)
		}
		availableVersion := resp.Body()
		if availableVersion.HostedControlPlaneEnabled() {
			availableUpgradeVersions = append(availableUpgradeVersions, availableVersion)
		}
	}

	return availableUpgradeVersions, nil
}

// Get the list of upgrade policies associated with a cluster
func GetScheduledUpgrades(ctx context.Context,
	client *cmv1.ClustersClient, clusterId string, machinePoolId string) ([]MachinePoolUpgrade, error) {
	upgrades := []MachinePoolUpgrade{}

	// Get the upgrade policies for the cluster
	upgradePolicies := []*cmv1.NodePoolUpgradePolicy{}
	upgradeClient := client.Cluster(clusterId).NodePools().NodePool(machinePoolId).UpgradePolicies()
	page := 1
	size := 100
	for {
		resp, err := upgradeClient.List().
			Page(page).
			Size(size).
			SendContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list control plane upgrade policies: %v", err)
		}
		upgradePolicies = append(upgradePolicies, resp.Items().Slice()...)
		if resp.Size() < size {
			break
		}
		page++
	}

	// For each upgrade policy, get its state
	for _, policy := range upgradePolicies {
		if policy.UpgradeType() != cmv1.UpgradeTypeNodePool {
			continue
		}
		resp, err := upgradeClient.NodePoolUpgradePolicy(policy.ID()).
			Get().
			SendContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get upgrade policy state: %v", err)
		}
		state := resp.Body().State()
		upgrades = append(upgrades, MachinePoolUpgrade{
			Policy:      policy,
			PolicyState: state,
		})
	}

	return upgrades, nil
}

// Check the provided list of upgrades, canceling pending upgrades that are not
// for the correct version, and returning an error if there is already an
// upgrade in progress that is not for the desired version
func CheckAndCancelUpgrades(
	ctx context.Context,
	client *cmv1.ClustersClient,
	upgrades []MachinePoolUpgrade, desiredVersion *semver.Version) (bool, error) {
	correctUpgradePending := false
	tenMinFromNow := time.Now().UTC().Add(10 * time.Minute)

	for _, upgrade := range upgrades {
		tflog.Debug(ctx, fmt.Sprintf("Found existing upgrade policy to %s in state %s", upgrade.Policy.Version(), upgrade.PolicyState.Value()))
		toVersion, err := semver.NewVersion(upgrade.Policy.Version())
		if err != nil {
			return false, fmt.Errorf("failed to parse upgrade version: %v", err)
		}
		switch upgrade.PolicyState.Value() {
		case cmv1.UpgradePolicyStateValueDelayed, cmv1.UpgradePolicyStateValueStarted:
			if desiredVersion.Equal(toVersion) {
				correctUpgradePending = true
			} else {
				return false, fmt.Errorf("a cluster upgrade is already in progress")
			}
		case cmv1.UpgradePolicyStateValuePending, cmv1.UpgradePolicyStateValueScheduled:
			if desiredVersion.Equal(toVersion) && upgrade.Policy.NextRun().Before(tenMinFromNow) {
				correctUpgradePending = true
			} else {
				// The upgrade is not one we want, so cancel it
				_, err := client.Cluster(upgrade.Policy.ClusterID()).
					NodePools().NodePool(upgrade.Policy.NodePoolID()).UpgradePolicies().
					NodePoolUpgradePolicy(upgrade.Policy.ID()).
					Delete().SendContext(ctx)
				if err != nil {
					return false, fmt.Errorf("failed to delete upgrade policy: %v", err)
				}
			}
		}
	}
	return correctUpgradePending, nil
}

func AckVersionGate(
	gateAgreementsClient *cmv1.VersionGateAgreementsClient,
	gateID string) error {
	agreement, err := cmv1.NewVersionGateAgreement().
		VersionGate(cmv1.NewVersionGate().ID(gateID)).
		Build()
	if err != nil {
		return err
	}
	response, err := gateAgreementsClient.Add().Body(agreement).Send()
	if err != nil {
		return common.HandleErr(response.Error(), err)
	}
	return nil
}

// Construct a list of missing gate agreements for upgrade to a given cluster version
// Returns: a list of all un-acked gate agreements, a string describing the ones that need user ack, and an error
func CheckMissingAgreements(version string,
	clusterKey string, upgradePoliciesClient *cmv1.NodePoolUpgradePoliciesClient) ([]*cmv1.VersionGate, string, error) {
	// Schedule an upgrade
	tenMinFromNow := time.Now().UTC().Add(10 * time.Minute)
	upgradePolicyBuilder := cmv1.NewNodePoolUpgradePolicy().
		ScheduleType(cmv1.ScheduleTypeManual).
		Version(version).
		NextRun(tenMinFromNow)
	upgradePolicy, err := upgradePolicyBuilder.Build()
	if err != nil {
		return []*cmv1.VersionGate{}, "", fmt.Errorf("failed to build upgrade policy: %v", err)
	}

	// check if the cluster upgrade requires gate agreements
	gates, err := getMissingGateAgreements(upgradePolicy, upgradePoliciesClient)
	if err != nil {
		return []*cmv1.VersionGate{}, "", fmt.Errorf("failed to check for missing gate agreements upgrade for "+
			"cluster '%s': %v", clusterKey, err)
	}
	str := "\nMissing required acknowledgements to schedule upgrade." +
		"\nRead the below description and acknowledge to proceed with upgrade." +
		"\nDescription:"
	counter := 1
	for _, gate := range gates {
		if !gate.STSOnly() { // STS-only gates don't require user acknowledgement
			str = fmt.Sprintf("%s\n%d) %s\n", str, counter, gate.Description())

			if gate.WarningMessage() != "" {
				str = fmt.Sprintf("%s   Warning:     %s\n", str, gate.WarningMessage())
			}
			str = fmt.Sprintf("%s   URL:         %s\n", str, gate.DocumentationURL())
			counter++
		}
	}
	return gates, str, nil
}

func getMissingGateAgreements(
	upgradePolicy *cmv1.NodePoolUpgradePolicy,
	upgradePoliciesClient *cmv1.NodePoolUpgradePoliciesClient) ([]*cmv1.VersionGate, error) {
	response, err := upgradePoliciesClient.Add().Parameter("dryRun", true).Body(upgradePolicy).Send()

	if err != nil {
		if response.Error() != nil {
			// parse gates list
			errorDetails, ok := response.Error().GetDetails()
			if !ok {
				return []*cmv1.VersionGate{}, common.HandleErr(response.Error(), err)
			}
			data, err := json.Marshal(errorDetails)
			if err != nil {
				return []*cmv1.VersionGate{}, common.HandleErr(response.Error(), err)
			}
			gates, err := cmv1.UnmarshalVersionGateList(data)
			if err != nil {
				return []*cmv1.VersionGate{}, common.HandleErr(response.Error(), err)
			}
			// return original error if invaild version gate detected
			if len(gates) > 0 && gates[0].ID() == "" {
				errType := weberr.ErrorType(response.Error().Status())
				return []*cmv1.VersionGate{}, errType.Set(weberr.Errorf(response.Error().Reason()))
			}
			return gates, nil
		}
	}
	return []*cmv1.VersionGate{}, nil
}
