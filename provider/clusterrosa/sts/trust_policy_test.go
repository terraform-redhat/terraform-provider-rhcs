// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package sts

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-online/ocm-common/pkg/aws/ststrust"
)

// stubIAMClient implements iamRoleGetter for deterministic offline tests.
type stubIAMClient struct {
	roles map[string]*iam.GetRoleOutput
	err   error
}

func (s *stubIAMClient) GetRole(_ context.Context, input *iam.GetRoleInput, _ ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	if s.err != nil {
		return nil, s.err
	}
	if out, ok := s.roles[aws.ToString(input.RoleName)]; ok {
		return out, nil
	}
	return nil, fmt.Errorf("NoSuchEntity: role %q not found", aws.ToString(input.RoleName))
}

func stubLoader(client iamRoleGetter) func(context.Context, string) (*iamTrustPolicyLoader, error) {
	return func(context.Context, string) (*iamTrustPolicyLoader, error) {
		return &iamTrustPolicyLoader{client: client}, nil
	}
}

func stubLoaderError(err error) func(context.Context, string) (*iamTrustPolicyLoader, error) {
	return func(context.Context, string) (*iamTrustPolicyLoader, error) {
		return nil, err
	}
}

func roleOutput(name, policyDoc string) *iam.GetRoleOutput {
	return &iam.GetRoleOutput{
		Role: &iamtypes.Role{
			RoleName:                 aws.String(name),
			AssumeRolePolicyDocument: aws.String(policyDoc),
		},
	}
}

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
		) error {
			Expect(entered).To(BeEmpty())
			return validateRequiredTrustPolicyExternalIDUnset(
				policyWithoutExternalID(),
				policyWithoutExternalID(),
			)
		}

		err := ValidateTrustPolicyExternalID(context.Background(), "", installerARN, supportARN, "us-east-1")
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns nil when external ID is empty and role ARNs are missing", func() {
		err := ValidateTrustPolicyExternalID(context.Background(), "", "", supportARN, "us-east-1")
		Expect(err).NotTo(HaveOccurred())
	})

	It("requires installer and support role ARNs when external ID is set", func() {
		err := ValidateTrustPolicyExternalID(context.Background(), externalID, "", supportARN, "us-east-1")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("installer and support role ARNs are required"))
	})

	It("validates format before calling TrustPolicyValidator", func() {
		TrustPolicyValidator = func(context.Context, string, string, string, string) error {
			Fail("TrustPolicyValidator should not be called for invalid format")
			return nil
		}

		err := ValidateTrustPolicyExternalID(context.Background(), "x", installerARN, supportARN, "us-east-1")
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ststrust.ErrExternalIDFormat)).To(BeTrue())
	})

	It("delegates membership validation to TrustPolicyValidator", func() {
		TrustPolicyValidator = func(
			_ context.Context, entered, _, _, _ string,
		) error {
			Expect(entered).To(Equal(externalID))
			return ststrust.ValidateEnteredForRoleTrustPolicies(
				entered,
				policyWithExternalID(externalID),
				policyWithExternalID(externalID),
			)
		}

		err := ValidateTrustPolicyExternalID(context.Background(), externalID, installerARN, supportARN, "us-east-1")
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns membership errors from TrustPolicyValidator", func() {
		TrustPolicyValidator = func(
			_ context.Context, entered, _, _, _ string,
		) error {
			return ststrust.ValidateEnteredForRoleTrustPolicies(
				entered,
				policyWithExternalID("other-id"),
				policyWithExternalID(externalID),
			)
		}

		err := ValidateTrustPolicyExternalID(context.Background(), externalID, installerARN, supportARN, "us-east-1")
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ststrust.ErrExternalIDNotInTrustPolicy)).To(BeTrue())
	})
})

var _ = Describe("validateRequiredTrustPolicyExternalIDUnset", func() {
	const externalID = "discovered-external-id"

	It("returns nil when neither role defines an external ID", func() {
		err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithoutExternalID(),
			policyWithoutExternalID(),
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("requires explicit trust_policy_external_id when a single ID is discoverable", func() {
		err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithExternalID(externalID),
			policyWithExternalID(externalID),
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("set sts.trust_policy_external_id = \"discovered-external-id\""))
	})

	It("fails when installer and support define external IDs with no value in common", func() {
		err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithExternalID("installer-id"),
			policyWithExternalID("support-id"),
		)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, errMismatchedSTSExternalIDTrustPolicies)).To(BeTrue())
	})

	It("fails when multiple external IDs are ambiguous", func() {
		err := validateRequiredTrustPolicyExternalIDUnset(
			policyWithExternalIDs("id-a", "id-b"),
			policyWithoutExternalID(),
		)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, errAmbiguousSTSExternalIDTrustPolicies)).To(BeTrue())
	})
})

