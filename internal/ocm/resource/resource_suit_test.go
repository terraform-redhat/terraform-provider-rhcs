// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OCM Resource Suite")
}
