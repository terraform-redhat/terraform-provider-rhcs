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

package tests

***REMOVED***
	"context"
	"encoding/json"
***REMOVED***
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"

	. "github.com/onsi/ginkgo" // nolint
***REMOVED*** // nolint
***REMOVED***

func TestProvider(t *testing.T***REMOVED*** {
	RegisterFailHandler(Fail***REMOVED***
	RunSpecs(t, "Provider"***REMOVED***
}

// terraformBinary is the path of the `terraform` binary that will be used in the tests.
var terraformBinary string

var _ = BeforeSuite(func(***REMOVED*** {
	var err error

	// Check that the Terraform binary is available in the path:
	terraformBinary, err = exec.LookPath("terraform"***REMOVED***
	if err != nil {
		Fail("The 'terraform' binary isn't installed", 1***REMOVED***
	}
}***REMOVED***

// TerraformRunner contains the data and logic needed to run Terraform.
type TerraformRunner struct {
	files map[string]string
	env   map[string]string
	vars  map[string]interface{}
	args  []string
}

// TerraformResult contains the result of executing Terraform.
type TerraformResult struct {
	exitCode int
}

// NewTerraformRunner creates a new Terraform runner.
func NewTerraformRunner(***REMOVED*** *TerraformRunner {
	return &TerraformRunner{
		env:   map[string]string{},
		vars:  map[string]interface{}{},
		files: map[string]string{},
	}
}

// File adds a file that will be created in the directory where Terraform will run. The content of
// the file will be generated from the given template and variables.
func (r *TerraformRunner***REMOVED*** File(name, template string, vars ...interface{}***REMOVED*** *TerraformRunner {
	r.files[name] = sdktesting.EvaluateTemplate(template, vars...***REMOVED***
	return r
}

// Env sets an environment variable that will be used when running Terraform.
func (r *TerraformRunner***REMOVED*** Env(name, value string***REMOVED*** *TerraformRunner {
	r.env[name] = value
	return r
}

// Var adds a Terraform variable.
func (r *TerraformRunner***REMOVED*** Var(name string, value interface{}***REMOVED*** *TerraformRunner {
	r.vars[name] = value
	return r
}

// Arg adds a command line argument to Terraform.
func (r *TerraformRunner***REMOVED*** Arg(value string***REMOVED*** *TerraformRunner {
	r.args = append(r.args, value***REMOVED***
	return r
}

// Args adds a set of command line arguments for the CLI command.
func (r *TerraformRunner***REMOVED*** Args(values ...string***REMOVED*** *TerraformRunner {
	r.args = append(r.args, values...***REMOVED***
	return r
}

// Run runs the command.
func (r *TerraformRunner***REMOVED*** Run(ctx context.Context***REMOVED*** *TerraformResult {
	var err error

	// Create a temporary directory for the files so that we don't interfere with the
	// configuration that may already exist for the user running the tests.
	tmpDir, err := ioutil.TempDir("", "ocm-test-*.d"***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	defer func(***REMOVED*** {
		err = os.RemoveAll(tmpDir***REMOVED***
		ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}(***REMOVED***

	// Create a CLI configuration file that tells Terraform to get plugins from the local
	// directory where `make install` puts them:
	currentDir, err := os.Getwd(***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	projectDir := filepath.Dir(currentDir***REMOVED***
	configPath := filepath.Join(tmpDir, "terraform.rc"***REMOVED***
	configText := sdktesting.EvaluateTemplate(
		`
		provider_installation {
		  filesystem_mirror {
		    path    = "{{ .Project }}/.terraform.d/plugins"
		    include = ["localhost/*/*"]
		  }
***REMOVED***
		`,
		"Project", strings.ReplaceAll(projectDir, "\\", "/"***REMOVED***,
	***REMOVED***
	err = ioutil.WriteFile(configPath, []byte(configText***REMOVED***, 0400***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

	// Create the files specified in the configuration:
	for fileName, fileContent := range r.files {
		filePath := filepath.Join(tmpDir, fileName***REMOVED***
		fileDir := filepath.Dir(filePath***REMOVED***
		err = os.MkdirAll(fileDir, 0700***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		err = ioutil.WriteFile(filePath, []byte(fileContent***REMOVED***, 0600***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}

	// Generate the file containing the Terraform variables:
	varsPath := filepath.Join(tmpDir, "terraform.tfvars.json"***REMOVED***
	varsJSON, err := json.Marshal(r.vars***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	err = ioutil.WriteFile(varsPath, varsJSON, 0400***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

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

	// Add the environment variables specified in the configuration:
	for envName, envValue := range r.env {
		envMap[envName] = envValue
	}

	// Add the environment variable that points to the configuration file:
	envMap["TF_CLI_CONFIG_FILE"] = configPath

	// Enable verbose debug:
	envMap["TF_LOG"] = "DEBUG"

	// Reconstruct the environment list:
	envList := make([]string, 0, len(envMap***REMOVED******REMOVED***
	for name, value := range envMap {
		envList = append(envList, name+"="+value***REMOVED***
	}

	// Run the init command:
	initCmd := exec.Command(terraformBinary, "init"***REMOVED***
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

	// Run the command:
	cmd := exec.Command(terraformBinary, r.args...***REMOVED***
	cmd.Env = envList
	cmd.Dir = tmpDir
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	err = cmd.Run(***REMOVED***
	switch err.(type***REMOVED*** {
	case *exec.ExitError:
		// Nothing, this is a normal situation and the caller is expected to check it using
		// the `ExitCode` method.
	default:
		ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}

	// Create the result:
	result := &TerraformResult{
		exitCode: cmd.ProcessState.ExitCode(***REMOVED***,
	}

	return result
}

// Apply runs the `apply` command.
func (r *TerraformRunner***REMOVED*** Apply(ctx context.Context***REMOVED*** *TerraformResult {
	return r.Args("apply", "-auto-approve"***REMOVED***.Run(ctx***REMOVED***
}

// ExitCode returns the exit code of the CLI command.
func (r *TerraformResult***REMOVED*** ExitCode(***REMOVED*** int {
	return r.exitCode
}