var _ = Describe("hasMismatchedSTSExternalIDTrustPolicies", func() {
	It("returns false when installer IDs are empty", func() {
		Expect(hasMismatchedSTSExternalIDTrustPolicies(nil, []string{"a"})).To(BeFalse())
	})

	It("returns false when support IDs are empty", func() {
		Expect(hasMismatchedSTSExternalIDTrustPolicies([]string{"a"}, nil)).To(BeFalse())
	})

	It("returns false when there is overlap", func() {
		Expect(hasMismatchedSTSExternalIDTrustPolicies([]string{"a", "b"}, []string{"b", "c"})).To(BeFalse())
	})

	It("returns true when there is no overlap", func() {
		Expect(hasMismatchedSTSExternalIDTrustPolicies([]string{"a"}, []string{"b"})).To(BeTrue())
	})
})

var _ = Describe("isSTSExternalIDDiscoveryAmbiguous", func() {
	It("returns false when both ID lists are empty", func() {
		Expect(isSTSExternalIDDiscoveryAmbiguous("", nil, nil)).To(BeFalse())
	})

	It("returns false when discovered is non-empty", func() {
		Expect(isSTSExternalIDDiscoveryAmbiguous("found", []string{"a"}, nil)).To(BeFalse())
	})

	It("returns true when discovered is empty but IDs exist", func() {
		Expect(isSTSExternalIDDiscoveryAmbiguous("", []string{"a", "b"}, nil)).To(BeTrue())
	})
})

var _ = Describe("roleNameFromARN", func() {
	It("extracts role name without path prefix", func() {
		name, err := roleNameFromARN("arn:aws:iam::123456789012:role/path/my-role")
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("my-role"))
	})

	It("extracts role name from advanced path ARN used in e2e profiles", func() {
		name, err := roleNameFromARN(
			"arn:aws:iam::123456789012:role/advanced/ci-rhcs-hcp-ad-1h2-pr-HCP-ROSA-Installer-Role",
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("ci-rhcs-hcp-ad-1h2-pr-HCP-ROSA-Installer-Role"))
	})

	It("extracts role name when ARN has no path", func() {
		name, err := roleNameFromARN("arn:aws:iam::123456789012:role/my-installer")
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("my-installer"))
	})

	It("rejects non-role ARNs", func() {
		_, err := roleNameFromARN("arn:aws:s3:::my-bucket")
		Expect(err).To(HaveOccurred())
	})

	It("rejects role ARN with empty role name", func() {
		_, err := roleNameFromARN("arn:aws:iam::123456789012:role/")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("missing role name"))
	})

	It("rejects role ARN with empty role name after path", func() {
		_, err := roleNameFromARN("arn:aws:iam::123456789012:role/path/")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("missing role name"))
	})
})

var _ = Describe("trustPolicyJSONFromRole", func() {
	It("preserves literal plus signs in RFC 3986 percent-encoded policy documents", func() {
		role := iamtypes.Role{
			RoleName:                 aws.String("test-role"),
			AssumeRolePolicyDocument: aws.String(`%7B%22ExternalId%22%3A%22a+b%22%7D`),
		}

		decoded, err := trustPolicyJSONFromRole(role)
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded).To(Equal(`{"ExternalId":"a+b"}`))
	})

	It("returns error when assume role policy document is nil", func() {
		role := iamtypes.Role{
			RoleName: aws.String("test-role"),
		}

		_, err := trustPolicyJSONFromRole(role)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no assume role policy document"))
	})
})

