/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

***REMOVED***
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ocmerrors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/pkg/errors"
	"github.com/zgalor/weberr"
***REMOVED***

const versionPrefix = "openshift-v"

// shouldPatchInt changed checks if the change between the given state and plan requires sending a
// patch request to the server. If it does it returns the value to add to the patch.
func ShouldPatchInt(state, plan types.Int64***REMOVED*** (value int64, ok bool***REMOVED*** {
	if plan.IsUnknown(***REMOVED*** || plan.IsNull(***REMOVED*** {
		return
	}
	if state.IsUnknown(***REMOVED*** || state.IsNull(***REMOVED*** {
		value = plan.ValueInt64(***REMOVED***
		ok = true
		return
	}
	if plan.ValueInt64(***REMOVED*** != state.ValueInt64(***REMOVED*** {
		value = plan.ValueInt64(***REMOVED***
		ok = true
	}
	return
}

// shouldPatchString changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it returns the value to add to the patch.
func ShouldPatchString(state, plan types.String***REMOVED*** (value string, ok bool***REMOVED*** {
	if plan.IsUnknown(***REMOVED*** || plan.IsNull(***REMOVED*** {
		return
	}
	if state.IsUnknown(***REMOVED*** || state.IsNull(***REMOVED*** {
		value = plan.ValueString(***REMOVED***
		ok = true
		return
	}
	if plan.ValueString(***REMOVED*** != state.ValueString(***REMOVED*** {
		value = plan.ValueString(***REMOVED***
		ok = true
	}
	return
}

// ShouldPatchBool changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it return the value to add to the patch.
func ShouldPatchBool(state, plan types.Bool***REMOVED*** (value bool, ok bool***REMOVED*** {
	if plan.IsUnknown(***REMOVED*** || plan.IsNull(***REMOVED*** {
		return
	}
	if state.IsUnknown(***REMOVED*** || state.IsNull(***REMOVED*** {
		value = plan.ValueBool(***REMOVED***
		ok = true
		return
	}
	if plan.ValueBool(***REMOVED*** != state.ValueBool(***REMOVED*** {
		value = plan.ValueBool(***REMOVED***
		ok = true
	}
	return
}

// ShouldPatchMap changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it return the value to add to the patch.
func ShouldPatchMap(state, plan types.Map***REMOVED*** (types.Map, bool***REMOVED*** {
	return plan, !reflect.DeepEqual(state.Elements(***REMOVED***, plan.Elements(***REMOVED******REMOVED***
}

func IsValidDomain(candidate string***REMOVED*** bool {
	var domainRegexp = regexp.MustCompile(`^(?i***REMOVED***[a-z0-9-]+(\.[a-z0-9-]+***REMOVED***+\.?$`***REMOVED***
	return domainRegexp.MatchString(candidate***REMOVED***
}

func IsValidEmail(candidate string***REMOVED*** bool {
	var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$"***REMOVED***
	return emailRegexp.MatchString(candidate***REMOVED***
}

func IsStringAttributeEmpty(param types.String***REMOVED*** bool {
	return param.IsUnknown(***REMOVED*** || param.IsNull(***REMOVED*** || param.ValueString(***REMOVED*** == ""
}

func IsGreaterThanOrEqual(version1, version2 string***REMOVED*** (bool, error***REMOVED*** {
	v1, err := version.NewVersion(strings.TrimPrefix(version1, versionPrefix***REMOVED******REMOVED***
	if err != nil {
		return false, err
	}
	v2, err := version.NewVersion(strings.TrimPrefix(version2, versionPrefix***REMOVED******REMOVED***
	if err != nil {
		return false, err
	}
	return v1.GreaterThanOrEqual(v2***REMOVED***, nil
}

func HandleErr(res *ocmerrors.Error, err error***REMOVED*** error {
	msg := res.Reason(***REMOVED***
	if msg == "" {
		msg = err.Error(***REMOVED***
	}
	errType := weberr.ErrorType(res.Status(***REMOVED******REMOVED***
	return errType.Set(errors.Errorf("%s", msg***REMOVED******REMOVED***
}
