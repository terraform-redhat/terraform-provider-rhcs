/*
Copyright (c) 2024 Red Hat, Inc.

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
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	"github.com/onsi/gomega/format"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

func TestProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	TestingT = t
	RunSpecs(t, "Common")
}

var testCtx *TestContext

var _ = BeforeEach(func() {
	format.MaxLength = 0 // set gomega format MaxLength to 0 to see all the diff when fails

	// Create a new test context for this test
	testCtx = NewTestContext()
})

var _ = AfterEach(func() {
	// Verify all handlers were consumed and close resources
	testCtx.VerifyAllHandlersConsumed()
	testCtx.Close()
})
