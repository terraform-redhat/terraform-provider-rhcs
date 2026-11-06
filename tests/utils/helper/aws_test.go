// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type fakeEC2SecurityGroupClient struct {
	describeVpcsOutput *ec2.DescribeVpcsOutput
	describeVpcsErr    error

	securityGroups            []types.SecurityGroup
	describeSecurityGroupsErr error

	networkInterfacesByGroup   map[string][]types.NetworkInterface
	describeNetworkIfacesError error

	deleteErrorsByGroup map[string][]error
	deleteCallsByGroup  map[string]int
}

func (f *fakeEC2SecurityGroupClient) DescribeVpcs(_ context.Context, _ *ec2.DescribeVpcsInput, _ ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	if f.describeVpcsOutput == nil {
		return &ec2.DescribeVpcsOutput{}, f.describeVpcsErr
	}
	return f.describeVpcsOutput, f.describeVpcsErr
}

func (f *fakeEC2SecurityGroupClient) DescribeSecurityGroups(_ context.Context, _ *ec2.DescribeSecurityGroupsInput, _ ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	if f.describeSecurityGroupsErr != nil {
		return nil, f.describeSecurityGroupsErr
	}
	return &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: f.securityGroups,
	}, nil
}

func (f *fakeEC2SecurityGroupClient) DescribeNetworkInterfaces(_ context.Context, input *ec2.DescribeNetworkInterfacesInput, _ ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	if f.describeNetworkIfacesError != nil {
		return nil, f.describeNetworkIfacesError
	}
	groupID := filterValue(input.Filters, "group-id")
	return &ec2.DescribeNetworkInterfacesOutput{
		NetworkInterfaces: f.networkInterfacesByGroup[groupID],
	}, nil
}

func (f *fakeEC2SecurityGroupClient) DeleteSecurityGroup(_ context.Context, input *ec2.DeleteSecurityGroupInput, _ ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
	if f.deleteCallsByGroup == nil {
		f.deleteCallsByGroup = make(map[string]int)
	}

	groupID := aws.ToString(input.GroupId)
	call := f.deleteCallsByGroup[groupID]
	f.deleteCallsByGroup[groupID] = call + 1

	errorsByCall := f.deleteErrorsByGroup[groupID]
	if call < len(errorsByCall) && errorsByCall[call] != nil {
		return nil, errorsByCall[call]
	}

	return &ec2.DeleteSecurityGroupOutput{}, nil
}

