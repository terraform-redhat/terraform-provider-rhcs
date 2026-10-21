// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSeaplaneSanity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Seaplane E2E")
}
