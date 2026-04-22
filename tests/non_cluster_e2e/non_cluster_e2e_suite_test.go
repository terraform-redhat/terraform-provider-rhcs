package non_cluster_e2e

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var ctx context.Context
var token string

func TestNonClusterE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Non-cluster e2e tests suite")
}

var _ = BeforeSuite(func() {
	token = config.GetRHCSOCMToken()
	var err error

	err = helper.AlignRHCSSourceVersion(config.GetManifestsDir())
	Expect(err).ToNot(HaveOccurred())

	ctx = context.Background()
})
