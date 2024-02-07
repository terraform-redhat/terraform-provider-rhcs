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

package hcp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

// All tests will use this API server and Terraform runner:
var (
	server    *Server
	terraform *TerraformRunner
	testingT  *testing.T
)

func TestProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	testingT = t
	RunSpecs(t, "HCP")
}

var _ = BeforeEach(func() {
	// Create the server:
	var ca string
	server, ca = MakeTCPTLSServer()
	// Create an access token:
	token := MakeTokenString("Bearer", 10*time.Minute)

	// Create the runner:
	terraform = NewTerraformRunner().
		URL(server.URL()).
		CA(ca).
		Token(token).
		Build()
})

var _ = AfterEach(func() {
	// Close the server:
	server.Close()

	// Close the runner:
	terraform.Close()
})

// TerraformRunnerBuilder contains the data and logic needed to build a terraform runner.
type TerraformRunnerBuilder struct {
	url   string
	ca    string
	token string
}

// TerraformRunner contains the data and logic needed to run Terraform.
type TerraformRunner struct {
	binary string
	dir    string
	env    []string
}

// NewTerraformRunner creates a new Terraform runner.
func NewTerraformRunner() *TerraformRunnerBuilder {
	return &TerraformRunnerBuilder{}
}

// URL sets the URL of the OCM API server.
func (b *TerraformRunnerBuilder) URL(value string) *TerraformRunnerBuilder {
	b.url = value
	return b
}

// CA sets the trusted certificates used to connect to the OCM API server.
func (b *TerraformRunnerBuilder) CA(value string) *TerraformRunnerBuilder {
	b.ca = value
	return b
}

// Token sets the authentication token used to connect to the OCM API server.
func (b *TerraformRunnerBuilder) Token(value string) *TerraformRunnerBuilder {
	b.token = value
	return b
}

// Build uses the information stored in the builder to create a new Terraform runner.
func (b *TerraformRunnerBuilder) Build() *TerraformRunner {
	// Check parameters:
	ExpectWithOffset(1, b.url).ToNot(BeEmpty())
	ExpectWithOffset(1, b.ca).ToNot(BeEmpty())
	ExpectWithOffset(1, b.token).ToNot(BeEmpty())

	// Check that the Terraform tfBinary is available in the path:
	tfBinary, err := exec.LookPath("terraform")
	Expect(err).ToNot(HaveOccurred())

	// Create a temporary directory for the files so that we don't interfere with the
	// configuration that may already exist for the user running the tests.
	tmpDir, err := ioutil.TempDir("", "rhcs-test-*.d")
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	// Create the main file:
	mainPath := filepath.Join(tmpDir, "main.tf")
	mainContent := EvaluateTemplate(`
		terraform {
		  required_providers {
		    rhcs = {
                source = "terraform.local/local/rhcs"
                version = ">= 0.0.1"
		    }
		  }
		}

		provider "rhcs" {
		  url         = "{{ .URL }}"
		  token       = "{{ .Token }}"
		  trusted_cas = file("{{ .CA }}")
		}
		`,
		"URL", b.url,
		"Token", b.token,
		"CA", strings.ReplaceAll(b.ca, "\\", "/"),
	)
	err = ioutil.WriteFile(mainPath, []byte(mainContent), 0600)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	// Parse the current environment into a map so that it is easy to update it:
	envMap := map[string]string{}
	for _, text := range os.Environ() {
		index := strings.Index(text, "=")
		var name string
		var value string
		if index > 0 {
			name = text[0:index]
			value = text[index+1:]
		} else {
			name = text
			value = ""
		}
		envMap[name] = value
	}

	// Enable verbose debug:
	envMap["TF_LOG"] = "DEBUG"

	// Reconstruct the environment list:
	envList := make([]string, 0, len(envMap))
	for name, value := range envMap {
		envList = append(envList, name+"="+value)
	}

	// Run the init command:
	initCmd := exec.Command(tfBinary, "init")
	initCmd.Env = envList
	initCmd.Dir = tmpDir
	initCmd.Stdout = GinkgoWriter
	initCmd.Stderr = GinkgoWriter
	err = initCmd.Run()
	if err != nil {
		message := fmt.Sprintf(
			"Terraform init finished with exit code %d",
			initCmd.ProcessState.ExitCode(),
		)
		Fail(message, 1)
	}

	// Create and populate the object:
	return &TerraformRunner{
		binary: tfBinary,
		dir:    tmpDir,
		env:    envList,
	}
}

// Source sets the Terraform source of the test.
func (r *TerraformRunner) Source(text string) {
	file := filepath.Join(r.dir, "test.tf")
	err := ioutil.WriteFile(file, []byte(text), 0600)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}

// Run runs a command.
func (r *TerraformRunner) Run(args ...string) int {
	var err error

	// Run the command:
	cmd := exec.Command(r.binary, args...)
	cmd.Env = r.env
	cmd.Dir = r.dir
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	err = cmd.Run()
	switch err.(type) {
	case *exec.ExitError:
		// Nothing, this is a normal situation and the caller is expected to check the
		// returned exit code.
	default:
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
	}

	// Return the exit code:
	return cmd.ProcessState.ExitCode()
}

// Validate runs the `validate` command.
func (r *TerraformRunner) Validate() int {
	return r.Run("validate")
}

// Apply runs the `apply` command.
func (r *TerraformRunner) Apply() int {
	return r.Run("apply", "-auto-approve")
}

// Destroy runs the `destroy` command.
func (r *TerraformRunner) Destroy() int {
	return r.Run("destroy", "-auto-approve")
}

// Import runs the `import` command.
func (r *TerraformRunner) Import(args ...string) int {
	return r.Run(append([]string{"import"}, args...)...)
}

// State returns the reads the Terraform state and returns the result of parsing
// it as a JSON document.
func (r *TerraformRunner) State() interface{} {
	path := filepath.Join(r.dir, "terraform.tfstate")
	_, err := os.Stat(path)
	var result interface{}
	if err == nil {
		var data []byte
		data, err = ioutil.ReadFile(path)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		err = json.Unmarshal(data, &result)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
	}
	return result
}

// Resource returns the resource stored in the state with the given type and identifier.
func (r *TerraformRunner) Resource(typ, name string) interface{} {
	state := r.State()
	filter := fmt.Sprintf(
		`.resources[] | select(.type == "%s" and .name == "%s") | .instances[]`,
		typ, name,
	)
	results, err := JQ(filter, state)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, results).To(
		HaveLen(1),
		"Expected exactly one resource with name type '%s' and name '%s', but found %d",
		typ, name, len(results),
	)
	return results[0]
}

// Close releases all the resources used by the Terraform runner and removes all
// temporary files and directories.
func (r *TerraformRunner) Close() {
	err := os.RemoveAll(r.dir)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}
