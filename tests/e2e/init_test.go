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

func TestRHCSProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	token = os.Getenv(CON.TokenENVName)
	clusterID = CI.PrepareRHCSClusterByProfileENV()
	ctx = context.Background()
	RunSpecs(t, "RHCS Provider Test")

}
