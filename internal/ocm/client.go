/*
Copyright (c) 2020 Red Hat, Inc.

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

package ocm

import (
	"fmt"
	"os"
	"strings"
	"time"

	sdk "github.com/openshift-online/ocm-sdk-go"
)

type Client struct {
	ocm *sdk.Connection
}

// ClientBuilder contains the information and logic needed to build a connection to OCM. Don't
// create instances of this type directly; use the NewClient function instead.
type ClientBuilder struct {
	cfg *config.Config
}

// NewClient creates a builder that can then be used to configure and build an OCM connection.
func NewClient() *ClientBuilder {
	return &ClientBuilder{}
}

// NewClientWithConnection creates a client with a preexisting connection for testing purpose
func NewClientWithConnection(connection *sdk.Connection) *Client {
	return &Client{
		ocm: connection,
	}
}
