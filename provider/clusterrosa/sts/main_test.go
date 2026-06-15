// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package sts

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "STS Suite")
}
