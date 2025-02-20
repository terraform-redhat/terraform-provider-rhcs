package proxy

import (
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
)

func TestResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Proxy utils test suite")
}

var _ = Describe("User arn property test", func() {
	It("assumed-role with path", func() {
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com,s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com",
			",")).To(Equal(""))
	})
})