func TestIsKnownLeakedSecurityGroup(t *testing.T) {
	const infraID = "2l26dls6dsbg0im4n971v4nlgorn2mku"
	ownedTag := fmt.Sprintf(clusterOwnedTagTemplate, infraID)

	tests := []struct {
		name  string
		group types.SecurityGroup
		want  bool
	}{
		{
			name: "vpce-private-router with managed tags",
			group: securityGroup(
				"sg-1",
				infraID+"-vpce-private-router",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
			want: true,
		},
		{
			name: "default-sg with owned cluster tag",
			group: securityGroup(
				"sg-2",
				infraID+"-default-sg",
				tag(ownedTag, "owned"),
			),
			want: true,
		},
		{
			name: "matching name but no ownership tags",
			group: securityGroup(
				"sg-3",
				infraID+"-vpce-private-router",
				tag("Name", infraID+"-vpce-private-router"),
			),
			want: false,
		},
		{
			name:  "default security group must be skipped",
			group: securityGroup("sg-4", "default"),
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isKnownLeakedSecurityGroup(tc.group)
			if got != tc.want {
				t.Fatalf("isKnownLeakedSecurityGroup() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDeleteExtraSecurityGroupsSkipsNonTestManagedVPC(t *testing.T) {
	const (
		vpcID   = "vpc-1"
		infraID = "2l26dls6dsbg0im4n971v4nlgorn2mku"
		groupID = "sg-1"
	)

	client := &fakeEC2SecurityGroupClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{{
				VpcId: aws.String(vpcID),
				Tags:  []types.Tag{tag("Name", "non-test-vpc")},
			}},
		},
		securityGroups: []types.SecurityGroup{
			securityGroup(
				groupID,
				infraID+"-vpce-private-router",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
		},
	}

	err := deleteExtraSecurityGroupsWithClient(context.Background(), client, vpcID, securityGroupCleanupOptions{
		maxAttempts:   1,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("deleteExtraSecurityGroupsWithClient() unexpected error: %v", err)
	}
	if client.deleteCallsByGroup[groupID] != 0 {
		t.Fatalf("expected no delete calls, got %d", client.deleteCallsByGroup[groupID])
	}
}

func TestDeleteExtraSecurityGroupsRetriesDependencyViolationAndSucceeds(t *testing.T) {
	const (
		vpcID   = "vpc-1"
		infraID = "2l26dls6dsbg0im4n971v4nlgorn2mku"
		groupID = "sg-1"
	)

	client := &fakeEC2SecurityGroupClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{{
				VpcId: aws.String(vpcID),
				Tags:  []types.Tag{tag(testManagedVPCTagKey, testManagedVPCTagValue)},
			}},
		},
		securityGroups: []types.SecurityGroup{
			securityGroup(
				groupID,
				infraID+"-vpce-private-router",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
		},
		deleteErrorsByGroup: map[string][]error{
			groupID: {
				&smithy.GenericAPIError{Code: "DependencyViolation", Message: "still referenced"},
				nil,
			},
		},
	}

	err := deleteExtraSecurityGroupsWithClient(context.Background(), client, vpcID, securityGroupCleanupOptions{
		maxAttempts:   3,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("deleteExtraSecurityGroupsWithClient() unexpected error: %v", err)
	}
	if client.deleteCallsByGroup[groupID] != 2 {
		t.Fatalf("expected 2 delete calls, got %d", client.deleteCallsByGroup[groupID])
	}
}

func TestDeleteExtraSecurityGroupsReturnsErrorWhenSecurityGroupStaysAttached(t *testing.T) {
	const (
		vpcID           = "vpc-1"
		infraID         = "2l26dls6dsbg0im4n971v4nlgorn2mku"
		groupID         = "sg-1"
		cleanupAttempts = 2
	)

	output := captureLogOutput(t)

	client := &fakeEC2SecurityGroupClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{{
				VpcId: aws.String(vpcID),
				Tags:  []types.Tag{tag(testManagedVPCTagKey, testManagedVPCTagValue)},
			}},
		},
		securityGroups: []types.SecurityGroup{
			securityGroup(
				groupID,
				infraID+"-default-sg",
				tag(clusterIDTagKey, infraID),
				tag(redHatManagedTagKey, redHatManagedTagValue),
			),
		},
		networkInterfacesByGroup: map[string][]types.NetworkInterface{
			groupID: {
				{
					NetworkInterfaceId: aws.String("eni-1"),
				},
				{
					NetworkInterfaceId: aws.String("eni-worker"),
					InterfaceType:      types.NetworkInterfaceTypeInterface,
					Status:             types.NetworkInterfaceStatusInUse,
					RequesterManaged:   aws.Bool(false),
					OwnerId:            aws.String("test-owner"),
					SubnetId:           aws.String("subnet-1"),
					Attachment: &types.NetworkInterfaceAttachment{
						InstanceId:      aws.String("i-worker"),
						InstanceOwnerId: aws.String("test-owner"),
						Status:          types.AttachmentStatusAttached,
					},
				},
				{
					NetworkInterfaceId: aws.String("eni-endpoint"),
					InterfaceType:      types.NetworkInterfaceTypeVpcEndpoint,
					Status:             types.NetworkInterfaceStatusInUse,
					Description:        aws.String("VPC Endpoint Interface vpce-1"),
					RequesterManaged:   aws.Bool(true),
					RequesterId:        aws.String("test-requester"),
				},
			},
		},
	}

	err := deleteExtraSecurityGroupsWithClient(context.Background(), client, vpcID, securityGroupCleanupOptions{
		maxAttempts:   cleanupAttempts,
		retryInterval: time.Millisecond,
		sleep:         func(time.Duration) {},
	})
	if err == nil {
		t.Fatal("expected error when security group stays attached, got nil")
	}
	if !strings.Contains(err.Error(), "still attached") {
		t.Fatalf("expected attachment error, got: %v", err)
	}
	if client.deleteCallsByGroup[groupID] != 0 {
		t.Fatalf("expected 0 delete calls, got %d", client.deleteCallsByGroup[groupID])
	}

	// Every attempt logs each blocking network interface with all diagnostic fields.
	// Missing optional metadata is logged as empty values and requester_managed=unknown.
	messagePrefix := fmt.Sprintf("Security group cleanup blocked by network interface: vpc_id=%q security_group_id=%q ", vpcID, groupID)
	expectedMessages := []string{
		messagePrefix + `eni_id="eni-1" interface_type="" status="" description="" ` +
			`requester_managed=unknown requester_id="" owner_id="" subnet_id="" ` +
			`instance_id="" instance_owner_id="" attachment_status=""`,
		messagePrefix + `eni_id="eni-worker" interface_type="interface" status="in-use" description="" ` +
			`requester_managed=false requester_id="" owner_id="test-owner" subnet_id="subnet-1" ` +
			`instance_id="i-worker" instance_owner_id="test-owner" attachment_status="attached"`,
		messagePrefix + `eni_id="eni-endpoint" interface_type="vpc_endpoint" status="in-use" ` +
			`description="VPC Endpoint Interface vpce-1" requester_managed=true requester_id="test-requester" ` +
			`owner_id="" subnet_id="" instance_id="" instance_owner_id="" attachment_status=""`,
	}
	logs := unescapedLogs(output)
	for _, message := range expectedMessages {
		if count := strings.Count(logs, message); count != cleanupAttempts {
			t.Errorf("expected message to be logged %d times, got %d: %s\nlogs:\n%s", cleanupAttempts, count, message, logs)
		}
	}
}

func TestIsSecurityGroupAttachedRedactsAccountIDsInLogs(t *testing.T) {
	const (
		ownerAccountID         = "111122223333"
		requesterAccountID     = "444455556666"
		instanceOwnerAccountID = "777788889999"
	)

	output := captureLogOutput(t)

	client := &fakeEC2SecurityGroupClient{
		networkInterfacesByGroup: map[string][]types.NetworkInterface{
			"sg-1": {
				{
					NetworkInterfaceId: aws.String("eni-account-ids"),
					OwnerId:            aws.String(ownerAccountID),
					RequesterId:        aws.String(requesterAccountID),
					SubnetId:           aws.String("subnet-0987654321098fedc"),
					Attachment: &types.NetworkInterfaceAttachment{
						InstanceId:      aws.String("i-0123456789012abcd"),
						InstanceOwnerId: aws.String(instanceOwnerAccountID),
					},
				},
				{
					NetworkInterfaceId: aws.String("eni-service-alias"),
					RequesterId:        aws.String("amazon-elb"),
					Attachment: &types.NetworkInterfaceAttachment{
						InstanceOwnerId: aws.String("amazon-aws"),
					},
				},
				{
					NetworkInterfaceId: aws.String("eni-no-metadata"),
				},
			},
		},
	}

	attached, err := isSecurityGroupAttached(context.Background(), client, "vpc-1", "sg-1")
	if err != nil {
		t.Fatalf("isSecurityGroupAttached() unexpected error: %v", err)
	}
	if !attached {
		t.Fatal("expected security group to be reported as attached")
	}

	logs := unescapedLogs(output)
	for _, accountID := range []string{ownerAccountID, requesterAccountID, instanceOwnerAccountID} {
		if strings.Contains(logs, accountID) {
			t.Fatalf("log output contains a raw account ID:\n%s", logs)
		}
	}

	redacted := fmt.Sprintf("%q", RedactValue)
	expectedFieldsByENI := map[string][]string{
		"eni-account-ids": {
			"owner_id=" + redacted,
			"requester_id=" + redacted,
			"instance_owner_id=" + redacted,
			// Resource IDs with 12-digit runs are not account IDs and stay visible.
			`subnet_id="subnet-0987654321098fedc"`,
			`instance_id="i-0123456789012abcd"`,
		},
		"eni-service-alias": {
			`owner_id=""`,
			`requester_id="amazon-elb"`,
			`instance_owner_id="amazon-aws"`,
		},
		"eni-no-metadata": {
			`owner_id=""`,
			`requester_id=""`,
			`instance_owner_id=""`,
			"requester_managed=unknown",
		},
	}
	for eniID, expectedFields := range expectedFieldsByENI {
		line := logLineContaining(logs, fmt.Sprintf("eni_id=%q", eniID))
		if line == "" {
			t.Fatalf("expected a log line for %s, got:\n%s", eniID, logs)
		}
		for _, field := range expectedFields {
			if !strings.Contains(line, " "+field) {
				t.Errorf("expected log line for %s to contain %s, got: %s", eniID, field, line)
			}
		}
	}
}

// nilNetworkInterfacesClient returns a nil DescribeNetworkInterfaces output so
// tests can exercise the defensive nil check in isSecurityGroupAttached.
type nilNetworkInterfacesClient struct {
	fakeEC2SecurityGroupClient
}

func (*nilNetworkInterfacesClient) DescribeNetworkInterfaces(_ context.Context, _ *ec2.DescribeNetworkInterfacesInput, _ ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return nil, nil
}

func TestIsSecurityGroupAttachedWithoutNetworkInterfacesLogsNothing(t *testing.T) {
	clients := map[string]ec2SecurityGroupAPI{
		"nil output":            &nilNetworkInterfacesClient{},
		"no network interfaces": &fakeEC2SecurityGroupClient{},
	}
	for name, client := range clients {
		t.Run(name, func(t *testing.T) {
			output := captureLogOutput(t)

			attached, err := isSecurityGroupAttached(context.Background(), client, "vpc-1", "sg-1")
			if err != nil {
				t.Fatalf("isSecurityGroupAttached() unexpected error: %v", err)
			}
			if attached {
				t.Fatal("expected security group to be reported as not attached")
			}
			if output.Len() != 0 {
				t.Fatalf("expected no log output, got:\n%s", output.String())
			}
		})
	}
}

// captureLogOutput redirects the shared Logger to a buffer for the rest of the
// test and restores the previous destination during cleanup. Because Logger is
// shared, tests that use it must not run in parallel.
func captureLogOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	output := &bytes.Buffer{}
	previousOutput := Logger.SetOutput(output)
	t.Cleanup(func() { Logger.SetOutput(previousOutput) })
	return output
}

// unescapedLogs returns the captured output with the quotes escaped by the
// text formatter restored, so assertions can match the original message text.
func unescapedLogs(output *bytes.Buffer) string {
	return strings.ReplaceAll(output.String(), `\"`, `"`)
}

// logLineContaining returns the first log line that contains substring, or an
// empty string when no line matches.
func logLineContaining(logs string, substring string) string {
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, substring) {
			return line
		}
	}
	return ""
}

func securityGroup(groupID string, groupName string, tags ...types.Tag) types.SecurityGroup {
	return types.SecurityGroup{
		GroupId:   aws.String(groupID),
		GroupName: aws.String(groupName),
		Tags:      tags,
	}
}

func tag(key string, value string) types.Tag {
	return types.Tag{
		Key:   aws.String(key),
		Value: aws.String(value),
	}
}

func filterValue(filters []types.Filter, name string) string {
	for _, filter := range filters {
		if aws.ToString(filter.Name) == name && len(filter.Values) > 0 {
			return filter.Values[0]
		}
	}
	return ""
}
