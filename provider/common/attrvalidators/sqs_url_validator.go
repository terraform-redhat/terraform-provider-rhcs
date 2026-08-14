// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package attrvalidators

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var sqsHostPattern = regexp.MustCompile(`^sqs(-fips)?\.([a-z0-9-]+)\.amazonaws\.com$`)

// sqsUrlValidator validates that a string Attribute's value is a valid SQS queue URL.
type sqsUrlValidator struct{}

// Description describes the validation in plain text formatting.
func (v sqsUrlValidator) Description(_ context.Context) string {
	return "value must be a valid SQS queue URL with HTTPS scheme, proper endpoint pattern, and account/queue path"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v sqsUrlValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v sqsUrlValidator) ValidateString(
	ctx context.Context, request validator.StringRequest, response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	// Parse URL
	parsedUrl, err := url.Parse(value)
	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Spot termination queue URL",
			fmt.Sprintf("URL could not be parsed: %s", err.Error()),
		)
		return
	}

	// Validate HTTPS scheme
	if parsedUrl.Scheme != "https" {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Spot termination queue URL",
			"Spot termination queue URL must use HTTPS scheme",
		)
		return
	}

	// Validate no userinfo
	if parsedUrl.User != nil && parsedUrl.User.String() != "" {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Spot termination queue URL",
			"Spot termination queue URL must not contain userinfo (username/password)",
		)
		return
	}

	// Validate host matches SQS endpoint pattern
	if !sqsHostPattern.MatchString(parsedUrl.Host) {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Spot termination queue URL",
			"Spot termination queue URL host must match SQS endpoint pattern "+
				"(sqs.<region>.amazonaws.com or sqs-fips.<region>.amazonaws.com)",
		)
		return
	}

	// Validate no query string
	if parsedUrl.RawQuery != "" {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Spot termination queue URL",
			"Spot termination queue URL must not contain a query string",
		)
		return
	}

	// Validate no fragment
	if parsedUrl.Fragment != "" {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Spot termination queue URL",
			"Spot termination queue URL must not contain a fragment",
		)
		return
	}

	// Validate path structure: /<account-id>/<queue-name>
	pathParts := strings.Split(strings.Trim(parsedUrl.Path, "/"), "/")
	if len(pathParts) != 2 || pathParts[0] == "" || pathParts[1] == "" {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Spot termination queue URL",
			"Spot termination queue URL path must contain an account ID and queue name (/<account-id>/<queue-name>)",
		)
		return
	}
}

// SqsUrlValidator returns a validator which ensures that the configured string
// is a valid SQS queue URL.
func SqsUrlValidator() validator.String {
	return sqsUrlValidator{}
}

// ExtractRegionFromSqsUrl extracts the AWS region from an SQS queue URL.
// Returns an error if the URL is invalid or doesn't match the SQS endpoint pattern.
func ExtractRegionFromSqsUrl(urlStr string) (string, error) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	matches := sqsHostPattern.FindStringSubmatch(parsedUrl.Host)
	if len(matches) < 3 {
		return "", fmt.Errorf("URL host does not match SQS endpoint pattern")
	}

	// matches[2] contains the region from the regex capture group
	return matches[2], nil
}
