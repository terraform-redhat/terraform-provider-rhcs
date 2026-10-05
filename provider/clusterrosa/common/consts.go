// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"regexp"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
)

const (
	VersionPrefix = "openshift-v"

	tagsPrefix             = "rosa_"
	TagsOpenShiftVersion   = tagsPrefix + "openshift_version"
	PropertyRosaTfVersion  = tagsPrefix + "tf_version"
	PropertyRosaTfCommit   = tagsPrefix + "tf_commit"
	PropertyRosaCreatorArn = tagsPrefix + "creator_arn"

	MaxHCPClusterWaitTimeoutInMinutes  = int64(45)
	MaxClusterWaitTimeoutInMinutes     = int64(60)
	MaxMachinePoolWaitTimeoutInMinutes = int64(60)
	DefaultPollingIntervalInMinutes    = 2
	NonPositiveTimeoutSummary          = "Can't poll cluster state with a non-positive timeout"
	NonPositiveTimeoutFormat           = "Can't poll state of cluster with identifier '%s', the timeout that was set is not a positive number"
	DestroyTimeoutNotCompleteSummary   = "Cluster wasn't deleted yet"
	DestroyTimeoutNotCompleteFormat    = "The cluster with identifier '%s' is not deleted yet, " +
		"but the polling finished due to a timeout (%d minutes). " +
		"The resource was kept in Terraform state so dependent STS resources (IAM roles, OIDC provider) " +
		"are not destroyed while OpenShift Cluster Manager (OCM) uninstall may still be in progress. " +
		"Resolve the OCM uninstall issue or increase destroy_timeout, then retry terraform destroy."

	MaxClusterNameLength         = 54
	MaxClusterDomainPrefixLength = 15

	// LogFieldClusterID is the tflog field key for cluster identifiers.
	LogFieldClusterID = "cluster_id"
)

var UserArnRE = regexp.MustCompile("^(arn:(?:aws|aws-us-gov|aws-cn):(?:iam|sts)::\\d{12}(?:|:(?:root|user|assumed-role|role)(?:\\/?.+\\/?)?)(?:\\/[0-9A-Za-z\\+\\.@_,-]{1,64}))$")

// StandardWorkerPoolNameRE matches HCP default worker pools ("workers" or
// "workers-N" for multi-AZ). Custom Classic pools may also use names that match
// this pattern; Classic default-pool identity is the OCM ID "worker".
var StandardWorkerPoolNameRE = regexp.MustCompile(`^workers?(-[0-9]+)?$`)

var OCMProperties = map[string]string{
	PropertyRosaTfVersion: build.Version,
	PropertyRosaTfCommit:  build.Commit,
}
