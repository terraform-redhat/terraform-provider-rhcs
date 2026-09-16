// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/openshift-online/ocm-common/pkg/aws/ststrust"
)

// TrustPolicyValidatorFunc validates trust_policy_external_id against installer and support IAM role trust policies.
// Tests may replace TrustPolicyValidator to avoid AWS calls.
type TrustPolicyValidatorFunc func(context.Context, string, string, string, string) error

// TrustPolicyValidator validates trust_policy_external_id against installer and support IAM role trust policies.
var TrustPolicyValidator TrustPolicyValidatorFunc = validateTrustPolicyExternalIDWithAWS

var errMismatchedSTSExternalIDTrustPolicies = errors.New(
	"installer and support role trust policies define STS external IDs with no value in common; " +
		"set sts.trust_policy_external_id explicitly",
)

var errAmbiguousSTSExternalIDTrustPolicies = errors.New(
	"could not determine a single STS external ID from the installer and support role trust policies; " +
		"set sts.trust_policy_external_id explicitly",
)

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

// CrossAccountWarning indicates that trust policy validation was skipped because the caller's
// AWS account differs from the account that owns the IAM roles.
type CrossAccountWarning struct {
	CallerAccount string
	RoleAccount   string
}

func (w *CrossAccountWarning) Error() string {
	return fmt.Sprintf(
		"cannot pre-validate trust_policy_external_id: IAM roles are in account %s "+
			"but credentials belong to account %s (cross-account); "+
			"if cluster creation fails, verify the external ID matches your role trust policies",
		w.RoleAccount, w.CallerAccount,
	)
}

// callerAccountIDFunc resolves the AWS account ID of the current caller. Tests may replace this.
type callerAccountIDFunc func(ctx context.Context, cfg aws.Config) (string, error)

var callerAccountID callerAccountIDFunc = getCallerAccountFromSTS

func getCallerAccountFromSTS(ctx context.Context, cfg aws.Config) (string, error) {
	client := sts.NewFromConfig(cfg)
	output, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return aws.ToString(output.Account), nil
}

// loadAWSConfig loads the default AWS configuration for the given region.
func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}
	return awsconfig.LoadDefaultConfig(ctx, opts...)
}

// ValidateTrustPolicyExternalID validates sts.trust_policy_external_id against installer and support IAM roles.
// When entered is set, it must appear in both trust policies. When unset, create fails if a single external ID
// can be discovered from IAM (explicit config required) or if IAM external IDs are ambiguous.
func ValidateTrustPolicyExternalID(
	ctx context.Context,
	entered, installerRoleARN, supportRoleARN, region string,
) error {
	if entered != "" {
		if err := ststrust.ValidateSTSExternalIDFormat(entered); err != nil {
			return err
		}
		if installerRoleARN == "" || supportRoleARN == "" {
			return fmt.Errorf(
				"installer and support role ARNs are required in sts when trust_policy_external_id is set",
			)
		}
		return TrustPolicyValidator(ctx, entered, installerRoleARN, supportRoleARN, region)
	}
	if installerRoleARN == "" || supportRoleARN == "" {
		return nil
	}
	return TrustPolicyValidator(ctx, "", installerRoleARN, supportRoleARN, region)
}

// isAWSAccountID reports whether s is a 12-digit AWS account ID.
func isAWSAccountID(s string) bool {
	if len(s) != 12 {
		return false
	}
	for i := 0; i < 12; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// isIAMRoleNameChar reports whether c is valid in an IAM role name ([\w+=,.@-]).
func isIAMRoleNameChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
		c == '_' || c == '+' || c == '=' || c == ',' || c == '.' || c == '@' || c == '-'
}

// parseIAMRoleARN parses an ARN and returns it only when it identifies an IAM role.
// Works with all AWS partitions (aws, aws-us-gov, aws-cn, etc.).
func parseIAMRoleARN(rawARN string) (awsarn.ARN, bool) {
	parsed, err := awsarn.Parse(rawARN)
	if err != nil || parsed.Service != "iam" || parsed.Region != "" || !isAWSAccountID(parsed.AccountID) {
		return awsarn.ARN{}, false
	}
	if !strings.HasPrefix(parsed.Resource, "role/") || strings.Contains(parsed.Resource, "//") {
		return awsarn.ARN{}, false
	}
	name := parsed.Resource[strings.LastIndex(parsed.Resource, "/")+1:]
	if name == "" {
		return awsarn.ARN{}, false
	}
	for i := 0; i < len(name); i++ {
		if !isIAMRoleNameChar(name[i]) {
			return awsarn.ARN{}, false
		}
	}
	return parsed, true
}

