package sts

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/openshift-online/ocm-common/pkg/aws/ststrust"
)

// TrustPolicyValidatorFunc validates trust_policy_external_id against installer and support IAM role trust policies.
// Tests may replace TrustPolicyValidator to avoid AWS calls. A non-nil warning allows create to proceed.
type TrustPolicyValidatorFunc func(context.Context, string, string, string, string) (warning error, err error)

// TrustPolicyValidator validates trust_policy_external_id against installer and support IAM role trust policies.
var TrustPolicyValidator TrustPolicyValidatorFunc = validateTrustPolicyExternalIDWithAWS

var errMismatchedSTSExternalIDTrustPolicies = errors.New(
	"installer and support role trust policies define STS external IDs with no value in common; " +
		"set sts.trust_policy_external_id explicitly",
)

// ambiguousSTSExternalIDTrustPoliciesWarning returns a warning when IAM trust policies define
// external IDs but discovery cannot select a single value.
func ambiguousSTSExternalIDTrustPoliciesWarning() error {
	return errors.New(
		"Could not determine a single STS external ID from the installer and support role trust policies. " +
			"The cluster will be created without an external ID unless you provide one with sts.trust_policy_external_id.",
	)
}

// trustPolicyExternalIDRequiredError reports that IAM roles require an explicit external ID in Terraform config.
type trustPolicyExternalIDRequiredError struct {
	Discovered string
}

// Error implements the error interface.
func (e *trustPolicyExternalIDRequiredError) Error() string {
	return fmt.Sprintf(
		"installer and support roles define STS external ID %q; set sts.trust_policy_external_id = %q",
		e.Discovered,
		e.Discovered,
	)
}

// ValidateTrustPolicyExternalID validates sts.trust_policy_external_id against installer and support IAM roles.
// When entered is set, it must appear in both trust policies. When unset, create fails if a single external ID
// can be discovered from IAM (explicit config required). Ambiguous IAM external IDs produce a warning and create
// proceeds without an external ID, matching ROSA CLI non-interactive behavior.
func ValidateTrustPolicyExternalID(
	ctx context.Context,
	entered, installerRoleARN, supportRoleARN, region string,
) (warning error, err error) {
	if entered != "" {
		if err := ststrust.ValidateSTSExternalIDFormat(entered); err != nil {
			return nil, err
		}
		if installerRoleARN == "" || supportRoleARN == "" {
			return nil, fmt.Errorf(
				"installer and support role ARNs are required in sts when trust_policy_external_id is set",
			)
		}
		return TrustPolicyValidator(ctx, entered, installerRoleARN, supportRoleARN, region)
	}
	if installerRoleARN == "" || supportRoleARN == "" {
		return nil, nil
	}
	return TrustPolicyValidator(ctx, "", installerRoleARN, supportRoleARN, region)
}

// validateTrustPolicyExternalIDWithAWS loads installer and support trust policies from IAM and validates the entered value.
func validateTrustPolicyExternalIDWithAWS(
	ctx context.Context,
	entered, installerRoleARN, supportRoleARN, region string,
) (warning error, err error) {
	if os.Getenv("IS_TEST") == "true" {
		return nil, nil
	}
	loader, err := newIAMTrustPolicyLoader(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration for trust policy validation: %w", err)
	}
	installerPolicy, err := loader.trustPolicyJSONForRoleARN(ctx, installerRoleARN)
	if err != nil {
		return nil, fmt.Errorf("failed to read installer role trust policy: %w", err)
	}
	supportPolicy, err := loader.trustPolicyJSONForRoleARN(ctx, supportRoleARN)
	if err != nil {
		return nil, fmt.Errorf("failed to read support role trust policy: %w", err)
	}
	if entered != "" {
		return nil, ststrust.ValidateEnteredForRoleTrustPolicies(entered, installerPolicy, supportPolicy)
	}
	return validateRequiredTrustPolicyExternalIDUnset(installerPolicy, supportPolicy)
}

