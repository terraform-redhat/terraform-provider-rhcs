// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

// Package seaplaneoperatorroles implements the rhcs_seaplane_operator_roles
// resource and data source, which manage the IAM operator roles (and worker
// instance role) required by ROSA Seaplane V2 (HCP) clusters in the customer
// AWS account.
//
// Role definitions and trust-policy templates come from the shared
// rosa-seaplane-api iam/ module (the single source of truth for V2); this
// package never hard-codes role suffixes, service accounts, policy ARNs, or
// trust-policy JSON.
package seaplaneoperatorroles

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AttachedPolicy is an AWS-managed policy attached to an IAM role.
type AttachedPolicy struct {
	ARN  string
	Name string
}

// InstanceProfileSummary describes an instance profile attached to a role.
type InstanceProfileSummary struct {
	Name string
	ARN  string
}

// RoleSummary is the observed state of an IAM role, as returned by the AWS
// API. Tags is the raw role tag set.
type RoleSummary struct {
	Name             string
	ARN              string
	TrustPolicy      string
	Tags             map[string]string
	AttachedPolicies []AttachedPolicy
	InstanceProfiles []InstanceProfileSummary
	PermissionsBound string
}

// IAMRoleClient is the subset of the AWS IAM/STS API used by this package.
// It is an interface so unit tests can run without AWS.
type IAMRoleClient interface {
	// CallerIdentity returns the account ID, partition, and ARN of the caller.
	CallerIdentity(ctx context.Context) (accountID, partition, callerARN string, err error)
	// GetRole returns the role summary, or (nil, nil) when the role does not exist.
	GetRole(ctx context.Context, name string) (*RoleSummary, error)
	// CreateRole creates the role with the given trust policy, tags, and
	// optional permissions boundary ARN.
	CreateRole(ctx context.Context, name, trustPolicy, permissionsBoundary string, tags map[string]string) (string, error)
	// UpdateTrustPolicy replaces the role's assume-role policy document.
	UpdateTrustPolicy(ctx context.Context, name, trustPolicy string) error
	// UpdatePermissionsBoundary sets (or clears, when empty) the role's
	// permissions boundary.
	UpdatePermissionsBoundary(ctx context.Context, name, permissionsBoundary string) error
	// TagRole merges the given tags into the role's tag set.
	TagRole(ctx context.Context, name string, tags map[string]string) error
	// AttachPolicy attaches an AWS-managed policy to the role. Attaching an
	// already-attached policy must be a no-op.
	AttachPolicy(ctx context.Context, roleName, policyARN string) error
	// DetachPolicy detaches an AWS-managed policy from the role. Detaching a
	// policy that is not attached must be a no-op.
	DetachPolicy(ctx context.Context, roleName, policyARN string) error
	// EnsureInstanceProfile creates the instance profile (if needed) and adds
	// the role to it.
	EnsureInstanceProfile(ctx context.Context, roleName, instanceProfileName string, tags map[string]string) error
	// DeleteInstanceProfile removes the role from the instance profile and
	// deletes the profile.
	DeleteInstanceProfile(ctx context.Context, roleName, instanceProfileName string) error
	// DeleteRole deletes the role. Deleting a missing role must be a no-op.
	DeleteRole(ctx context.Context, name string) error
	// ListRolesByPrefix returns all IAM roles whose names start with the given
	// prefix ("<prefix>-"), with their tags.
	ListRolesByPrefix(ctx context.Context, prefix string) ([]RoleSummary, error)
}

// awsIAMClient is the production IAMRoleClient, backed by aws-sdk-go-v2.
type awsIAMClient struct {
	iam *iam.Client
	sts *sts.Client
}

// NewAWSIAMClient loads the default AWS configuration (credentials from the
// standard AWS credential chain, region from AWS_REGION/AWS_DEFAULT_REGION)
// and returns an IAMRoleClient. The IAM API is global, so no specific region
// is required.
func NewAWSIAMClient(ctx context.Context) (IAMRoleClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS configuration: %w", err)
	}
	return &awsIAMClient{
		iam: iam.NewFromConfig(cfg),
		sts: sts.NewFromConfig(cfg),
	}, nil
}

