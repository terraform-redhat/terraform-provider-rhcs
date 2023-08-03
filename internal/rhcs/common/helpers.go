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
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

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

func IsStringAttributeEmpty(param *string) bool {
	return param == nil || *param == ""
}

func IsListAttributeEmpty[T any](param []T) bool {
	return param == nil || len(param) < 1
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
				hash, err := sha1Hash(cert.Raw)
				if err != nil {
					return "", err
				}
				return hash, nil
			}
		}
	}

	// Fall back to using the last certficiate in the chain
	cert := certChain[len(certChain)-1]
	return sha1Hash(cert.Raw)
}

// sha1Hash computes the SHA1 of the byte array and returns the hex encoding as a string.
func sha1Hash(data []byte) (string, error) {
	// nolint:gosec
	hasher := sha1.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", fmt.Errorf("Couldn't calculate hash:\n %v", err)
	}
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed), nil
}

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type DefaultHttpClient struct {
}

func (c DefaultHttpClient) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

// Wait till the cluster is in ready or error state
func waitTillClusterChangedHisState(ctx context.Context, clusterCollection *cmv1.ClustersClient,
	clusterID string, timeout int64) (cmv1.ClusterState, error) {
	// Wait till the cluster is ready or error:
	resource := clusterCollection.Cluster(clusterID)
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	var object *cmv1.Cluster

	defer cancel()
	_, err := resource.Poll().
		Interval(30 * time.Second).
		Predicate(func(get *cmv1.ClusterGetResponse) bool {
			object = get.Body()
			tflog.Debug(ctx, fmt.Sprintf("cluster state is %s", object.State()))
			switch object.State() {
			case cmv1.ClusterStateReady,
				cmv1.ClusterStateError:
				return true
			}
			return false
		}).
		StartContext(pollCtx)
	if err != nil {
		return cmv1.ClusterStateUnknown, err
	}

	return object.State(), nil
}

// RetryClusterReadiness If the cluster become to `ready` or `error` return without any error
func RetryClusterReadiness(ctx context.Context, attempts int, sleep time.Duration, clusterCollection *cmv1.ClustersClient,
	clusterID string, timeout int64) (cmv1.ClusterState, error) {
	clusterState, err := waitTillClusterChangedHisState(ctx, clusterCollection, clusterID, timeout)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return RetryClusterReadiness(ctx, attempts, 2*sleep, clusterCollection, clusterID, timeout)
		}
		return cmv1.ClusterStateUnknown, err
	}

	return clusterState, nil
}

// WaitTillClusterIsReadyOrFailWithTimeout If the cluster is in `ready` (only) state return without any error, otherwise, return an error
func WaitTillClusterIsReadyOrFailWithTimeout(ctx context.Context, clusterCollection *cmv1.ClustersClient,
	clusterID string, timeout int64) error {
	clusterState, err := waitTillClusterChangedHisState(ctx, clusterCollection, clusterID, timeout)
	if err != nil {
		return err
	}

	if clusterState == cmv1.ClusterStateError {
		return fmt.Errorf("cluster state is error")
	}

	return nil
}

// WaitTillClusterIsReadyOrFail - Use the default timeout - 60 minutes
func WaitTillClusterIsReadyOrFail(ctx context.Context, clusterCollection *cmv1.ClustersClient, clusterID string) error {
	return WaitTillClusterIsReadyOrFailWithTimeout(ctx, clusterCollection, clusterID, 60)
}
