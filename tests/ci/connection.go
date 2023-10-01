package ci

/*
Copyright (c) 2018 Red Hat, Inc.

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

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"

	client "github.com/openshift-online/ocm-sdk-go"
)

var (
	RHCSOCMToken = os.Getenv(CON.TokenENVName)
)

// Regular users in the organization 'Red Hat-Service Delivery-tester'
var (
	RHCSConnection = createConnectionWithToken(RHCSOCMToken)
)

var (
	// Create a logger:
	logger = createLogger()
)

func createConnectionWithToken(token string) *client.Connection {
	gatewayURL := gatewayURL()

	// Create the connection:
	connection, err := client.NewConnectionBuilder().
		Logger(logger).
		Insecure(true).
		TokenURL(tokenURL).
		URL(gatewayURL).
		Client(clientID, clientSecret).
		Tokens(token).
		Build()
	if err != nil {
		fmt.Printf("ERROR occurred when create connection with token: %s!! %s\n", token, err)
	}
	return connection

}

func createLogger() client.Logger {
	logger, _ := client.NewStdLoggerBuilder().
		Streams(GinkgoWriter, GinkgoWriter).
		Build()

	return logger
}

type Response struct {
	StatusCode int
	Body       []byte
}

func (r *Response) Status() int {
	return r.StatusCode
}
func (r *Response) Bytes() []byte {
	return r.Body
}

func (r *Response) String() string {
	return string(r.Bytes())
}