// validateTrustPolicyExternalIDWithAWS loads installer and support trust policies from IAM and validates
// the entered value.
func validateTrustPolicyExternalIDWithAWS(
	ctx context.Context,
	entered, installerRoleARN, supportRoleARN, region string,
) error {
	if os.Getenv("IS_TEST") == "true" {
		return nil
	}

	cfg, err := loadAWSConfig(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration for trust policy validation: %w", err)
	}

	installerARNParsed, installerIsRole := parseIAMRoleARN(installerRoleARN)
	supportARNParsed, supportIsRole := parseIAMRoleARN(supportRoleARN)
	if installerIsRole && supportIsRole {
		if installerARNParsed.AccountID != supportARNParsed.AccountID {
			return fmt.Errorf(
				"installer role (account %s) and support role (account %s) must be in the same AWS account",
				installerARNParsed.AccountID, supportARNParsed.AccountID,
			)
		}
		account, stsErr := callerAccountID(ctx, cfg)
		if stsErr == nil && account != "" && account != installerARNParsed.AccountID {
			return &CrossAccountWarning{
				CallerAccount: account,
				RoleAccount:   installerARNParsed.AccountID,
			}
		}
	}

	loader, err := newTrustPolicyLoader(cfg)
	if err != nil {
		return fmt.Errorf("failed to create IAM client for trust policy validation: %w", err)
	}
	installerPolicy, err := loader.trustPolicyJSONForRoleARN(ctx, installerRoleARN)
	if err != nil {
		return fmt.Errorf("failed to read installer role trust policy: %w", err)
	}
	supportPolicy, err := loader.trustPolicyJSONForRoleARN(ctx, supportRoleARN)
	if err != nil {
		return fmt.Errorf("failed to read support role trust policy: %w", err)
	}
	if entered != "" {
		return ststrust.ValidateEnteredForRoleTrustPolicies(entered, installerPolicy, supportPolicy)
	}
	return validateRequiredTrustPolicyExternalIDUnset(installerPolicy, supportPolicy)
}

// validateRequiredTrustPolicyExternalIDUnset enforces explicit config when IAM requires an external ID
// or when discovery is ambiguous.
func validateRequiredTrustPolicyExternalIDUnset(installerPolicy, supportPolicy string) error {
	discovered, err := ststrust.DiscoverSTSExternalID(installerPolicy, supportPolicy)
	if err != nil {
		return err
	}
	if discovered != "" {
		return &trustPolicyExternalIDRequiredError{Discovered: discovered}
	}

	installerIDs, err := ststrust.CollectSTSExternalIDsFromTrustPolicy(installerPolicy)
	if err != nil {
		return err
	}
	supportIDs, err := ststrust.CollectSTSExternalIDsFromTrustPolicy(supportPolicy)
	if err != nil {
		return err
	}
	if len(installerIDs) == 0 && len(supportIDs) == 0 {
		return nil
	}
	if hasMismatchedSTSExternalIDTrustPolicies(installerIDs, supportIDs) {
		return errMismatchedSTSExternalIDTrustPolicies
	}
	if isSTSExternalIDDiscoveryAmbiguous(discovered, installerIDs, supportIDs) {
		return errAmbiguousSTSExternalIDTrustPolicies
	}
	return nil
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

// isSTSExternalIDDiscoveryAmbiguous reports whether trust policies contain external IDs that cannot be
// resolved to one value, using precomputed discovery and ID lists.
func isSTSExternalIDDiscoveryAmbiguous(discovered string, installerIDs, supportIDs []string) bool {
	if len(installerIDs) == 0 && len(supportIDs) == 0 {
		return false
	}
	return discovered == ""
}

// iamRoleGetter abstracts the IAM GetRole API so tests can inject a stub.
type iamRoleGetter interface {
	GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error)
}

// iamTrustPolicyLoader reads IAM role trust policies.
type iamTrustPolicyLoader struct {
	client iamRoleGetter
}

// newTrustPolicyLoader constructs an iamTrustPolicyLoader. Tests may replace this to inject a stub.
var newTrustPolicyLoader = newIAMTrustPolicyLoader

// newIAMTrustPolicyLoader constructs an IAM trust policy loader from an existing AWS config.
func newIAMTrustPolicyLoader(cfg aws.Config) (*iamTrustPolicyLoader, error) {
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
	if output.Role == nil {
		return "", fmt.Errorf("GetRole returned no role for %q", roleName)
	}
	return trustPolicyJSONFromRole(*output.Role)
}

// roleNameFromARN extracts the IAM role name from a role ARN for use with GetRole.
// IAM role names are unique per account; GetRole expects the name without the path prefix.
func roleNameFromARN(roleARN string) (string, error) {
	parsed, ok := parseIAMRoleARN(roleARN)
	if !ok {
		return "", fmt.Errorf("invalid role ARN %q: expected IAM role resource", roleARN)
	}
	const rolePrefix = "role/"
	roleResource := parsed.Resource[len(rolePrefix):]
	if idx := strings.LastIndex(roleResource, "/"); idx >= 0 {
		roleResource = roleResource[idx+1:]
	}
	return roleResource, nil
}

// trustPolicyJSONFromRole decodes the assume-role policy document attached to an IAM role.
func trustPolicyJSONFromRole(role iamtypes.Role) (string, error) {
	if role.AssumeRolePolicyDocument == nil {
		return "", fmt.Errorf("role %q has no assume role policy document", aws.ToString(role.RoleName))
	}
	decoded, err := url.PathUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		return "", fmt.Errorf("failed to decode trust policy for role %q: %w", aws.ToString(role.RoleName), err)
	}
	return decoded, nil
}
