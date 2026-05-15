// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

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

var _ = Describe("RemoveNoProxyZeroEgressDefaultDomains", func() {
	It("removes original zero egress default domains", func() {
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com",
			",")).To(Equal(""))
	})

	It("removes new zero egress default domains", func() {
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			".s3.us-east-1.amazonaws.com,api.ecr.us-east-1.amazonaws.com,.dkr.ecr.us-east-1.amazonaws.com",
			",")).To(Equal(""))
	})

	It("removes all five zero egress default domains", func() {
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.us-east-1.amazonaws.com,.s3.us-east-1.amazonaws.com,sts.us-east-1.amazonaws.com,api.ecr.us-east-1.amazonaws.com,.dkr.ecr.us-east-1.amazonaws.com",
			",")).To(Equal(""))
	})

	It("preserves user-specified no_proxy entries", func() {
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"test.example.com,s3.dualstack.us-east-1.amazonaws.com,.s3.us-east-1.amazonaws.com,sts.us-east-1.amazonaws.com,api.ecr.us-east-1.amazonaws.com,.dkr.ecr.us-east-1.amazonaws.com",
			",")).To(Equal("test.example.com"))
	})

	It("removes duplicated zero egress default domains", func() {
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com,s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com",
			",")).To(Equal(""))
	})

	It("handles govcloud regions", func() {
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.us-gov-west-1.amazonaws.com,.s3.us-gov-west-1.amazonaws.com,sts.us-gov-west-1.amazonaws.com,api.ecr.us-gov-west-1.amazonaws.com,.dkr.ecr.us-gov-west-1.amazonaws.com",
			",")).To(Equal(""))
	})
})
