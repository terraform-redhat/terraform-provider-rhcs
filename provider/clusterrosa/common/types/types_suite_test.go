// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Rosa Common Types Suite")
}
