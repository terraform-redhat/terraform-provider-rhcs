/*
Copyright (c) 2021 Red Hat, Inc.

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

package hyperfleet

import (
	"regexp"
	"strings"

	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// azRegionRE extracts the AWS region prefix from any AZ name, including
// standard (us-east-1a), GovCloud (us-gov-east-1a), Local Zone
// (us-east-1-bos-1a), and Wavelength Zone (us-east-1-wl1-bos-wlz-1).
var azRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

// regionFromAZ derives the AWS region from an availability zone name.
func regionFromAZ(az string) string {
	return azRegionRE.FindString(az)
}

// prefixAndPartitionFromRolesRef derives the operator roles prefix and AWS
// partition from a cluster's RolesRef (used during import).
func prefixAndPartitionFromRolesRef(rolesRef hypershiftv1beta1.AWSRolesRef) (prefix, partition string) {
	arn := rolesRef.NodePoolManagementARN
	// ARN format: arn:<partition>:iam::<account>:role/<prefix>-node-pool-management
	parts := strings.SplitN(arn, ":", 6)
	if len(parts) < 6 {
		return "", "aws"
	}
	partition = parts[1]
	slash := strings.LastIndex(parts[5], "/")
	if slash < 0 {
		return "", partition
	}
	roleName := parts[5][slash+1:]
	prefix, _ = strings.CutSuffix(roleName, "-node-pool-management")
	return prefix, partition
}
