package types

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm"
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
	OcmClient         *ocm.Client
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