func (c *awsIAMClient) CallerIdentity(ctx context.Context) (string, string, string, error) {
	output, err := c.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", "", fmt.Errorf("getting AWS caller identity: %w", err)
	}
	callerARN := aws.ToString(output.Arn)
	partition := "aws"
	if parsed, err := awsarn.Parse(callerARN); err == nil && parsed.Partition != "" {
		partition = parsed.Partition
	}
	return aws.ToString(output.Account), partition, callerARN, nil
}

// isNotFound reports whether the error is an IAM NoSuchEntity error (the
// requested role/policy/profile does not exist).
func isNotFound(err error) bool {
	var nf *iamtypes.NoSuchEntityException
	return errors.As(err, &nf)
}

func (c *awsIAMClient) GetRole(ctx context.Context, name string) (*RoleSummary, error) {
	output, err := c.iam.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String(name)})
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting IAM role %q: %w", name, err)
	}
	summary := &RoleSummary{
		Name:        name,
		ARN:         aws.ToString(output.Role.Arn),
		TrustPolicy: aws.ToString(output.Role.AssumeRolePolicyDocument),
		Tags:        tagsToMap(output.Role.Tags),
	}
	if output.Role.PermissionsBoundary != nil {
		summary.PermissionsBound = aws.ToString(output.Role.PermissionsBoundary.PermissionsBoundaryArn)
	}
	if err := c.fillRoleDetails(ctx, name, summary); err != nil {
		return nil, err
	}
	return summary, nil
}

// fillRoleDetails populates attached policies and instance profiles.
func (c *awsIAMClient) fillRoleDetails(ctx context.Context, name string, summary *RoleSummary) error {
	policies, err := ListAttachedRolePolicies(ctx, c.iam, name)
	if err != nil {
		return err
	}
	summary.AttachedPolicies = policies

	profilePages := iam.NewListInstanceProfilesForRolePaginator(c.iam,
		&iam.ListInstanceProfilesForRoleInput{RoleName: aws.String(name)})
	for profilePages.HasMorePages() {
		page, err := profilePages.NextPage(ctx)
		if err != nil {
			if isNotFound(err) {
				return nil
			}
			return fmt.Errorf("listing instance profiles for IAM role %q: %w", name, err)
		}
		for _, profile := range page.InstanceProfiles {
			summary.InstanceProfiles = append(summary.InstanceProfiles, InstanceProfileSummary{
				Name: aws.ToString(profile.InstanceProfileName),
				ARN:  aws.ToString(profile.Arn),
			})
		}
	}
	return nil
}

// ListAttachedRolePolicies returns the AWS-managed policies attached to a role.
func ListAttachedRolePolicies(ctx context.Context, client *iam.Client, name string) ([]AttachedPolicy, error) {
	var result []AttachedPolicy
	paginator := iam.NewListAttachedRolePoliciesPaginator(client,
		&iam.ListAttachedRolePoliciesInput{RoleName: aws.String(name)})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if isNotFound(err) {
				return nil, nil
			}
			return nil, fmt.Errorf("listing attached policies for IAM role %q: %w", name, err)
		}
		for _, policy := range page.AttachedPolicies {
			result = append(result, AttachedPolicy{
				ARN:  aws.ToString(policy.PolicyArn),
				Name: aws.ToString(policy.PolicyName),
			})
		}
	}
	return result, nil
}

func (c *awsIAMClient) CreateRole(ctx context.Context, name, trustPolicy, permissionsBoundary string,
	tags map[string]string) (string, error) {
	input := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	}
	if permissionsBoundary != "" {
		input.PermissionsBoundary = aws.String(permissionsBoundary)
	}
	for key, value := range tags {
		input.Tags = append(input.Tags, iamtypes.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
	output, err := c.iam.CreateRole(ctx, input)
	if err != nil {
		return "", fmt.Errorf("creating IAM role %q: %w", name, err)
	}
	return aws.ToString(output.Role.Arn), nil
}

func (c *awsIAMClient) UpdateTrustPolicy(ctx context.Context, name, trustPolicy string) error {
	_, err := c.iam.UpdateAssumeRolePolicy(ctx, &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(name),
		PolicyDocument: aws.String(trustPolicy),
	})
	if err != nil {
		return fmt.Errorf("updating trust policy of IAM role %q: %w", name, err)
	}
	return nil
}

