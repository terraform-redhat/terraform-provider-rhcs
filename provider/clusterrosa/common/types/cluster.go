package types

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
// available in the channel group
func (b *BaseCluster) GetAndValidateVersionInChannelGroup(ctx context.Context,
	topology ClusterTopology,
	stateChannelGroup string, stateVersion string) (string, error) {
	versionList, err := b.getVersionList(ctx, topology, stateChannelGroup)
	if err != nil {
		return "", err
	}

	version := versionList[0]
	if stateVersion != "" {
		version = stateVersion
	}

	tflog.Debug(ctx, fmt.Sprintf("Validating if cluster version %s is in the list of supported versions: %v", version, versionList))
	for _, v := range versionList {
		if v == version {
			return version, nil
		}
	}

	return "", fmt.Errorf("version %s is not in the list of supported versions: %v", version, versionList)
}

// getVersionList returns a list of versions for the given channel group, sorted by
// descending semver
func (b *BaseCluster) getVersionList(ctx context.Context,
	topology ClusterTopology, channelGroup string) (versionList []string, err error) {
	vs, err := b.getVersions(ctx, topology, channelGroup)
	if err != nil {
		err = fmt.Errorf("Failed to retrieve versions: %s", err)
		return
	}

	for _, v := range vs {
		versionList = append(versionList, v.RawID())
	}

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

// getAvailableChannelsForVersion fetches all available channels for a version
// by querying all variants (stable, fast, candidate, etc.) and combining their
// available_channels into a unique list.
func (b *BaseCluster) getAvailableChannelsForVersion(ctx context.Context,
	topology ClusterTopology, versionID string) ([]string, error) {
	searchParams := []string{
		"enabled = 'true'",
		"rosa_enabled = 'true'",
		fmt.Sprintf("raw_id = '%s'", versionID),
	}
	if topology == Hcp {
		searchParams = append(searchParams, "hosted_control_plane_enabled = 'true'")
	}
	filter := strings.Join(searchParams, " AND ")

	response, err := b.VersionCollection.List().
		Search(filter).
		Order("default desc, id desc").
		Page(1).
		Size(10).
		Send()
	if err != nil {
		tflog.Debug(ctx, err.Error())
		return nil, err
	}

	if response.Total() == 0 {
		return nil, fmt.Errorf("version %s not found", versionID)
	}

	// Collect all available_channels from all variants and deduplicate
	channelSet := make(map[string]struct{})
	for _, version := range response.Items().Slice() {
		for _, channel := range version.AvailableChannels() {
			channelSet[channel] = struct{}{}
		}
	}

	// Convert map to slice
	var availableChannels []string
	for channel := range channelSet {
		availableChannels = append(availableChannels, channel)
	}

	return availableChannels, nil
}

// ValidateChannelVersionCompatibility validates that the specified channel
// is available for the given version.
// This prevents invalid configurations where a channel like "stable-4.16"
// is used with a version "4.17.x" that may exist in the "stable" channel group
// but not in the specific "stable-4.16" channel.
func (b *BaseCluster) ValidateChannelVersionCompatibility(ctx context.Context,
	topology ClusterTopology, channel string, version string) error {
	if channel == "" || version == "" {
		return nil
	}

	// Fetch all available channels for this version across all variants
	availableChannels, err := b.getAvailableChannelsForVersion(ctx, topology, version)
	if err != nil {
		return fmt.Errorf("failed to fetch version %s: %w", version, err)
	}

	// Check if the specified channel is in the combined available channels
	if !slices.Contains(availableChannels, channel) {
		return fmt.Errorf(
			"channel '%s' is not available for version '%s'. Available channels: %v",
			channel, version, availableChannels,
		)
	}

	tflog.Debug(ctx, fmt.Sprintf("Channel %s is available for version %s", channel, version))
	return nil
}
