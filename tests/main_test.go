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

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"

	. "github.com/onsi/ginkgo" // nolint
	. "github.com/onsi/gomega" // nolint
)

func TestProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provider")
}

// terraformBinary is the path of the `terraform` binary that will be used in the tests.
var terraformBinary string

var _ = BeforeSuite(func() {
	var err error

	// Check that the Terraform binary is available in the path:
	terraformBinary, err = exec.LookPath("terraform")
	if err != nil {
		Fail("The 'terraform' binary isn't installed", 1)
	}
})

// TerraformRunner contains the data and logic needed to run Terraform.
type TerraformRunner struct {
	files map[string]string
	env   map[string]string
	vars  map[string]interface{}
	args  []string
}

// TerraformResult contains the result of executing Terraform.
type TerraformResult struct {
	exitCode    int
	stateObject interface{}
}

// NewTerraformRunner creates a new Terraform runner.
func NewTerraformRunner() *TerraformRunner {
	return &TerraformRunner{
		env:   map[string]string{},
		vars:  map[string]interface{}{},
		files: map[string]string{},
	}
}

// File adds a file that will be created in the directory where Terraform will run. The content of
// the file will be generated from the given template and variables.
func (r *TerraformRunner) File(name, template string, vars ...interface{}) *TerraformRunner {
	r.files[name] = sdktesting.EvaluateTemplate(template, vars...)
	return r
}

// Env sets an environment variable that will be used when running Terraform.
func (r *TerraformRunner) Env(name, value string) *TerraformRunner {
	r.env[name] = value
	return r
}

// Var adds a Terraform variable.
func (r *TerraformRunner) Var(name string, value interface{}) *TerraformRunner {
	r.vars[name] = value
	return r
}

// Arg adds a command line argument to Terraform.
func (r *TerraformRunner) Arg(value string) *TerraformRunner {
	r.args = append(r.args, value)
	return r
}

// Args adds a set of command line arguments for the CLI command.
func (r *TerraformRunner) Args(values ...string) *TerraformRunner {
	r.args = append(r.args, values...)
	return r
}

// Run runs the command.
func (r *TerraformRunner) Run(ctx context.Context) *TerraformResult {
	var err error

	// Create a temporary directory for the files so that we don't interfere with the
	// configuration that may already exist for the user running the tests.
	tmpDir, err := ioutil.TempDir("", "ocm-test-*.d")
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	/*
		defer func() {
			err = os.RemoveAll(tmpDir)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
		}()
	*/

	// Create a CLI configuration file that tells Terraform to get plugins from the local
	// directory where `make install` puts them:
	currentDir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())
	projectDir := filepath.Dir(currentDir)
	configPath := filepath.Join(tmpDir, "terraform.rc")
	configText := sdktesting.EvaluateTemplate(
		`
		provider_installation {
		  filesystem_mirror {
		    path    = "{{ .Project }}/.terraform.d/plugins"
		    include = ["localhost/*/*"]
		  }
		}
		`,
		"Project", strings.ReplaceAll(projectDir, "\\", "/"),
	)
	err = ioutil.WriteFile(configPath, []byte(configText), 0400)
	Expect(err).ToNot(HaveOccurred())

	// Create the files specified in the configuration:
	for fileName, fileContent := range r.files {
		filePath := filepath.Join(tmpDir, fileName)
		fileDir := filepath.Dir(filePath)
		err = os.MkdirAll(fileDir, 0700)
		Expect(err).ToNot(HaveOccurred())
		err = ioutil.WriteFile(filePath, []byte(fileContent), 0600)
		Expect(err).ToNot(HaveOccurred())
	}

	// Generate the file containing the Terraform variables:
	varsPath := filepath.Join(tmpDir, "terraform.tfvars.json")
	varsJSON, err := json.Marshal(r.vars)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(varsPath, varsJSON, 0400)
	Expect(err).ToNot(HaveOccurred())

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

	// Add the environment variables specified in the configuration:
	for envName, envValue := range r.env {
		envMap[envName] = envValue
	}

	// Add the environment variable that points to the configuration file:
	envMap["TF_CLI_CONFIG_FILE"] = configPath

	// Enable verbose debug:
	envMap["TF_LOG"] = "DEBUG"

	// Reconstruct the environment list:
	envList := make([]string, 0, len(envMap))
	for name, value := range envMap {
		envList = append(envList, name+"="+value)
	}

	// Run the init command:
	initCmd := exec.Command(terraformBinary, "init")
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

	// Run the command:
	cmd := exec.Command(terraformBinary, r.args...)
	cmd.Env = envList
	cmd.Dir = tmpDir
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	err = cmd.Run()
	switch err.(type) {
	case *exec.ExitError:
		// Nothing, this is a normal situation and the caller is expected to check it using
		// the `ExitCode` method.
	default:
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
	}

	// Read the state:
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	_, err = os.Stat(statePath)
	var stateObject map[string]interface{}
	if err == nil {
		var stateBytes []byte
		stateBytes, err = ioutil.ReadFile(statePath)
		Expect(err).ToNot(HaveOccurred())
		err = json.Unmarshal(stateBytes, &stateObject)
		Expect(err).ToNot(HaveOccurred())
	}

	// Create the result:
	result := &TerraformResult{
		exitCode:    cmd.ProcessState.ExitCode(),
		stateObject: stateObject,
	}

	return result
}

// Apply runs the `apply` command.
func (r *TerraformRunner) Apply(ctx context.Context) *TerraformResult {
	return r.Args("apply", "-auto-approve").Run(ctx)
}

// ExitCode returns the exit code of the CLI command.
func (r *TerraformResult) ExitCode() int {
	return r.exitCode
}

// State returns the result of parsing the JSON content of the `terraform.tfstate` file.
func (r *TerraformResult) State(path string) interface{} {
	return r.stateObject
}

// Resource returns the resource stored in the state with the given type and identifier.
func (r *TerraformResult) Resource(typ, name string) interface{} {
	filter := fmt.Sprintf(
		`.resources[] | select(.type == "%s" and .name == "%s") | .instances[]`,
		typ, name,
	)
	results, err := JQ(filter, r.stateObject)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, results).To(
		HaveLen(1),
		"Expected exactly one resource with name type '%s' and name '%s', but found %d",
		typ, name, len(results),
	)
	return results[0]
}
