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
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
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

var EmailRegexp = regexp.MustCompile("^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$")

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

func IsValidDomain(candidate string) bool {
	var domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
	return domainRegexp.MatchString(candidate)
}

func IsValidEmail(candidate string) bool {
	return EmailRegexp.MatchString(candidate)
}

func IsStringAttributeEmpty(param types.String) bool {
	return param.IsUnknown() || param.IsNull() || param.ValueString() == ""
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

func GetThumbprint(oidcEndpointURL string, httpClient HttpClient) (thumbprint string, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			fmt.Fprintf(os.Stderr, "recovering from: %q\n", panicErr)
			thumbprint = ""
			err = fmt.Errorf("recovering from: %q", panicErr)
		}
	}()

	connect, err := url.ParseRequestURI(oidcEndpointURL)
	if err != nil {
		return "", err
	}

	response, err := httpClient.Get(fmt.Sprintf("https://%s:443", connect.Host))
	if err != nil {
		return "", err
	}

	certChain := response.TLS.PeerCertificates

	// Grab the CA in the chain
	for _, cert := range certChain {
		if cert.IsCA {
			if bytes.Equal(cert.RawIssuer, cert.RawSubject) {
				hash, err := Sha1Hash(cert.Raw)
				if err != nil {
					return "", err
				}
				return hash, nil
			}
		}
	}

	// Fall back to using the last certficiate in the chain
	cert := certChain[len(certChain)-1]
	return Sha1Hash(cert.Raw)
}

// sha1Hash computes the SHA1 of the byte array and returns the hex encoding as a string.
func Sha1Hash(data []byte) (string, error) {
	// nolint:gosec
	hasher := sha1.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", fmt.Errorf("Couldn't calculate hash:\n %v", err)
	}
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed), nil
}

// HasValue checks if the given terraform value is set.
func HasValue(val attr.Value) bool {
	return !val.IsUnknown() && !val.IsNull()
}