// validateRequiredTrustPolicyExternalIDUnset enforces explicit config when IAM requires an external ID, or warns when discovery is ambiguous.
func validateRequiredTrustPolicyExternalIDUnset(installerPolicy, supportPolicy string) (warning error, err error) {
	discovered, err := ststrust.DiscoverSTSExternalID(installerPolicy, supportPolicy)
	if err != nil {
		return nil, err
	}
	if discovered != "" {
		return nil, &trustPolicyExternalIDRequiredError{Discovered: discovered}
	}

	installerIDs, err := ststrust.CollectSTSExternalIDsFromTrustPolicy(installerPolicy)
	if err != nil {
		return nil, err
	}
	supportIDs, err := ststrust.CollectSTSExternalIDsFromTrustPolicy(supportPolicy)
	if err != nil {
		return nil, err
	}
	if len(installerIDs) == 0 && len(supportIDs) == 0 {
		return nil, nil
	}
	if hasMismatchedSTSExternalIDTrustPolicies(installerIDs, supportIDs) {
		return nil, errMismatchedSTSExternalIDTrustPolicies
	}
	ambiguous, err := isSTSExternalIDDiscoveryAmbiguous(installerPolicy, supportPolicy)
	if err != nil {
		return nil, err
	}
	if ambiguous {
		return ambiguousSTSExternalIDTrustPoliciesWarning(), nil
	}
	return nil, nil
}

// hasMismatchedSTSExternalIDTrustPolicies reports when installer and support define external IDs with no overlap.
func hasMismatchedSTSExternalIDTrustPolicies(installerIDs, supportIDs []string) bool {
	if len(installerIDs) == 0 || len(supportIDs) == 0 {
		return false
	}
	lookup := make(map[string]struct{}, len(installerIDs))
	for _, id := range installerIDs {
		lookup[id] = struct{}{}
	}
	for _, id := range supportIDs {
		if _, ok := lookup[id]; ok {
			return false
		}
	}
	return true
}

// isSTSExternalIDDiscoveryAmbiguous reports whether trust policies contain external IDs that cannot be resolved to one value.
func isSTSExternalIDDiscoveryAmbiguous(installerPolicy, supportPolicy string) (bool, error) {
	installerIDs, err := ststrust.CollectSTSExternalIDsFromTrustPolicy(installerPolicy)
	if err != nil {
		return false, err
	}
	supportIDs, err := ststrust.CollectSTSExternalIDsFromTrustPolicy(supportPolicy)
	if err != nil {
		return false, err
	}
	if len(installerIDs) == 0 && len(supportIDs) == 0 {
		return false, nil
	}
	discovered, err := ststrust.DiscoverSTSExternalID(installerPolicy, supportPolicy)
	if err != nil {
		return false, err
	}
	return discovered == "", nil
}

// iamTrustPolicyLoader reads IAM role trust policies using the AWS SDK default credential chain.
type iamTrustPolicyLoader struct {
	client *iam.Client
}

// newIAMTrustPolicyLoader constructs an IAM trust policy loader for the given AWS region.
func newIAMTrustPolicyLoader(ctx context.Context, region string) (*iamTrustPolicyLoader, error) {
	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &iamTrustPolicyLoader{client: iam.NewFromConfig(cfg)}, nil
}

// trustPolicyJSONForRoleARN returns the decoded assume-role policy document for the given role ARN.
func (l *iamTrustPolicyLoader) trustPolicyJSONForRoleARN(ctx context.Context, roleARN string) (string, error) {
	if roleARN == "" {
		return "", nil
	}
	roleName, err := roleNameFromARN(roleARN)
	if err != nil {
		return "", err
	}
	output, err := l.client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return "", err
	}
	return trustPolicyJSONFromRole(*output.Role)
}

// roleNameFromARN extracts the IAM role name, including path prefix, from a role ARN.
func roleNameFromARN(roleARN string) (string, error) {
	parsed, err := awsarn.Parse(roleARN)
	if err != nil {
		return "", fmt.Errorf("invalid role ARN %q: %w", roleARN, err)
	}
	const rolePrefix = "role/"
	if !strings.HasPrefix(parsed.Resource, rolePrefix) {
		return "", fmt.Errorf("invalid role ARN %q: expected IAM role resource", roleARN)
	}
	return parsed.Resource[len(rolePrefix):], nil
}

// trustPolicyJSONFromRole decodes the assume-role policy document attached to an IAM role.
func trustPolicyJSONFromRole(role iamtypes.Role) (string, error) {
	if role.AssumeRolePolicyDocument == nil {
		return "", fmt.Errorf("role %q has no assume role policy document", aws.ToString(role.RoleName))
	}
	decoded, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		return "", fmt.Errorf("failed to decode trust policy for role %q: %w", aws.ToString(role.RoleName), err)
	}
	return decoded, nil
}
