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

package provider

***REMOVED***
	"encoding/json"
***REMOVED***
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

// All tests will use this API server and Terraform runner:
var (
	server    *Server
	terraform *TerraformRunner
	testingT  *testing.T
***REMOVED***

func TestProvider(t *testing.T***REMOVED*** {
	RegisterFailHandler(Fail***REMOVED***
	testingT = t
	RunSpecs(t, "Provider"***REMOVED***
}

var _ = BeforeEach(func(***REMOVED*** {
	// Create the server:
	var ca string
	server, ca = MakeTCPTLSServer(***REMOVED***
	// Create an access token:
	token := MakeTokenString("Bearer", 10*time.Minute***REMOVED***

	// Create the runner:
	terraform = NewTerraformRunner(***REMOVED***.
		URL(server.URL(***REMOVED******REMOVED***.
		CA(ca***REMOVED***.
		Token(token***REMOVED***.
		Build(***REMOVED***
}***REMOVED***

var _ = AfterEach(func(***REMOVED*** {
	// Close the server:
	server.Close(***REMOVED***

	// Close the runner:
	terraform.Close(***REMOVED***
}***REMOVED***

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
func NewTerraformRunner(***REMOVED*** *TerraformRunnerBuilder {
	return &TerraformRunnerBuilder{}
}

// URL sets the URL of the OCM API server.
func (b *TerraformRunnerBuilder***REMOVED*** URL(value string***REMOVED*** *TerraformRunnerBuilder {
	b.url = value
	return b
}

// CA sets the trusted certificates used to connect to the OCM API server.
func (b *TerraformRunnerBuilder***REMOVED*** CA(value string***REMOVED*** *TerraformRunnerBuilder {
	b.ca = value
	return b
}

// Token sets the authentication token used to connect to the OCM API server.
func (b *TerraformRunnerBuilder***REMOVED*** Token(value string***REMOVED*** *TerraformRunnerBuilder {
	b.token = value
	return b
}

// Build uses the information stored in the builder to create a new Terraform runner.
func (b *TerraformRunnerBuilder***REMOVED*** Build(***REMOVED*** *TerraformRunner {
	// Check parameters:
	ExpectWithOffset(1, b.url***REMOVED***.ToNot(BeEmpty(***REMOVED******REMOVED***
	ExpectWithOffset(1, b.ca***REMOVED***.ToNot(BeEmpty(***REMOVED******REMOVED***
	ExpectWithOffset(1, b.token***REMOVED***.ToNot(BeEmpty(***REMOVED******REMOVED***

	// Check that the Terraform tfBinary is available in the path:
	tfBinary, err := exec.LookPath("terraform"***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

	// Create a temporary directory for the files so that we don't interfere with the
	// configuration that may already exist for the user running the tests.
	tmpDir, err := ioutil.TempDir("", "rhcs-test-*.d"***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

	// Create the main file:
	mainPath := filepath.Join(tmpDir, "main.tf"***REMOVED***
	mainContent := EvaluateTemplate(`
		terraform {
		  required_providers {
		    rhcs = {
                source = "terraform.local/local/rhcs"
                version = ">= 0.0.1"
		    }
		  }
***REMOVED***

		provider "rhcs" {
		  url         = "{{ .URL }}"
		  token       = "{{ .Token }}"
		  trusted_cas = file("{{ .CA }}"***REMOVED***
***REMOVED***
		`,
		"URL", b.url,
		"Token", b.token,
		"CA", strings.ReplaceAll(b.ca, "\\", "/"***REMOVED***,
	***REMOVED***
	err = ioutil.WriteFile(mainPath, []byte(mainContent***REMOVED***, 0600***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

	// Parse the current environment into a map so that it is easy to update it:
	envMap := map[string]string{}
	for _, text := range os.Environ(***REMOVED*** {
		index := strings.Index(text, "="***REMOVED***
		var name string
		var value string
		if index > 0 {
			name = text[0:index]
			value = text[index+1:]
***REMOVED*** else {
			name = text
			value = ""
***REMOVED***
		envMap[name] = value
	}

	// Enable verbose debug:
	envMap["TF_LOG"] = "DEBUG"

	// Reconstruct the environment list:
	envList := make([]string, 0, len(envMap***REMOVED******REMOVED***
	for name, value := range envMap {
		envList = append(envList, name+"="+value***REMOVED***
	}

	// Run the init command:
	initCmd := exec.Command(tfBinary, "init"***REMOVED***
	initCmd.Env = envList
	initCmd.Dir = tmpDir
	initCmd.Stdout = GinkgoWriter
	initCmd.Stderr = GinkgoWriter
	err = initCmd.Run(***REMOVED***
	if err != nil {
		message := fmt.Sprintf(
			"Terraform init finished with exit code %d",
			initCmd.ProcessState.ExitCode(***REMOVED***,
		***REMOVED***
		Fail(message, 1***REMOVED***
	}

	// Create and populate the object:
	return &TerraformRunner{
		binary: tfBinary,
		dir:    tmpDir,
		env:    envList,
	}
}

// Source sets the Terraform source of the test.
func (r *TerraformRunner***REMOVED*** Source(text string***REMOVED*** {
	file := filepath.Join(r.dir, "test.tf"***REMOVED***
	err := ioutil.WriteFile(file, []byte(text***REMOVED***, 0600***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
}

// Run runs a command.
func (r *TerraformRunner***REMOVED*** Run(args ...string***REMOVED*** int {
	var err error

	// Run the command:
	cmd := exec.Command(r.binary, args...***REMOVED***
	cmd.Env = r.env
	cmd.Dir = r.dir
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	err = cmd.Run(***REMOVED***
	switch err.(type***REMOVED*** {
	case *exec.ExitError:
		// Nothing, this is a normal situation and the caller is expected to check the
		// returned exit code.
	default:
		ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}

	// Return the exit code:
	return cmd.ProcessState.ExitCode(***REMOVED***
}

// Validate runs the `validate` command.
func (r *TerraformRunner***REMOVED*** Validate(***REMOVED*** int {
	return r.Run("validate"***REMOVED***
}

// Apply runs the `apply` command.
func (r *TerraformRunner***REMOVED*** Apply(***REMOVED*** int {
	return r.Run("apply", "-auto-approve"***REMOVED***
}

// Destroy runs the `destroy` command.
func (r *TerraformRunner***REMOVED*** Destroy(***REMOVED*** int {
	return r.Run("destroy", "-auto-approve"***REMOVED***
}

// State returns the reads the Terraform state and returns the result of parsing
// it as a JSON document.
func (r *TerraformRunner***REMOVED*** State(***REMOVED*** interface{} {
	path := filepath.Join(r.dir, "terraform.tfstate"***REMOVED***
	_, err := os.Stat(path***REMOVED***
	var result interface{}
	if err == nil {
		var data []byte
		data, err = ioutil.ReadFile(path***REMOVED***
		ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		err = json.Unmarshal(data, &result***REMOVED***
		ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}
	return result
}

// Resource returns the resource stored in the state with the given type and identifier.
func (r *TerraformRunner***REMOVED*** Resource(typ, name string***REMOVED*** interface{} {
	state := r.State(***REMOVED***
	filter := fmt.Sprintf(
		`.resources[] | select(.type == "%s" and .name == "%s"***REMOVED*** | .instances[]`,
		typ, name,
	***REMOVED***
	results, err := JQ(filter, state***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	ExpectWithOffset(1, results***REMOVED***.To(
		HaveLen(1***REMOVED***,
		"Expected exactly one resource with name type '%s' and name '%s', but found %d",
		typ, name, len(results***REMOVED***,
	***REMOVED***
	return results[0]
}

// Close releases all the resources used by the Terraform runner and removes all
// temporary files and directories.
func (r *TerraformRunner***REMOVED*** Close(***REMOVED*** {
	err := os.RemoveAll(r.dir***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
}
