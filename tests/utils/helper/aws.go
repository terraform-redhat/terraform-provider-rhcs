package helper

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// TODO: Remove this once this bug is fixed (https://issues.redhat.com/browse/OCPBUGS-74960)
func DeleteExtraSecurityGroups(region string, vpcId string) error {
	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}
	client := ec2.NewFromConfig(cfg)

	var securityGroups []types.SecurityGroup
	paginator := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		securityGroups = append(securityGroups, page.SecurityGroups...)
	}

	namePattern, err := regexp.Compile(`^\w{32}-vpce-private-router$`)
	if err != nil {
		return err
	}
	for _, securityGroup := range securityGroups {
		if namePattern.MatchString(aws.ToString(securityGroup.GroupName)) {
			// Found a matching extra security group
			_, err := client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
				GroupId: securityGroup.GroupId,
			})
			if err != nil {
				return fmt.Errorf("failed to delete security group '%s': %w", aws.ToString(securityGroup.GroupId), err)
			}
		}
	}

	return nil
}
