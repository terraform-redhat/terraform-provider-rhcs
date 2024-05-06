package common

import (
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
)

func TestResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Rosa HCP Resource Suite")
}

var _ = Describe("User arn property test", func() {
	It("assumed-role with path", func() {
		Expect(UserArnRE.MatchString("arn:aws:sts::123456789012:assumed-role/123456789012-poweruser/someuser")).To(BeTrue())
	})

	It("assumed-role", func() {
		Expect(UserArnRE.MatchString("arn:aws:sts::123456789012:assumed-role/someuser")).To(BeTrue())
	})

	It("user", func() {
		Expect(UserArnRE.MatchString("arn:aws:iam::123456789012:user/dummy")).To(BeTrue())
	})

	It("user with path", func() {
		Expect(UserArnRE.MatchString("arn:aws:iam::123456789012:user/path/dummy")).To(BeFalse())
	})
})
