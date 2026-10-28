// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package cms

import (
	"github.com/Masterminds/semver"
	client "github.com/openshift-online/ocm-sdk-go"
)

// Day2AWSTagsMinimalVersion is the first HCP control-plane version that supports editing AWS tags.
const Day2AWSTagsMinimalVersion = "4.20.0-0.a"

// SupportsDay2AWSTags reports whether the cluster control plane supports editing AWS tags.
func SupportsDay2AWSTags(connection *client.Connection, clusterID string) (bool, error) {
	resp, err := RetrieveClusterDetail(connection, clusterID)
	if err != nil {
		return false, err
	}
	current, err := semver.NewVersion(resp.Body().Version().RawID())
	if err != nil {
		return false, err
	}
	return !current.LessThan(semver.MustParse(Day2AWSTagsMinimalVersion)), nil
}

// RetrieveClusterUserTags returns the customer-managed cluster AWS tags, omitting protected tags.
func RetrieveClusterUserTags(connection *client.Connection, clusterID string) (map[string]string, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().
		Parameter("fetchUserTagsOnly", true).Send()
	if err != nil {
		return nil, err
	}
	return resp.Body().AWS().Tags(), nil
}
