// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package awstags

import (
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
)

func TestAWSTags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS tags Suite")
}
