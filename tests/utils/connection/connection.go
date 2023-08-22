package connection

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

	. "github.com/onsi/ginkgo/v2"

	client "github.com/openshift-online/ocm-sdk-go"
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

	if token == "" {
		fmt.Println("[WARNING]: Token shouldn't be empty")
	}

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
