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

// Package providerdata defines the shared data passed from the provider Configure
// method to every resource and data source via ResourceData/DataSourceData.
package providerdata

import (
	sdk "github.com/openshift-online/ocm-sdk-go"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
)

// ProviderSharedData is set as resp.ResourceData and resp.DataSourceData by the
// provider Configure method. Resources extract the client they need from here.
type ProviderSharedData struct {
	// OCMConnection is the OCM API connection. Nil when no OCM credentials are
	// configured (hyperfleet-only mode).
	OCMConnection *sdk.Connection

	// HyperfleetClient is the Platform API v2 clientset. Nil when hyperfleet_url
	// is not set in the provider configuration.
	HyperfleetClient *hyperfleet.Clientset

	// HyperfleetAccountID is the AWS account ID used to namespace Platform API
	// calls (e.g. Clusters(accountID)).
	HyperfleetAccountID string

	// HyperfleetCallerARN is the ARN of the AWS caller identity forwarded to the
	// Platform API as the cluster CreatorARN.
	HyperfleetCallerARN string
}

// OCMConn extracts the OCM SDK connection from whatever value was stored as
// ProviderData. It handles both the legacy *sdk.Connection type (left over from
// any cached Terraform state) and the current *ProviderSharedData type.
//
// Returns (nil, false) when data is nil or does not contain an OCM connection.
func OCMConn(data any) (*sdk.Connection, bool) {
	switch v := data.(type) {
	case *sdk.Connection:
		return v, v != nil
	case *ProviderSharedData:
		if v == nil {
			return nil, false
		}
		return v.OCMConnection, v.OCMConnection != nil
	default:
		return nil, false
	}
}
