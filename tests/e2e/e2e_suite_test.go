package e2e

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

var ctx context.Context
var token string
var clusterID string
var profile *CI.Profile

func TestRHCSProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e tests suite")
}

var _ = BeforeSuite(func() {
	token = os.Getenv(CON.TokenENVName)
	var err error

	profile = CI.LoadProfileYamlFileByENV()
	clusterID, err = CI.PrepareRHCSClusterByProfileENV()
	Expect(err).ToNot(HaveOccurred())
	ctx = context.Background()
})
