package ci

/*
Copyright (c***REMOVED*** 2018 Red Hat, Inc.

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

***REMOVED***
***REMOVED***
	"os"

***REMOVED***
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"

	// . "github.com/onsi/gomega"

	client "github.com/openshift-online/ocm-sdk-go"
***REMOVED***

var (
	RHCSOCMToken = os.Getenv(CON.TokenENVName***REMOVED***
***REMOVED***

// Regular users in the organization 'Red Hat-Service Delivery-tester'
var (
	RHCSConnection = createConnectionWithToken(RHCSOCMToken***REMOVED***
***REMOVED***

var (
	// Create a logger:
	logger = createLogger(***REMOVED***
***REMOVED***

func createConnectionWithToken(token string***REMOVED*** *client.Connection {
	gatewayURL := gatewayURL(***REMOVED***

	// Create the connection:
	connection, err := client.NewConnectionBuilder(***REMOVED***.
		Logger(logger***REMOVED***.
		Insecure(true***REMOVED***.
		TokenURL(tokenURL***REMOVED***.
		URL(gatewayURL***REMOVED***.
		Client(clientID, clientSecret***REMOVED***.
		Tokens(token***REMOVED***.
		Build(***REMOVED***
	if err != nil {
		fmt.Printf("ERROR occurred when create connection with token: %s!! %s\n", token, err***REMOVED***
	}
	return connection

}

func createLogger(***REMOVED*** client.Logger {
	logger, _ := client.NewStdLoggerBuilder(***REMOVED***.
		Streams(GinkgoWriter, GinkgoWriter***REMOVED***.
		Build(***REMOVED***

	return logger
}

type Response struct {
	StatusCode int
	Body       []byte
}

func (r *Response***REMOVED*** Status(***REMOVED*** int {
	return r.StatusCode
}
func (r *Response***REMOVED*** Bytes(***REMOVED*** []byte {
	return r.Body
}

func (r *Response***REMOVED*** String(***REMOVED*** string {
	return string(r.Bytes(***REMOVED******REMOVED***
}
