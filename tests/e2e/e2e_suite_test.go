package e2e

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	H "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var ctx context.Context
var token string
var clusterID string

func TestRHCSProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e tests suite")
}

var _ = BeforeSuite(func() {
	token = os.Getenv(CON.TokenENVName)
	var err error

	err = H.AlignRHCSSourceVersion(CON.ManifestsConfigurationDir)
	Expect(err).ToNot(HaveOccurred())

	clusterID, err = CI.PrepareRHCSClusterByProfileENV()
	Expect(err).ToNot(HaveOccurred())
	ctx = context.Background()
})
