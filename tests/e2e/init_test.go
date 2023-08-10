package e2e

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

var ctx context.Context
var token string

func TestRHCSProvider(t *testing.T) {
	token = os.Getenv(CON.TokenENVName)
	ctx = context.Background()
	RegisterFailHandler(Fail)
	RunSpecs(t, "RHCS Provider Test")

}
