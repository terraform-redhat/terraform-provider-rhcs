package attrvalidators

import (
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
)

func TestAttrValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Attribute validators Resource Suite")
}
