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

func stubLoader(client iamRoleGetter) func(aws.Config) (*iamTrustPolicyLoader, error) {
	return func(aws.Config) (*iamTrustPolicyLoader, error) {
		return &iamTrustPolicyLoader{client: client}, nil
	}
}

func stubLoaderError(err error) func(aws.Config) (*iamTrustPolicyLoader, error) {
	return func(aws.Config) (*iamTrustPolicyLoader, error) {
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

var _ = Describe("parseIAMRoleARN", func() {
	It("accepts a standard IAM role ARN", func() {
		parsed, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/my-role")
		Expect(ok).To(BeTrue())
		Expect(parsed.Service).To(Equal("iam"))
		Expect(parsed.AccountID).To(Equal("123456789012"))
	})

	It("rejects a non-IAM ARN with role/ resource prefix", func() {
		_, ok := parseIAMRoleARN("arn:aws:s3::999999999999:role/fake")
		Expect(ok).To(BeFalse())
	})

	It("rejects an IAM ARN that is not a role", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::123456789012:user/my-user")
		Expect(ok).To(BeFalse())
	})

	It("rejects an ARN with empty account ID", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam:::role/my-role")
		Expect(ok).To(BeFalse())
	})

	It("rejects an ARN ending in role/ with no name", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/")
		Expect(ok).To(BeFalse())
	})

	It("rejects an ARN ending in role/path/ with no name after path", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/path/")
		Expect(ok).To(BeFalse())
	})

	It("accepts a valid role ARN with a path prefix", func() {
		parsed, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/path/my-role")
		Expect(ok).To(BeTrue())
		Expect(parsed.AccountID).To(Equal("123456789012"))
	})

	It("rejects a role name containing a wildcard", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/my-role-*")
		Expect(ok).To(BeFalse())
	})

	It("rejects a role name containing whitespace", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/my role")
		Expect(ok).To(BeFalse())
	})

	It("accepts a role name with valid special characters", func() {
		parsed, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/My_Role+=,.@-test")
		Expect(ok).To(BeTrue())
		Expect(parsed.AccountID).To(Equal("123456789012"))
	})

	It("rejects a leading slash after role/ prefix", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role//my-role")
		Expect(ok).To(BeFalse())
	})

	It("accepts a path with special characters valid in IAM paths", func() {
		parsed, ok := parseIAMRoleARN("arn:aws:iam::123456789012:role/team!/my-role")
		Expect(ok).To(BeTrue())
		Expect(parsed.AccountID).To(Equal("123456789012"))
	})

	It("rejects an ARN with a non-empty region", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam:us-east-1:123456789012:role/my-role")
		Expect(ok).To(BeFalse())
	})

	It("rejects a non-12-digit account ID", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::12345:role/my-role")
		Expect(ok).To(BeFalse())
	})

	It("rejects a non-numeric account ID", func() {
		_, ok := parseIAMRoleARN("arn:aws:iam::12345678901a:role/my-role")
		Expect(ok).To(BeFalse())
	})

	It("accepts a GovCloud IAM role ARN", func() {
		parsed, ok := parseIAMRoleARN("arn:aws-us-gov:iam::123456789012:role/my-role")
		Expect(ok).To(BeTrue())
		Expect(parsed.AccountID).To(Equal("123456789012"))
	})

	It("rejects an unparseable string", func() {
		_, ok := parseIAMRoleARN("not-an-arn")
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("parseIAMRoleARN table-driven", func() {
	type testCase struct {
		arn    string
		accept bool
	}

	DescribeTable("accepts valid IAM role ARNs",
		func(tc testCase) {
			_, ok := parseIAMRoleARN(tc.arn)
			Expect(ok).To(Equal(tc.accept), "ARN: %s", tc.arn)
		},

		// --- Valid IAM role ARNs (must accept) ---
		Entry("simple role", testCase{"arn:aws:iam::123456789012:role/MyRole", true}),
		Entry("role with path", testCase{"arn:aws:iam::123456789012:role/application/MyRole", true}),
		Entry("role with deep path", testCase{"arn:aws:iam::123456789012:role/org/team/app/MyRole", true}),
		Entry("ROSA installer role", testCase{"arn:aws:iam::123456789012:role/ManagedOpenShift-Installer-Role", true}),
		Entry("ROSA support role", testCase{"arn:aws:iam::123456789012:role/ManagedOpenShift-Support-Role", true}),
		Entry("ROSA worker role", testCase{"arn:aws:iam::123456789012:role/ManagedOpenShift-Worker-Role", true}),
		Entry("ROSA control plane role", testCase{"arn:aws:iam::123456789012:role/ManagedOpenShift-ControlPlane-Role", true}),
		Entry("role with advanced path (CI)", testCase{"arn:aws:iam::123456789012:role/advanced/ci-rhcs-hcp-ad-1h2-pr-HCP-ROSA-Installer-Role", true}),
		Entry("service-linked role", testCase{"arn:aws:iam::123456789012:role/aws-service-role/elasticmapreduce.amazonaws.com/AWSServiceRoleForEMR", true}),
		Entry("GovCloud role", testCase{"arn:aws-us-gov:iam::123456789012:role/MyRole", true}),
		Entry("role with special chars +=,.@-", testCase{"arn:aws:iam::123456789012:role/My_Role+=,.@-test", true}),

		// --- IAM non-role resources (must reject) ---
		Entry("IAM user", testCase{"arn:aws:iam::123456789012:user/my-user", false}),
		Entry("IAM user with path", testCase{"arn:aws:iam::123456789012:user/division/my-user", false}),
		Entry("IAM group", testCase{"arn:aws:iam::123456789012:group/my-group", false}),
		Entry("IAM policy", testCase{"arn:aws:iam::123456789012:policy/my-policy", false}),
		Entry("IAM policy with path", testCase{"arn:aws:iam::123456789012:policy/service-role/my-policy", false}),
		Entry("IAM instance profile", testCase{"arn:aws:iam::123456789012:instance-profile/my-profile", false}),
		Entry("IAM OIDC provider", testCase{"arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com", false}),
		Entry("IAM SAML provider", testCase{"arn:aws:iam::123456789012:saml-provider/my-saml", false}),
		Entry("IAM MFA device", testCase{"arn:aws:iam::123456789012:mfa/my-mfa", false}),
		Entry("IAM server certificate", testCase{"arn:aws:iam::123456789012:server-certificate/my-cert", false}),
		Entry("IAM root", testCase{"arn:aws:iam::123456789012:root", false}),

		// --- Non-IAM services (must reject) ---
		Entry("S3 bucket", testCase{"arn:aws:s3:::my-bucket", false}),
		Entry("S3 object", testCase{"arn:aws:s3:::my-bucket/my-key", false}),
		Entry("EC2 instance", testCase{"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0", false}),
		Entry("EC2 VPC", testCase{"arn:aws:ec2:us-east-1:123456789012:vpc/vpc-12345678", false}),
		Entry("EC2 subnet", testCase{"arn:aws:ec2:us-east-1:123456789012:subnet/subnet-12345678", false}),
		Entry("EC2 security group", testCase{"arn:aws:ec2:us-east-1:123456789012:security-group/sg-12345678", false}),
		Entry("Lambda function", testCase{"arn:aws:lambda:us-east-1:123456789012:function:my-function", false}),
		Entry("STS assumed role session", testCase{"arn:aws:sts::123456789012:assumed-role/my-role/my-session", false}),
		Entry("STS federated user", testCase{"arn:aws:sts::123456789012:federated-user/my-user", false}),
		Entry("EKS cluster", testCase{"arn:aws:eks:us-east-1:123456789012:cluster/my-cluster", false}),
		Entry("KMS key", testCase{"arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012", false}),
		Entry("SNS topic", testCase{"arn:aws:sns:us-east-1:123456789012:my-topic", false}),
		Entry("SQS queue", testCase{"arn:aws:sqs:us-east-1:123456789012:my-queue", false}),
		Entry("DynamoDB table", testCase{"arn:aws:dynamodb:us-east-1:123456789012:table/my-table", false}),
		Entry("CloudFormation stack", testCase{"arn:aws:cloudformation:us-east-1:123456789012:stack/my-stack/guid", false}),
		Entry("Secrets Manager secret", testCase{"arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret-AbCdEf", false}),
		Entry("RDS instance", testCase{"arn:aws:rds:us-east-1:123456789012:db:my-database", false}),
		Entry("ElastiCache cluster", testCase{"arn:aws:elasticache:us-east-1:123456789012:cluster:my-cluster", false}),
		Entry("ELB load balancer", testCase{"arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/1234567890", false}),
		Entry("Route53 hosted zone", testCase{"arn:aws:route53:::hostedzone/Z1234567890", false}),
		Entry("CloudWatch log group", testCase{"arn:aws:logs:us-east-1:123456789012:log-group:/aws/lambda/my-function", false}),

		// --- Malformed / edge cases (must reject) ---
		Entry("non-IAM service with role/ resource", testCase{"arn:aws:s3::999999999999:role/fake", false}),
		Entry("wildcard in role name", testCase{"arn:aws:iam::123456789012:role/my-role-*", false}),
		Entry("double slash in resource", testCase{"arn:aws:iam::123456789012:role//my-role", false}),
		Entry("trailing slash", testCase{"arn:aws:iam::123456789012:role/path/", false}),
		Entry("empty after role/", testCase{"arn:aws:iam::123456789012:role/", false}),
		Entry("IAM with region", testCase{"arn:aws:iam:us-east-1:123456789012:role/my-role", false}),
		Entry("short account ID", testCase{"arn:aws:iam::12345:role/my-role", false}),
		Entry("non-numeric account", testCase{"arn:aws:iam::12345678901a:role/my-role", false}),
		Entry("not an ARN", testCase{"not-an-arn", false}),
		Entry("empty string", testCase{"", false}),
	)
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

	It("rejects non-IAM ARN with role/ resource prefix", func() {
		_, err := roleNameFromARN("arn:aws:s3::999999999999:role/fake")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected IAM role resource"))
	})

	It("rejects role ARN with empty role name", func() {
		_, err := roleNameFromARN("arn:aws:iam::123456789012:role/")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected IAM role resource"))
	})

	It("rejects role ARN with empty role name after path", func() {
		_, err := roleNameFromARN("arn:aws:iam::123456789012:role/path/")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected IAM role resource"))
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
	savedCallerAccount := callerAccountID

	BeforeEach(func() {
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "123456789012", nil
		}
	})

	AfterEach(func() {
		newTrustPolicyLoader = savedLoader
		callerAccountID = savedCallerAccount
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

	It("returns CrossAccountWarning when caller account differs from role account", func() {
		os.Unsetenv("IS_TEST")
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "999999999999", nil
		}

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, supportARN, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		var w *CrossAccountWarning
		Expect(errors.As(err, &w)).To(BeTrue())
		Expect(w.CallerAccount).To(Equal("999999999999"))
		Expect(w.RoleAccount).To(Equal("123456789012"))
	})

	It("rejects installer and support roles in different accounts", func() {
		os.Unsetenv("IS_TEST")
		splitAccountSupportARN := "arn:aws:iam::888888888888:role/my-support"

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, splitAccountSupportARN, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		var w *CrossAccountWarning
		Expect(errors.As(err, &w)).To(BeFalse())
		Expect(err.Error()).To(ContainSubstring("must be in the same AWS account"))
		Expect(err.Error()).To(ContainSubstring("123456789012"))
		Expect(err.Error()).To(ContainSubstring("888888888888"))
	})

	It("does not emit CrossAccountWarning for non-IAM ARN with role/ resource prefix", func() {
		os.Unsetenv("IS_TEST")
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "111111111111", nil
		}
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			err: fmt.Errorf("should not be reached"),
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID,
			"arn:aws:s3::999999999999:role/fake",
			"arn:aws:s3::999999999999:role/fake-support",
			"us-east-1",
		)
		Expect(err).To(HaveOccurred())
		var w *CrossAccountWarning
		Expect(errors.As(err, &w)).To(BeFalse(), "non-IAM ARN must not trigger CrossAccountWarning")
		Expect(err.Error()).To(ContainSubstring("expected IAM role resource"))
	})

	It("does not emit CrossAccountWarning for IAM role ARN with invalid name characters", func() {
		os.Unsetenv("IS_TEST")
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "111111111111", nil
		}
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			err: fmt.Errorf("should not be reached"),
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID,
			"arn:aws:iam::999999999999:role/my-role-*",
			"arn:aws:iam::999999999999:role/my-support-*",
			"us-east-1",
		)
		Expect(err).To(HaveOccurred())
		var w *CrossAccountWarning
		Expect(errors.As(err, &w)).To(BeFalse(), "invalid role name must not trigger CrossAccountWarning")
		Expect(err.Error()).To(ContainSubstring("expected IAM role resource"))
	})

	It("skips cross-account check when support ARN is not an IAM role", func() {
		os.Unsetenv("IS_TEST")
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "123456789012", nil
		}
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			roles: map[string]*iam.GetRoleOutput{
				"my-installer": roleOutput("my-installer", policyWithExternalID(externalID)),
			},
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, "arn:aws:s3:::my-bucket", "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		var w *CrossAccountWarning
		Expect(errors.As(err, &w)).To(BeFalse())
		Expect(err.Error()).To(ContainSubstring("expected IAM role resource"))
	})

	It("skips cross-account check when installer ARN is a non-role IAM resource in another account", func() {
		os.Unsetenv("IS_TEST")
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "123456789012", nil
		}
		iamUserARN := "arn:aws:iam::999999999999:user/my-user"
		supportInSameAccount := "arn:aws:iam::999999999999:role/my-support"
		newTrustPolicyLoader = stubLoader(&stubIAMClient{
			err: fmt.Errorf("should not be reached"),
		})

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, iamUserARN, supportInSameAccount, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		var w *CrossAccountWarning
		Expect(errors.As(err, &w)).To(BeFalse())
		Expect(err.Error()).To(ContainSubstring("expected IAM role resource"))
	})

	It("proceeds with validation when caller account matches role account", func() {
		os.Unsetenv("IS_TEST")
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "123456789012", nil
		}
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

	It("falls through to GetRole when callerAccountID fails", func() {
		os.Unsetenv("IS_TEST")
		callerAccountID = func(_ context.Context, _ aws.Config) (string, error) {
			return "", fmt.Errorf("STS unavailable")
		}
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

	It("returns error when loader construction fails", func() {
		os.Unsetenv("IS_TEST")
		newTrustPolicyLoader = stubLoaderError(fmt.Errorf("no credentials"))

		err := validateTrustPolicyExternalIDWithAWS(
			context.Background(), externalID, installerARN, supportARN, "us-east-1",
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no credentials"))
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
	It("constructs a loader from an AWS config", func() {
		cfg := aws.Config{Region: "us-east-1"}
		loader, err := newIAMTrustPolicyLoader(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(loader).NotTo(BeNil())
		Expect(loader.client).NotTo(BeNil())
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
