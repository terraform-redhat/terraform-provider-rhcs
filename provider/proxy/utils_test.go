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
		defaultDomains := []string{
			"s3.dualstack.ap-southeast-4.amazonaws.com",
			"sts.ap-southeast-4.amazonaws.com",
			"012345678912.dkr.ecr.ap-southeast-4.amazonaws.com",
		}
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com",
			",", defaultDomains, "012345678912")).To(Equal(""))
	})

	It("removes new zero egress default domains", func() {
		defaultDomains := []string{
			".s3.us-east-1.amazonaws.com",
			"api.ecr.us-east-1.amazonaws.com",
			".dkr.ecr.us-east-1.amazonaws.com",
		}
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			".s3.us-east-1.amazonaws.com,api.ecr.us-east-1.amazonaws.com,.dkr.ecr.us-east-1.amazonaws.com",
			",", defaultDomains, "")).To(Equal(""))
	})

	It("removes all five zero egress default domains", func() {
		defaultDomains := []string{
			"s3.dualstack.us-east-1.amazonaws.com",
			".s3.us-east-1.amazonaws.com",
			"sts.us-east-1.amazonaws.com",
			"api.ecr.us-east-1.amazonaws.com",
			".dkr.ecr.us-east-1.amazonaws.com",
		}
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.us-east-1.amazonaws.com,.s3.us-east-1.amazonaws.com,sts.us-east-1.amazonaws.com,api.ecr.us-east-1.amazonaws.com,.dkr.ecr.us-east-1.amazonaws.com",
			",", defaultDomains, "")).To(Equal(""))
	})

	It("preserves user-specified no_proxy entries", func() {
		defaultDomains := []string{
			"s3.dualstack.us-east-1.amazonaws.com",
			".s3.us-east-1.amazonaws.com",
			"sts.us-east-1.amazonaws.com",
			"api.ecr.us-east-1.amazonaws.com",
			".dkr.ecr.us-east-1.amazonaws.com",
		}
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"test.example.com,s3.dualstack.us-east-1.amazonaws.com,.s3.us-east-1.amazonaws.com,sts.us-east-1.amazonaws.com,api.ecr.us-east-1.amazonaws.com,.dkr.ecr.us-east-1.amazonaws.com",
			",", defaultDomains, "")).To(Equal("test.example.com"))
	})

	It("removes duplicated zero egress default domains", func() {
		defaultDomains := []string{
			"s3.dualstack.ap-southeast-4.amazonaws.com",
			"sts.ap-southeast-4.amazonaws.com",
			"012345678912.dkr.ecr.ap-southeast-4.amazonaws.com",
		}
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com,s3.dualstack.ap-southeast-4.amazonaws.com,sts.ap-southeast-4.amazonaws.com,012345678912.dkr.ecr.ap-southeast-4.amazonaws.com",
			",", defaultDomains, "012345678912")).To(Equal(""))
	})

	It("handles govcloud regions", func() {
		defaultDomains := []string{
			"s3.dualstack.us-gov-west-1.amazonaws.com",
			".s3.us-gov-west-1.amazonaws.com",
			"sts.us-gov-west-1.amazonaws.com",
			"api.ecr.us-gov-west-1.amazonaws.com",
			".dkr.ecr.us-gov-west-1.amazonaws.com",
		}
		Expect(RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"s3.dualstack.us-gov-west-1.amazonaws.com,.s3.us-gov-west-1.amazonaws.com,sts.us-gov-west-1.amazonaws.com,api.ecr.us-gov-west-1.amazonaws.com,.dkr.ecr.us-gov-west-1.amazonaws.com",
			",", defaultDomains, "")).To(Equal(""))
	})

	// Backward compatibility tests for SREP-3100 ECR format change
	It("filters old ECR format when API returns new format (backward compatibility)", func() {
		// Simulates: cluster created before SREP-3100 with account-prefixed ECR domain
		// API returns new format (.dkr.ecr...), but cluster has old format (account.dkr.ecr...)
		defaultDomains := []string{
			".dkr.ecr.us-east-1.amazonaws.com", // API returns new format
		}
		result := RemoveNoProxyZeroEgressDefaultDomains(
			"123456789012.dkr.ecr.us-east-1.amazonaws.com", // Cluster has old format
			",", defaultDomains, "123456789012")
		Expect(result).To(Equal("")) // Old format should be filtered by regex
	})

	It("filters old ECR format with user domains (backward compatibility)", func() {
		defaultDomains := []string{
			"s3.dualstack.us-east-1.amazonaws.com",
			".dkr.ecr.us-east-1.amazonaws.com", // New format from API
		}
		result := RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"custom.domain,s3.dualstack.us-east-1.amazonaws.com,123456789012.dkr.ecr.us-east-1.amazonaws.com,another.custom.com",
			",", defaultDomains, "123456789012")
		Expect(result).To(Equal("custom.domain,another.custom.com"))
	})

	It("filters both old and new ECR formats", func() {
		defaultDomains := []string{
			".dkr.ecr.us-east-1.amazonaws.com",
		}
		result := RemoveNoProxyZeroEgressDefaultDomains(
			//nolint:lll
			"custom.domain,123456789012.dkr.ecr.us-east-1.amazonaws.com,.dkr.ecr.us-east-1.amazonaws.com",
			",", defaultDomains, "123456789012")
		Expect(result).To(Equal("custom.domain"))
	})

	It("filters old ECR format across different regions", func() {
		defaultDomains := []string{
			".dkr.ecr.ap-southeast-4.amazonaws.com",
		}
		result := RemoveNoProxyZeroEgressDefaultDomains(
			"987654321098.dkr.ecr.ap-southeast-4.amazonaws.com",
			",", defaultDomains, "987654321098")
		Expect(result).To(Equal(""))
	})

	It("does not filter ECR format with different AWS account ID", func() {
		// Simulates: User has a different AWS account's ECR domain in their no_proxy
		// Should NOT be filtered because account ID doesn't match
		defaultDomains := []string{
			".dkr.ecr.us-east-1.amazonaws.com",
		}
		result := RemoveNoProxyZeroEgressDefaultDomains(
			"999888777666.dkr.ecr.us-east-1.amazonaws.com", // Different account
			",", defaultDomains, "123456789012") // Cluster's account
		Expect(result).To(Equal("999888777666.dkr.ecr.us-east-1.amazonaws.com")) // Should NOT be filtered
	})
})