var _ = Describe("validateTrustPolicyExternalIDWithAWS", func() {
	const (
		installerARN = "arn:aws:iam::123456789012:role/my-installer"
		supportARN   = "arn:aws:iam::123456789012:role/my-support"
		externalID   = "valid-external-id-123"
	)

	savedLoader := newTrustPolicyLoader

	AfterEach(func() {
		newTrustPolicyLoader = savedLoader
	})

	It("returns nil when IS_TEST is set", func() {
		os.Setenv("IS_TEST", "true")
		defer os.Unsetenv("IS_TEST")

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(),
			externalID,
			installerARN,
			supportARN,
			"us-east-1",
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns error when loader construction fails", func() {
		os.Unsetenv("IS_TEST")
		newTrustPolicyLoader = stubLoaderError(fmt.Errorf("no credentials"))

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, supportARN, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to load AWS configuration"))
	})

	It("returns error when installer role lookup fails", func() {
		os.Unsetenv("IS_TEST")
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			err: fmt.Errorf("access denied"),
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, supportARN, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to read installer role trust policy"))
	})

	It("returns error when support role lookup fails", func() {
		os.Unsetenv("IS_TEST")
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			roles: map[string]*iam.GetRoleOutput{
				"my-installer": roleOutput("my-installer", policyWithExternalID(externalID)),
			},
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, supportARN, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to read support role trust policy"))
	})

	It("validates entered external ID against both policies", func() {
		os.Unsetenv("IS_TEST")
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			roles: map[string]*iam.GetRoleOutput{
				"my-installer": roleOutput("my-installer", policyWithExternalID(externalID)),
				"my-support":   roleOutput("my-support", policyWithExternalID(externalID)),
			},
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, supportARN, "us-east-1",
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns error when entered external ID is not in trust policies", func() {
		os.Unsetenv("IS_TEST")
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			roles: map[string]*iam.GetRoleOutput{
				"my-installer": roleOutput("my-installer", policyWithExternalID("other-id")),
				"my-support":   roleOutput("my-support", policyWithExternalID("other-id")),
			},
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, supportARN, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ststrust.ErrExternalIDNotInTrustPolicy)).To(BeTrue())
	})

	It("delegates to validateRequiredTrustPolicyExternalIDUnset when entered is empty", func() {
		os.Unsetenv("IS_TEST")
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			roles: map[string]*iam.GetRoleOutput{
				"my-installer": roleOutput("my-installer", policyWithoutExternalID()),
				"my-support":   roleOutput("my-support", policyWithoutExternalID()),
			},
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), "", installerARN, supportARN, "us-east-1",
		)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("newIAMTrustPolicyLoader", func() {
	It("constructs a loader with a region", func() {
		loader, err := newIAMTrustPolicyLoader(context.Background(), "us-east-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(loader).NotTo(BeNil())
		Expect(loader.client).NotTo(BeNil())
	})

	It("constructs a loader without a region", func() {
		loader, err := newIAMTrustPolicyLoader(context.Background(), "")
		Expect(err).NotTo(HaveOccurred())
		Expect(loader).NotTo(BeNil())
	})
})

var _ = Describe("trustPolicyJSONForRoleARN", func() {
	It("returns empty string for empty role ARN", func() {
		loader := &iamTrustPolicyLoader{client: &stubIAMClient{}}
		result, err := loader.trustPolicyJSONForRoleARN(context.Background(), "")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("returns error for invalid ARN", func() {
		loader := &iamTrustPolicyLoader{client: &stubIAMClient{}}
		_, err := loader.trustPolicyJSONForRoleARN(context.Background(), "not-an-arn")
		Expect(err).To(HaveOccurred())
	})

	It("returns error when GetRole fails", func() {
		loader := &iamTrustPolicyLoader{client: &stubIAMClient{
			err: fmt.Errorf("access denied"),
		}}
		_, err := loader.trustPolicyJSONForRoleARN(
			context.Background(), "arn:aws:iam::123456789012:role/my-role",
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("access denied"))
	})

	It("returns error when GetRole returns nil role", func() {
		loader := &iamTrustPolicyLoader{client: &stubIAMClient{
			roles: map[string]*iam.GetRoleOutput{
				"my-role": {Role: nil},
			},
		}}
		_, err := loader.trustPolicyJSONForRoleARN(
			context.Background(), "arn:aws:iam::123456789012:role/my-role",
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("GetRole returned no role"))
	})

	It("decodes and returns the trust policy document", func() {
		loader := &iamTrustPolicyLoader{client: &stubIAMClient{
			roles: map[string]*iam.GetRoleOutput{
				"my-role": roleOutput("my-role", policyWithoutExternalID()),
			},
		}}
		result, err := loader.trustPolicyJSONForRoleARN(
			context.Background(), "arn:aws:iam::123456789012:role/my-role",
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(ContainSubstring("sts:AssumeRole"))
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