func (c *awsIAMClient) UpdatePermissionsBoundary(ctx context.Context, name, permissionsBoundary string) error {
	if permissionsBoundary == "" {
		_, err := c.iam.DeleteRolePermissionsBoundary(ctx, &iam.DeleteRolePermissionsBoundaryInput{
			RoleName: aws.String(name),
		})
		if err != nil {
			return fmt.Errorf("removing permissions boundary of IAM role %q: %w", name, err)
		}
		return nil
	}
	_, err := c.iam.PutRolePermissionsBoundary(ctx, &iam.PutRolePermissionsBoundaryInput{
		RoleName:            aws.String(name),
		PermissionsBoundary: aws.String(permissionsBoundary),
	})
	if err != nil {
		return fmt.Errorf("updating permissions boundary of IAM role %q: %w", name, err)
	}
	return nil
}

func (c *awsIAMClient) TagRole(ctx context.Context, name string, tags map[string]string) error {
	input := &iam.TagRoleInput{RoleName: aws.String(name)}
	for key, value := range tags {
		input.Tags = append(input.Tags, iamtypes.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
	_, err := c.iam.TagRole(ctx, input)
	if err != nil {
		return fmt.Errorf("tagging IAM role %q: %w", name, err)
	}
	return nil
}

func (c *awsIAMClient) AttachPolicy(ctx context.Context, roleName, policyARN string) error {
	_, err := c.iam.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyARN),
	})
	if err != nil {
		return fmt.Errorf("attaching policy %q to IAM role %q: %w", policyARN, roleName, err)
	}
	return nil
}

func (c *awsIAMClient) DetachPolicy(ctx context.Context, roleName, policyARN string) error {
	_, err := c.iam.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyARN),
	})
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return fmt.Errorf("detaching policy %q from IAM role %q: %w", policyARN, roleName, err)
	}
	return nil
}

func (c *awsIAMClient) EnsureInstanceProfile(ctx context.Context, roleName, instanceProfileName string,
	tags map[string]string) error {
	createInput := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	}
	for key, value := range tags {
		createInput.Tags = append(createInput.Tags, iamtypes.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
	_, err := c.iam.CreateInstanceProfile(ctx, createInput)
	if err != nil {
		var exists *iamtypes.EntityAlreadyExistsException
		if !errors.As(err, &exists) {
			return fmt.Errorf("creating instance profile %q: %w", instanceProfileName, err)
		}
	}
	_, err = c.iam.AddRoleToInstanceProfile(ctx, &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
		RoleName:            aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("adding IAM role %q to instance profile %q: %w", roleName, instanceProfileName, err)
	}
	return nil
}

func (c *awsIAMClient) DeleteInstanceProfile(ctx context.Context, roleName, instanceProfileName string) error {
	_, err := c.iam.RemoveRoleFromInstanceProfile(ctx, &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
		RoleName:            aws.String(roleName),
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("removing IAM role %q from instance profile %q: %w", roleName, instanceProfileName, err)
	}
	_, err = c.iam.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting instance profile %q: %w", instanceProfileName, err)
	}
	return nil
}

func (c *awsIAMClient) DeleteRole(ctx context.Context, name string) error {
	_, err := c.iam.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: aws.String(name)})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting IAM role %q: %w", name, err)
	}
	return nil
}

func (c *awsIAMClient) ListRolesByPrefix(ctx context.Context, prefix string) ([]RoleSummary, error) {
	var result []RoleSummary
	namePrefix := prefix + "-"
	paginator := iam.NewListRolesPaginator(c.iam, &iam.ListRolesInput{PathPrefix: aws.String("/")})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing IAM roles: %w", err)
		}
		for _, role := range page.Roles {
			name := aws.ToString(role.RoleName)
			if !strings.HasPrefix(name, namePrefix) {
				continue
			}
			summary := &RoleSummary{
				Name:        name,
				ARN:         aws.ToString(role.Arn),
				TrustPolicy: aws.ToString(role.AssumeRolePolicyDocument),
				Tags:        tagsToMap(role.Tags),
			}
			if err := c.fillRoleDetails(ctx, name, summary); err != nil {
				return nil, err
			}
			result = append(result, *summary)
		}
	}
	return result, nil
}

func tagsToMap(tags []iamtypes.Tag) map[string]string {
	result := make(map[string]string, len(tags))
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}
