package helper

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type IAMPolicy struct {
	Version   string
	Statement []IAMPolicyStatement
}

type IAMPolicyStatement struct {
	Effect   string
	Action   interface{}
	Resource interface{}
}

func GetRoleAttachedPolicies(roleName string) ([]IAMPolicy, error) {
	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	svc := iam.NewFromConfig(cfg)

	var policies []IAMPolicy

	// List attached policies
	attachedPolicies, err := listAttachedRolePolicies(ctx, svc, roleName)
	if err != nil {
		return nil, err
	}

	for _, policy := range attachedPolicies {
		policyDoc, err := getPolicyDocument(ctx, svc, aws.ToString(policy.PolicyArn))
		if err != nil {
			return nil, err
		}
		policies = append(policies, policyDoc)
	}

	return policies, nil
}

func listAttachedRolePolicies(ctx context.Context, svc *iam.Client, roleName string) ([]types.AttachedPolicy, error) {
	var attachedPolicies []types.AttachedPolicy
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	paginator := iam.NewListAttachedRolePoliciesPaginator(svc, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)
	}

	return attachedPolicies, nil
}

func getPolicyDocument(ctx context.Context, svc *iam.Client, policyArn string) (IAMPolicy, error) {
	input := &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	}

	result, err := svc.GetPolicy(ctx, input)
	if err != nil {
		return IAMPolicy{}, err
	}

	policyVersionInput := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(policyArn),
		VersionId: result.Policy.DefaultVersionId,
	}

	versionResult, err := svc.GetPolicyVersion(ctx, policyVersionInput)
	if err != nil {
		return IAMPolicy{}, err
	}

	var policy IAMPolicy
	policyDocument := *versionResult.PolicyVersion.Document
	policyDocumentDecoded, err := url.QueryUnescape(policyDocument)
	if err != nil {
		return IAMPolicy{}, err
	}

	err = json.Unmarshal([]byte(policyDocumentDecoded), &policy)
	return policy, err
}
