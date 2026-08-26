// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package provider

import "testing"

func TestRegionFromHyperfleetURL(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"https://abc123.execute-api.us-east-1.amazonaws.com/prod", "us-east-1"},
		{"https://abc123.execute-api.eu-west-2.amazonaws.com/prod", "eu-west-2"},
		{"https://abc123.execute-api.ap-southeast-1.amazonaws.com/prod", "ap-southeast-1"},
		{"https://abc123.execute-api.us-gov-east-1.amazonaws.com/prod", "us-gov-east-1"},
		{"https://abc123.execute-api.us-gov-west-1.amazonaws.com/prod", "us-gov-west-1"},
		// Path suffix with region-like segment — only the first match (hostname) is returned
		{"https://abc123.execute-api.us-east-1.amazonaws.com", "us-east-1"},
		// No region in URL
		{"https://example.com/api", ""},
		{"", ""},
	}
	for _, tc := range cases {
		got := regionFromHyperfleetURL(tc.url)
		if got != tc.want {
			t.Errorf("regionFromHyperfleetURL(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}
