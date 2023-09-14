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
	"bytes"
	"crypto/sha1"
	"encoding/hex"
***REMOVED***
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
***REMOVED***

const versionPrefix = "openshift-v"

var EmailRegexp = regexp.MustCompile("^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$"***REMOVED***

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
	return EmailRegexp.MatchString(candidate***REMOVED***
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

func GetThumbprint(oidcEndpointURL string, httpClient HttpClient***REMOVED*** (thumbprint string, err error***REMOVED*** {
	defer func(***REMOVED*** {
		if panicErr := recover(***REMOVED***; panicErr != nil {
			fmt.Fprintf(os.Stderr, "recovering from: %q\n", panicErr***REMOVED***
			thumbprint = ""
			err = fmt.Errorf("recovering from: %q", panicErr***REMOVED***
***REMOVED***
	}(***REMOVED***

	connect, err := url.ParseRequestURI(oidcEndpointURL***REMOVED***
	if err != nil {
		return "", err
	}

	response, err := httpClient.Get(fmt.Sprintf("https://%s:443", connect.Host***REMOVED******REMOVED***
	if err != nil {
		return "", err
	}

	certChain := response.TLS.PeerCertificates

	// Grab the CA in the chain
	for _, cert := range certChain {
		if cert.IsCA {
			if bytes.Equal(cert.RawIssuer, cert.RawSubject***REMOVED*** {
				hash, err := Sha1Hash(cert.Raw***REMOVED***
				if err != nil {
					return "", err
		***REMOVED***
				return hash, nil
	***REMOVED***
***REMOVED***
	}

	// Fall back to using the last certficiate in the chain
	cert := certChain[len(certChain***REMOVED***-1]
	return Sha1Hash(cert.Raw***REMOVED***
}

// sha1Hash computes the SHA1 of the byte array and returns the hex encoding as a string.
func Sha1Hash(data []byte***REMOVED*** (string, error***REMOVED*** {
	// nolint:gosec
	hasher := sha1.New(***REMOVED***
	_, err := hasher.Write(data***REMOVED***
	if err != nil {
		return "", fmt.Errorf("Couldn't calculate hash:\n %v", err***REMOVED***
	}
	hashed := hasher.Sum(nil***REMOVED***
	return hex.EncodeToString(hashed***REMOVED***, nil
}

// HasValue checks if the given terraform value is set.
func HasValue(val attr.Value***REMOVED*** bool {
	return !val.IsUnknown(***REMOVED*** && !val.IsNull(***REMOVED***
}
