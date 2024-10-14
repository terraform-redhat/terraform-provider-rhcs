package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
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
	token = cms.RHCSOCMToken
	var err error

	err = helper.AlignRHCSSourceVersion(manifests.ManifestsConfigurationDir)
	Expect(err).ToNot(HaveOccurred())

	profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
	Expect(err).ToNot(HaveOccurred())
	clusterID, err = profileHandler.RetrieveClusterID()
	Expect(err).ToNot(HaveOccurred())
	ctx = context.Background()
})
