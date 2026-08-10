// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHyperfleetSanity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hyperfleet E2E")
}
