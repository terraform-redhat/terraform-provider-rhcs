/*
Copyright (c) 2021 Red Hat, Inc.

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

package common

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ocmerrors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/pkg/errors"
	"github.com/zgalor/weberr"
)

const (
	versionPrefix                         = "openshift-v"
	AssertionErrorSummaryMessage          = "Attribute value cannot be changed"
	AssertionErrorDetailsMessage          = "Attribute %s, cannot be changed from %v to %v"
	ValueCannotBeChangedStringDescription = "After the creation of the resource, it is not possible to update the attribute value."
)

// shouldPatchInt changed checks if the change between the given state and plan requires sending a
// patch request to the server. If it does it returns the value to add to the patch.
func ShouldPatchInt(state, plan types.Int64) (value int64, ok bool) {
	if plan.IsUnknown() || plan.IsNull() {
		return
	}
	if state.IsUnknown() || state.IsNull() {
		value = plan.ValueInt64()
		ok = true
		return
	}
	if plan.ValueInt64() != state.ValueInt64() {
		value = plan.ValueInt64()
		ok = true
	}
	return
}

// shouldPatchString changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it returns the value to add to the patch.
func ShouldPatchString(state, plan types.String) (value string, ok bool) {
	if plan.IsUnknown() || plan.IsNull() {
		return
	}
	if state.IsUnknown() || state.IsNull() {
		value = plan.ValueString()
		ok = true
		return
	}
	if plan.ValueString() != state.ValueString() {
		value = plan.ValueString()
		ok = true
	}
	return
}

// ShouldPatchBool changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it return the value to add to the patch.
func ShouldPatchBool(state, plan types.Bool) (value bool, ok bool) {
	if plan.IsUnknown() || plan.IsNull() {
		return
	}
	if state.IsUnknown() || state.IsNull() {
		value = plan.ValueBool()
		ok = true
		return
	}
	if plan.ValueBool() != state.ValueBool() {
		value = plan.ValueBool()
		ok = true
	}
	return
}

// ShouldPatchMap changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it return the value to add to the patch.
func ShouldPatchMap(state, plan types.Map) (types.Map, bool) {
	return plan, !reflect.DeepEqual(state.Elements(), plan.Elements())
}

// ShouldPatchList changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it return the value to add to the patch.
func ShouldPatchList(state, plan types.List) (types.List, bool) {
	return plan, !reflect.DeepEqual(state.Elements(), plan.Elements())
}

func IsValidDomain(candidate string) bool {
	var domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
	return domainRegexp.MatchString(candidate)
}

func EmptiableStringToStringType(s string) types.String {
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}

func IsStringAttributeUnknownOrEmpty(param types.String) bool {
	return param.IsUnknown() || param.IsNull() || param.ValueString() == ""
}

func IsStringAttributeKnownAndEmpty(param types.String) bool {
	return !param.IsUnknown() && (param.IsNull() || param.ValueString() == "")
}

func IsGreaterThanOrEqual(version1, version2 string) (bool, error) {
	v1, err := version.NewVersion(strings.TrimPrefix(version1, versionPrefix))
	if err != nil {
		return false, err
	}
	v2, err := version.NewVersion(strings.TrimPrefix(version2, versionPrefix))
	if err != nil {
		return false, err
	}
	return v1.GreaterThanOrEqual(v2), nil
}

func HandleErr(res *ocmerrors.Error, err error) error {
	msg := res.Reason()
	if msg == "" {
		msg = err.Error()
	}
	errType := weberr.ErrorType(res.Status())
	return errType.Set(errors.Errorf("%s", msg))
}

// HasValue checks if the given terraform value is set.
func HasValue(val attr.Value) bool {
	return !val.IsUnknown() && !val.IsNull()
}

// ValidateStateAndPlanEquals checks if given two attributes are equal, if not add error to diagnostic
func ValidateStateAndPlanEquals(stateAttr attr.Value, planAttr attr.Value, attrName string, diags *diag.Diagnostics) {
	if !stateAttr.Equal(planAttr) {
		diags.AddError(AssertionErrorSummaryMessage, fmt.Sprintf(AssertionErrorDetailsMessage, attrName, stateAttr, planAttr))
	}
}
