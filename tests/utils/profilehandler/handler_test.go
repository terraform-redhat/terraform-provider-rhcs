// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package profilehandler

import (
	"errors"
	"testing"
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
