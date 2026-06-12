package sts

import (
	"context"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-online/ocm-common/pkg/aws/ststrust"
)

var _ = Describe("ValidateTrustPolicyExternalID", func() {
	const (
		installerARN = "arn:aws:iam::123456789012:role/my-installer"
		supportARN   = "arn:aws:iam::123456789012:role/my-support"
		externalID   = "valid-external-id-123"
	)

	AfterEach(func() {
		TrustPolicyValidator = validateTrustPolicyExternalIDWithAWS
	})

	It("returns nil when external ID is empty and roles do not require one", func() {
		TrustPolicyValidator = func(
			_ context.Context, entered, _, _, _ string,
		) (warning error, err error) {
			Expect(entered).To(BeEmpty())
			return validateRequiredTrustPolicyExternalIDUnset(
				policyWithoutExternalID(),
				policyWithoutExternalID(),
			)
		}

		warning, err := ValidateTrustPolicyExternalID(context.Background(), "", installerARN, supportARN, "us-east-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(warning).To(BeNil())
	})

	It("returns nil when external ID is empty and role ARNs are missing", func() {
		warning, err := ValidateTrustPolicyExternalID(context.Background(), "", "", supportARN, "us-east-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(warning).To(BeNil())
	})

	It("requires installer and support role ARNs when external ID is set", func() {
		warning, err := ValidateTrustPolicyExternalID(context.Background(), externalID, "", supportARN, "us-east-1")
		Expect(err).To(HaveOccurred())
		Expect(warning).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("installer and support role ARNs are required"))
	})

	It("validates format before calling TrustPolicyValidator", func() {
		TrustPolicyValidator = func(context.Context, string, string, string, string) (error, error) {
			Fail("TrustPolicyValidator should not be called for invalid format")
			return nil, nil
		}

		warning, err := ValidateTrustPolicyExternalID(context.Background(), "x", installerARN, supportARN, "us-east-1")
		Expect(err).To(HaveOccurred())
		Expect(warning).To(BeNil())
		Expect(errors.Is(err, ststrust.ErrExternalIDFormat)).To(BeTrue())
	})

	It("delegates membership validation to TrustPolicyValidator", func() {
		TrustPolicyValidator = func(
			_ context.Context, entered, _, _, _ string,
		) (warning error, err error) {
			Expect(entered).To(Equal(externalID))
			return nil, ststrust.ValidateEnteredForRoleTrustPolicies(
				entered,
				policyWithExternalID(externalID),
				policyWithExternalID(externalID),
			)
		}

		warning, err := ValidateTrustPolicyExternalID(context.Background(), externalID, installerARN, supportARN, "us-east-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(warning).To(BeNil())
	})

	It("returns membership errors from TrustPolicyValidator", func() {
		TrustPolicyValidator = func(
			_ context.Context, entered, _, _, _ string,
		) (warning error, err error) {
			return nil, ststrust.ValidateEnteredForRoleTrustPolicies(
				entered,
				policyWithExternalID("other-id"),
				policyWithExternalID(externalID),
			)
		}

		warning, err := ValidateTrustPolicyExternalID(context.Background(), externalID, installerARN, supportARN, "us-east-1")
		Expect(err).To(HaveOccurred())
		Expect(warning).To(BeNil())
		Expect(errors.Is(err, ststrust.ErrExternalIDNotInTrustPolicy)).To(BeTrue())
	})
})

var _ = Describe("validateRequiredTrustPolicyExternalIDUnset", func() {
	const externalID = "discovered-external-id"

	It("returns nil when neither role defines an external ID", func() {
		warning, err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithoutExternalID(),
			policyWithoutExternalID(),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(warning).To(BeNil())
	})

	It("requires explicit trust_policy_external_id when a single ID is discoverable", func() {
		warning, err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithExternalID(externalID),
			policyWithExternalID(externalID),
		)
		Expect(err).To(HaveOccurred())
		Expect(warning).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("set sts.trust_policy_external_id = \"discovered-external-id\""))
	})

	It("fails when installer and support define external IDs with no value in common", func() {
		warning, err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithExternalID("installer-id"),
			policyWithExternalID("support-id"),
		)
		Expect(err).To(HaveOccurred())
		Expect(warning).To(BeNil())
		Expect(errors.Is(err, errMismatchedSTSExternalIDTrustPolicies)).To(BeTrue())
	})

	It("warns when multiple external IDs are ambiguous", func() {
		warning, err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithExternalIDs("id-a", "id-b"),
			policyWithoutExternalID(),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(warning).To(HaveOccurred())
		Expect(warning.Error()).To(ContainSubstring("sts.trust_policy_external_id"))
	})
})

var _ = Describe("roleNameFromARN", func() {
	It("extracts role name from a valid ARN", func() {
		name, err := roleNameFromARN("arn:aws:iam::123456789012:role/path/my-role")
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("path/my-role"))
	})

	It("rejects non-role ARNs", func() {
		_, err := roleNameFromARN("arn:aws:s3:::my-bucket")
		Expect(err).To(HaveOccurred())
	})
})

// policyWithExternalID returns IAM trust policy JSON with a single sts:ExternalId condition.
func policyWithExternalID(externalID string) string {
	return `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": "sts:AssumeRole",
			"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
			"Condition": {"StringEquals": {"sts:ExternalId": "` + externalID + `"}}
		}]
	}`
}

// policyWithExternalIDs returns IAM trust policy JSON with multiple sts:ExternalId condition values.
func policyWithExternalIDs(externalIDs ...string) string {
	values := make([]string, len(externalIDs))
	for i, externalID := range externalIDs {
		values[i] = `"` + externalID + `"`
	}
	return `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": "sts:AssumeRole",
			"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
			"Condition": {"StringEquals": {"sts:ExternalId": [` + strings.Join(values, ", ") + `]}}
		}]
	}`
}

// policyWithoutExternalID returns IAM trust policy JSON without an sts:ExternalId condition.
func policyWithoutExternalID() string {
	return `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": "sts:AssumeRole",
			"Principal": {"AWS": "arn:aws:iam::123456789012:root"}
		}]
	}`
}
