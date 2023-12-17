/*
Copyright (c) 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeletconfig

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	description                  = "Validates the requested podPidsLimit value for the KubeletConfig"
	MinPodPidsLimit        int64 = 4096
	MaxPodPidsLimit        int64 = 16384
	MaxUnsafePidsLimit     int64 = 3694303
	PidsTooLowSummary            = "The requested podPidsLimit is too low"
	PidsTooHighSummary           = "The requested podPidsLimit is too high"
	PidsTooLowDescription        = "The requested podPidsLimit of '%d' is below the minimum allowable value of '%d'"
	PidsTooHighDescription       = "The requested podPidsLimit of '%d' is above the default maximum value of '%d', " +
		"and above the maximum supported value of '%d'"
	PidsTooHighWarningSummary     = "The requested podPidsLimit is possibly too high"
	PidsTooHighWarningDescription = "The requested podPidsLimit of '%d' is above the default maximum of '%d'. " +
		"This is not supported unless agreed with Red Hat in advance."
)

type PidsLimitValidator struct {
}

func (p PidsLimitValidator) Description(_ context.Context) string {
	return description
}

func (p PidsLimitValidator) MarkdownDescription(_ context.Context) string {
	return description
}

func (p PidsLimitValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {

	requestedPidsLimit := req.ConfigValue.ValueInt64()
	if requestedPidsLimit < MinPodPidsLimit {
		resp.Diagnostics.AddAttributeError(req.Path, PidsTooLowSummary,
			fmt.Sprintf(PidsTooLowDescription, requestedPidsLimit, MinPodPidsLimit))
	}

	// Fail if the user is trying to go beyond all supported limits of Pids
	if requestedPidsLimit > MaxUnsafePidsLimit {
		resp.Diagnostics.AddAttributeError(req.Path, PidsTooHighSummary,
			fmt.Sprintf(PidsTooHighDescription, requestedPidsLimit, MaxPodPidsLimit, MaxUnsafePidsLimit))
	}

	// If the user is trying to exceed the MaxPidsLimit, warn them that this is dependent upon them
	// having agreed this with Red Hat support in advance and having the capability to do so.
	if requestedPidsLimit > MaxPodPidsLimit && requestedPidsLimit <= MaxUnsafePidsLimit {
		resp.Diagnostics.AddAttributeWarning(req.Path, PidsTooHighWarningSummary,
			fmt.Sprintf(PidsTooHighWarningDescription, requestedPidsLimit, MaxPodPidsLimit))
	}
}

var _ validator.Int64 = PidsLimitValidator{}
