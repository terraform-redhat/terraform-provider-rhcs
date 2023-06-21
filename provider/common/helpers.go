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
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ocmerrors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/pkg/errors"
	"github.com/zgalor/weberr"
)

const versionPrefix = "openshift-v"

// shouldPatchInt changed checks if the change between the given state and plan requires sending a
// patch request to the server. If it does it returns the value to add to the patch.
func ShouldPatchInt(state, plan types.Int64) (value int64, ok bool) {
	if plan.Unknown || plan.Null {
		return
	}
	if state.Unknown || state.Null {
		value = plan.Value
		ok = true
		return
	}
	if plan.Value != state.Value {
		value = plan.Value
		ok = true
	}
	return
}

// shouldPatchString changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it returns the value to add to the patch.
func ShouldPatchString(state, plan types.String) (value string, ok bool) {
	if plan.Unknown || plan.Null {
		return
	}
	if state.Unknown || state.Null {
		value = plan.Value
		ok = true
		return
	}
	if plan.Value != state.Value {
		value = plan.Value
		ok = true
	}
	return
}

// ShouldPatchBool changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it return the value to add to the patch.
func ShouldPatchBool(state, plan types.Bool) (value bool, ok bool) {
	if plan.Unknown || plan.Null {
		return
	}
	if state.Unknown || state.Null {
		value = plan.Value
		ok = true
		return
	}
	if plan.Value != state.Value {
		value = plan.Value
		ok = true
	}
	return
}

// ShouldPatchMap changed checks if the change between the given state and plan requires sending
// a patch request to the server. If it does it return the value to add to the patch.
func ShouldPatchMap(state, plan types.Map) (types.Map, bool) {
	return plan, !reflect.DeepEqual(state.Elems, plan.Elems)
}

// TF types converter functions
func StringArrayToList(arr []string) types.List {
	list := types.List{
		ElemType: types.StringType,
		Elems:    []attr.Value{},
	}

	for _, elm := range arr {
		list.Elems = append(list.Elems, types.String{Value: elm})
	}

	return list
}

func StringListToArray(list types.List) ([]string, error) {
	arr := []string{}
	for _, elm := range list.Elems {
		stype, ok := elm.(types.String)
		if !ok {
			return arr, errors.New("Failed to convert TF list to string slice.")
		}
		arr = append(arr, stype.Value)
	}
	return arr, nil
}

func IsValidDomain(candidate string) bool {
	var domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
	return domainRegexp.MatchString(candidate)
}

func IsValidEmail(candidate string) bool {
	var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$")
	return emailRegexp.MatchString(candidate)
}

func IsStringAttributeEmpty(param types.String) bool {
	return param.Unknown || param.Null || param.Value == ""
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
