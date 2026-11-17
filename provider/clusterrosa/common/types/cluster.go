// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ocmUtils "github.com/openshift-online/ocm-common/pkg/ocm/utils"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type CloudProvider string

const (
	Aws CloudProvider = "aws"
)

type Product string

const (
	Rosa Product = "rosa"
)

type ClusterTopology string

const (
	Classic ClusterTopology = "classic"
	Hcp     ClusterTopology = "hcp"
)

const (
	PoolMessage = "This attribute specifically applies to the Worker Machine Pool and becomes irrelevant once the resource is created. Any modifications to the initial Machine Pool should be made through the Terraform imported Machine Pool resource. For more details, refer to [Worker Machine Pool in ROSA Cluster](../guides/worker-machine-pool.md)"
)

type BaseCluster struct {
	ClusterCollection *cmv1.ClustersClient
	VersionCollection *cmv1.VersionsClient
	ClusterWait       common.ClusterWait
}

// getAndValidateVersionInChannelGroup ensures that the cluster version is
// available in the channel group or optionally filtered by a specific channel stream.
// Pass the channel as the last optional parameter to filter by a specific channel (e.g., "stable-4.10").
func (b *BaseCluster) GetAndValidateVersionInChannelGroup(ctx context.Context,
	topology ClusterTopology,
	stateChannelGroup string, stateVersion string, optionalChannel ...string) (string, error) {
	versionList, listErr := b.getVersionList(ctx, topology, stateChannelGroup, optionalChannel...)

	if stateVersion != "" {
		if listErr == nil && slices.Contains(versionList, stateVersion) {
			return stateVersion, nil
		}
		validated, fallbackErr := b.validateExplicitVersion(
			ctx, topology, stateChannelGroup, stateVersion, optionalChannel...)
		if fallbackErr == nil {
			tflog.Debug(ctx, fmt.Sprintf(
				"Validated explicit version %s via OCM version GET (not present in paginated list)", stateVersion))
			return validated, nil
		}
		if listErr != nil {
			return "", fallbackErr
		}
		return "", fmt.Errorf(
			"version %s is not in the list of supported versions: %v: %w",
			stateVersion, versionList, fallbackErr)
	}

	if listErr != nil {
		return "", listErr
	}

	return versionList[0], nil
}

// validateExplicitVersion loads a single version by ID when paginated list results are
// incomplete (for example OCP 5.x missing from list search while GET by ID succeeds).
func (b *BaseCluster) validateExplicitVersion(ctx context.Context,
	topology ClusterTopology, channelGroup, stateVersion string, optionalChannel ...string) (string, error) {
	versionID := ocmUtils.CreateVersionId(stateVersion, channelGroup)
	response, err := b.VersionCollection.Version(versionID).Get().Send()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve version %s: %w", stateVersion, err)
	}
	v := response.Body()

	if !v.Enabled() {
		return "", fmt.Errorf("version %s is not enabled", stateVersion)
	}
	if !v.ROSAEnabled() {
		return "", fmt.Errorf("version %s is not ROSA enabled", stateVersion)
	}
	if v.ChannelGroup() != channelGroup {
		return "", fmt.Errorf(
			"version %s is in channel group %q, expected %q",
			stateVersion, v.ChannelGroup(), channelGroup)
	}
	if topology == Hcp && !v.HostedControlPlaneEnabled() {
		return "", fmt.Errorf("version %s is not enabled for hosted control plane", stateVersion)
	}
	if len(optionalChannel) > 0 && optionalChannel[0] != "" {
		ch := optionalChannel[0]
		if !slices.Contains(v.AvailableChannels(), ch) {
			return "", fmt.Errorf("version %s does not support channel %s", stateVersion, ch)
		}
	}
	if rawID := v.RawID(); rawID != "" {
		return rawID, nil
	}
	return stateVersion, nil
}

// filterByAvailableChannel filters versions to only those that support the specified channel
func filterByAvailableChannel(versions []*cmv1.Version, channel string) (filtered []*cmv1.Version) {
	for _, v := range versions {
		if slices.Contains(v.AvailableChannels(), channel) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// convertVersionsToList converts Version objects to a list of version strings
func convertVersionsToList(versions []*cmv1.Version) (versionList []string) {
	for _, v := range versions {
		versionList = append(versionList, v.RawID())
	}
	return versionList
}

// getVersionList returns a list of versions for the given channel group, sorted by
// descending semver. Optionally filters by a specific channel if provided.
func (b *BaseCluster) getVersionList(ctx context.Context,
	topology ClusterTopology, channelGroup string, optionalChannel ...string) (versionList []string, err error) {
	vs, err := b.getVersions(ctx, topology, channelGroup)
	if err != nil {
		err = fmt.Errorf("Failed to retrieve versions: %s", err)
		return
	}

	// Filter by specific channel if provided (e.g., "stable-4.10")
	if len(optionalChannel) > 0 && optionalChannel[0] != "" {
		stateChannel := optionalChannel[0]
		vs = filterByAvailableChannel(vs, stateChannel)
		if len(vs) == 0 {
			err = fmt.Errorf("no versions found for channel %s", stateChannel)
			return
		}
	}

	versionList = convertVersionsToList(vs)

	if len(versionList) == 0 {
		err = fmt.Errorf("Could not find versions")
		return
	}

	return
}

func (b *BaseCluster) getVersions(ctx context.Context, topology ClusterTopology, channelGroup string) (versions []*cmv1.Version, err error) {
	page := 1
	size := 100
	searchParams := []string{
		"enabled = 'true'",
		"rosa_enabled = 'true'",
		fmt.Sprintf("channel_group = '%s'", channelGroup),
	}
	if topology == Hcp {
		searchParams = append(searchParams, "hosted_control_plane_enabled = 'true'")
	}
	filter := strings.Join(searchParams, " AND ")
	for {
		var response *cmv1.VersionsListResponse
		response, err = b.VersionCollection.List().
			Search(filter).
			Order("default desc, id desc").
			Page(page).
			Size(size).
			Send()
		if err != nil {
			tflog.Debug(ctx, err.Error())
			return nil, err
		}
		versions = append(versions, response.Items().Slice()...)
		if response.Size() < size {
			break
		}
		page++
	}

	// Sort list in descending order
	sort.Slice(versions, func(i, j int) bool {
		a, erra := semver.NewVersion(versions[i].RawID())
		b, errb := semver.NewVersion(versions[j].RawID())
		if erra != nil || errb != nil {
			return false
		}
		return a.GreaterThan(b)
	})

	return
}
