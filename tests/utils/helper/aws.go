package helper

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// TODO: Remove this once this bug is fixed (https://issues.redhat.com/browse/OCPBUGS-74960)
func DeleteExtraSecurityGroups(region string, vpcId string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return err
	}
	client := ec2.New(sess)

	var securityGroups []*ec2.SecurityGroup
	params := ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{&vpcId},
			},
		},
	}
	err = client.DescribeSecurityGroupsPages(&params, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		securityGroups = append(securityGroups, page.SecurityGroups...)
		return !lastPage
	})
	namePattern, err := regexp.Compile(`^\w{32}-vpce-private-router$`)
	if err != nil {
		return err
	}
	for _, securityGroup := range securityGroups {
		if namePattern.MatchString(*securityGroup.GroupName) {
			// Found a matching extra security group
			out, err := client.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
				GroupId: securityGroup.GroupId,
			})
			if err != nil {
				return fmt.Errorf("Failed to delete security group '%s': %s", *securityGroup.GroupId, out.String())
			}
		}
	}

	return nil
}
