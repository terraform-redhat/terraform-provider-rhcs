package e2e

import (
	"context"
	"strings"
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

var _ = BeforeSuite(func() {
	token = config.GetRHCSOCMToken()
	var err error

	err = helper.AlignRHCSSourceVersion(config.GetManifestsDir())
	Expect(err).ToNot(HaveOccurred())

	// Determine if we should skip cluster profile loading
	// Skip only if CLUSTER_PROFILE is not set AND we're running non-cluster tests only
	clusterProfile := config.GetClusterProfile()
	labelFilter := GinkgoLabelFilter()

	// Extract the label string from the constant (Labels is []string)
	nonClusterTestLabel := ci.NonClusterTest[0]

	// Simple heuristic: skip cluster profile if not set and filter explicitly includes non-cluster-test
	// without negation (!) - handles "non-cluster-test", "non-cluster-test && day1", etc.
	skipClusterProfile := clusterProfile == "" && labelFilter != "" &&
		strings.Contains(labelFilter, nonClusterTestLabel) &&
		!strings.Contains(labelFilter, "!"+nonClusterTestLabel)

	if !skipClusterProfile {
		// Load cluster profile - this will fail if CLUSTER_PROFILE is not set
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		if err != nil {
			if clusterProfile == "" {
				Fail("CLUSTER_PROFILE environment variable is required for cluster tests. " +
					"To run only non-cluster tests, use: ginkgo --label-filter=" + nonClusterTestLabel)
			}
			Expect(err).ToNot(HaveOccurred())
		}
		clusterID, err = profileHandler.RetrieveClusterID()
		Expect(err).ToNot(HaveOccurred())
	}
	// If skipClusterProfile is true, non-cluster tests will create their own standalone handlers

	ctx = context.Background()
})
