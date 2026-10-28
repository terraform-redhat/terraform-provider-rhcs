// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var ctx context.Context
var token string
var clusterID string

func TestRHCSProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e tests suite")
}

// clusterIDLabels are the spec labels whose tests rely on BeforeSuite's pre-provisioned
// clusterID (as opposed to e.g. day1-supplemental specs, which create and tear down their
// own clusters via profilehandler.NewRandomProfileHandler and never read clusterID).
var clusterIDLabels = []Labels{ci.Day1Post, ci.Day2, ci.Day3, ci.Upgrade}

// importanceLabels mirrors the importance label every spec in this suite is decorated with.
// Label filters commonly AND a day-label with an importance label (e.g. the general-tests
// step's default "(Critical,High)&&(day1-post,day2)&&!Exclude"), so checking a bare
// day-label in isolation can report a false "no match" even though real specs carrying it
// would be selected.
var importanceLabels = []Labels{ci.Critical, ci.High, ci.Medium, ci.Low}

// suiteNeedsExistingCluster reports whether the configured --label-filter could select any
// spec that depends on clusterID. An empty filter runs everything, so it conservatively
// requires a cluster.
func suiteNeedsExistingCluster() bool {
	filter := GinkgoLabelFilter()
	if filter == "" {
		return true
	}
	for _, label := range clusterIDLabels {
		if label.MatchesLabelFilter(filter) {
			return true
		}
		for _, importance := range importanceLabels {
			if append(label, importance...).MatchesLabelFilter(filter) {
				return true
			}
		}
	}
	return false
}

var _ = BeforeSuite(func() {
	token = config.GetRHCSOCMToken()
	var err error

	err = helper.AlignRHCSSourceVersion(config.GetManifestsDir())
	Expect(err).ToNot(HaveOccurred())

	if config.GetClusterProfile() == "" && !suiteNeedsExistingCluster() {
		ctx = context.Background()
		return
	}

	profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
	Expect(err).ToNot(HaveOccurred())
	clusterID, err = profileHandler.RetrieveClusterID()
	Expect(err).ToNot(HaveOccurred())
	ctx = context.Background()
})
