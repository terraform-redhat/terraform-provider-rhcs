package common

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/build"
)

const (
	VersionPrefix = "openshift-v"

	tagsPrefix            = "rosa_"
	TagsOpenShiftVersion  = tagsPrefix + "openshift_version"
	PropertyRosaTfVersion = tagsPrefix + "tf_version"
	PropertyRosaTfCommit  = tagsPrefix + "tf_commit"

	DefaultWaitTimeoutForHCPControlPlaneInMinutes = int64(20)
	DefaultWaitTimeoutInMinutes                   = int64(60)
	DefaultPollingIntervalInMinutes               = 2
	NonPositiveTimeoutSummary                     = "Can't poll cluster state with a non-positive timeout"
	NonPositiveTimeoutFormat                      = "Can't poll state of cluster with identifier '%s', the timeout that was set is not a positive number"

	MaxClusterNameLength = 15
)

var OCMProperties = map[string]string{
	PropertyRosaTfVersion: build.Version,
	PropertyRosaTfCommit:  build.Commit,
}
