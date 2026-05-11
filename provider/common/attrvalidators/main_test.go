// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

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
