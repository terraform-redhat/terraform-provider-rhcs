// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package profilehandler

import (
	"errors"
	"testing"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

func TestIsVPCCleanupDependencyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "aws dependency violation error code",
			err:  errors.New("An error occurred (DependencyViolation) when calling the DeleteVpc operation"),
			want: true,
		},
		{
			name: "plain dependency wording",
			err:  errors.New("The vpc 'vpc-1' has dependencies and cannot be deleted"),
			want: true,
		},
		{
			name: "unrelated error",
			err:  errors.New("invalid parameter"),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isVPCCleanupDependencyError(tc.err)
			if got != tc.want {
				t.Fatalf("isVPCCleanupDependencyError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewProfileHandlerByName(t *testing.T) {
	// The seaplane profiles are checked into the repo, so the by-name
	// constructor must resolve them without CLUSTER_PROFILE being set.
	t.Setenv("CLUSTER_PROFILE", "")

	handler, err := NewProfileHandlerByName("sp-no-caps")
	if err != nil {
		t.Fatalf("loading profile by name: %v", err)
	}
	if handler.Profile().GetName() != "sp-no-caps" {
		t.Errorf("profile name = %q, want sp-no-caps", handler.Profile().GetName())
	}
	if got := handler.Profile().GetClusterType(); got != constants.SEAPLANE {
		t.Errorf("cluster type = %v, want seaplane", got)
	}
	if len(handler.Profile().GetDisableClusterCapabilities()) == 0 {
		t.Error("profile must define disable_cluster_capabilities")
	}

	if _, err := NewProfileHandlerByName("no-such-profile"); err == nil {
		t.Error("expected an error for a nonexistent profile")
	}
}
