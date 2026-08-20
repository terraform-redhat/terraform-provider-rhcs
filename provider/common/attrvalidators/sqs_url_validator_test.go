// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package attrvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SQS URL validator", func() {
	DescribeTable("should validate correctly",
		func(request validator.StringRequest, expectedErr bool) {
			response := validator.StringResponse{}
			SqsUrlValidator().ValidateString(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("valid SQS URL -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-east-1.amazonaws.com/123456789012/my-queue"),
			},
			false,
		),
		Entry("valid SQS FIPS URL -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs-fips.us-east-1.amazonaws.com/123456789012/my-queue"),
			},
			false,
		),
		Entry("valid SQS FIFO queue URL -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-east-1.amazonaws.com/123456789012/my-queue.fifo"),
			},
			false,
		),
		Entry("valid GovCloud SQS URL -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-gov-west-1.amazonaws.com/123456789012/my-queue"),
			},
			false,
		),
		Entry("valid GovCloud FIPS SQS URL -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs-fips.us-gov-west-1.amazonaws.com/123456789012/my-queue"),
			},
			false,
		),
		Entry("valid GovCloud FIPS FIFO SQS URL -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs-fips.us-gov-west-1.amazonaws.com/123456789012/my-queue.fifo"),
			},
			false,
		),
		Entry("HTTP scheme -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("http://sqs.us-east-1.amazonaws.com/123456789012/my-queue"),
			},
			true,
		),
		Entry("missing account ID segment -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-east-1.amazonaws.com/my-queue"),
			},
			true,
		),
		Entry("no path segments -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-east-1.amazonaws.com/"),
			},
			true,
		),
		Entry("wrong service (S3) -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://s3.us-east-1.amazonaws.com/123456789012/my-queue"),
			},
			true,
		),
		Entry("query string present -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-east-1.amazonaws.com/123456789012/my-queue?foo=bar"),
			},
			true,
		),
		Entry("fragment present -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-east-1.amazonaws.com/123456789012/my-queue#fragment"),
			},
			true,
		),
		Entry("userinfo present -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://user:pass@sqs.us-east-1.amazonaws.com/123456789012/my-queue"),
			},
			true,
		),
		Entry("unparseable URL with control char -> error",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringValue("https://sqs.us-east-1.amazonaws.com/\x7f/queue"),
			},
			true,
		),
		Entry("null value -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringNull(),
			},
			false,
		),
		Entry("unknown value -> ok",
			validator.StringRequest{
				Path:           path.Root("spot_termination_queue_url"),
				PathExpression: path.MatchRoot("spot_termination_queue_url"),
				ConfigValue:    types.StringUnknown(),
			},
			false,
		),
	)
})

var _ = Describe("ExtractRegionFromSqsUrl", func() {
	DescribeTable("should extract region correctly",
		func(url string, expectedRegion string, expectErr bool) {
			region, err := ExtractRegionFromSqsUrl(url)
			if expectErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(region).To(Equal(expectedRegion))
			}
		},
		Entry("SQS URL",
			"https://sqs.us-east-1.amazonaws.com/123456789012/my-queue",
			"us-east-1",
			false,
		),
		Entry("FIFO SQS URL",
			"https://sqs.us-east-1.amazonaws.com/123456789012/my-queue.fifo",
			"us-east-1",
			false,
		),
		Entry("FIPS SQS URL",
			"https://sqs-fips.us-west-2.amazonaws.com/123456789012/my-queue",
			"us-west-2",
			false,
		),
		Entry("eu-west-1 region",
			"https://sqs.eu-west-1.amazonaws.com/123456789012/my-queue",
			"eu-west-1",
			false,
		),
		Entry("GovCloud SQS URL",
			"https://sqs.us-gov-west-1.amazonaws.com/123456789012/my-queue",
			"us-gov-west-1",
			false,
		),
		Entry("GovCloud FIPS FIFO SQS URL",
			"https://sqs-fips.us-gov-west-1.amazonaws.com/123456789012/my-queue.fifo",
			"us-gov-west-1",
			false,
		),
		Entry("invalid host pattern",
			"https://s3.us-east-1.amazonaws.com/bucket/key",
			"",
			true,
		),
		Entry("unparseable URL",
			"not-a-url",
			"",
			true,
		),
	)
})
