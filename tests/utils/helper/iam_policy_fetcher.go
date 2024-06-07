package helper

import (
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
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
	sess := session.Must(session.NewSession())
	svc := iam.New(sess)

	var policies []IAMPolicy

	// List attached policies
	attachedPolicies, err := listAttachedRolePolicies(svc, roleName)
	if err != nil {
		return nil, err
	}

	for _, policy := range attachedPolicies {
		policyDoc, err := getPolicyDocument(svc, *policy.PolicyArn)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policyDoc)
	}

	return policies, nil
}

func listAttachedRolePolicies(svc *iam.IAM, roleName string) ([]*iam.AttachedPolicy, error) {
	var attachedPolicies []*iam.AttachedPolicy
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	err := svc.ListAttachedRolePoliciesPages(input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)
		return !lastPage
	})

	return attachedPolicies, err
}

func getPolicyDocument(svc *iam.IAM, policyArn string) (IAMPolicy, error) {
	input := &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	}

	result, err := svc.GetPolicy(input)
	if err != nil {
		return IAMPolicy{}, err
	}

	policyVersionInput := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(policyArn),
		VersionId: result.Policy.DefaultVersionId,
	}

	versionResult, err := svc.GetPolicyVersion(policyVersionInput)
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
