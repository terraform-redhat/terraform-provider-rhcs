package exec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type TerraformExecutor interface {
	RunTerraformInit() (string, error)
	RunTerraformPlan(argObj interface{}) (string, error)
	RunTerraformApply(argObj interface{}) (string, error)
	RunTerraformDestroy() (string, error)
	RunTerraformOutput() (string, error)
	RunTerraformOutputIntoObject(obj any) error
	RunTerraformState(subcommand string, options ...string) (string, error)
	GetStateResource(resourceType string, resoureName string) (interface{}, error)
	RunTerraformImport(importArgs ...string) (string, error)

	ReadTerraformVars(obj interface{}) error
	WriteTerraformVars(obj interface{}) error
	DeleteTerraformVars() error
}

type terraformExecutorContext struct {
	manifestsDir string
	tfWorkspace  string
}

func NewTerraformExecutor(tfWorkspace string, manifestsDir string) TerraformExecutor {
	return &terraformExecutorContext{
		manifestsDir: manifestsDir,
		tfWorkspace:  tfWorkspace,
	}
}

// ************************ TF CMD***********************************
func (ctx *terraformExecutorContext) runTerraformCommand(tfCmd string, cmdFlags ...string) (string, error) {
	Logger.Infof("Running terraform %s against the dir %s", tfCmd, ctx.manifestsDir)
	cmd, flags := getTerraformCommand(tfCmd, cmdFlags...)
	Logger.Debugf("Running terraform command: %v", flags)
	return ctx.execCommand(cmd, flags)
}

func getTerraformCommand(tfCmd string, cmdFlags ...string) (string, []string) {
	flags := []string{tfCmd}
	flags = append(flags, cmdFlags...)
	return "terraform", flags
}

func (ctx *terraformExecutorContext) execCommand(cmd string, flags []string) (output string, err error) {
	finalCmd := exec.Command(cmd, flags...)
	if ctx.tfWorkspace != "" {
		finalCmd.Env = append(os.Environ(), fmt.Sprintf("TF_WORKSPACE=%s", ctx.tfWorkspace))
	}
	finalCmd.Dir = ctx.manifestsDir
	var stdoutput bytes.Buffer
	finalCmd.Stdout = &stdoutput
	finalCmd.Stderr = &stdoutput
	err = finalCmd.Run()
	output = helper.Strip(stdoutput.String(), "\n")
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return
	}
	Logger.Debugf(output)
	return
}

func (ctx *terraformExecutorContext) RunTerraformInit() (string, error) {
	return ctx.runTerraformCommand("init", "-no-color")
}

func (ctx *terraformExecutorContext) RunTerraformPlan(argObj interface{}) (output string, err error) {
	tempFile, err := WriteTemporaryTFVarsFile(argObj)
	if err != nil {
		return "", err
	}
	defer DeleteTFvarsFile(tempFile)
	planArgs := append([]string{"-no-color"}, "-var-file", tempFile)
	return ctx.runTerraformCommand("plan", planArgs...)
}

func (ctx *terraformExecutorContext) RunTerraformApply(argObj interface{}) (string, error) {
	tempFile, err := WriteTemporaryTFVarsFile(argObj)
	if err != nil {
		return "", err
	}
	defer DeleteTFvarsFile(tempFile)

	output, err := ctx.runTerraformCommand("apply", "-auto-approve", "-no-color", "-var-file", tempFile)
	if err == nil {
		// If it works, tf vars are officially recorded
		err = ctx.WriteTerraformVars(argObj)
	}
	return output, err
}

func (ctx *terraformExecutorContext) RunTerraformDestroy() (output string, err error) {
	output, err = ctx.runTerraformCommand("destroy", "-auto-approve", "-no-color", "-var-file", ctx.grantTFvarsFile())
	if err == nil {
		ctx.DeleteTerraformVars()
	}
	return
}

func (ctx *terraformExecutorContext) RunTerraformOutput() (string, error) {
	outputArgs := []string{"-json"}
	cmd, flags := getTerraformCommand("output", outputArgs...)
	tfCmd := strings.Join(append([]string{cmd}, flags...), " ")
	parseTFOutputCmd := "jq 'with_entries(.value |= .value)'" // Needed to get only values from TF output
	finalCmd := strings.Join([]string{tfCmd, parseTFOutputCmd}, " | ")
	return ctx.execCommand("bash", []string{"-c", finalCmd})
}

func (ctx *terraformExecutorContext) RunTerraformOutputIntoObject(obj any) error {
	output, err := ctx.RunTerraformOutput()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(output), obj)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *terraformExecutorContext) RunTerraformState(subcommand string, options ...string) (string, error) {
	stateArgs := []string{subcommand}
	if len(options) > 0 {
		stateArgs = append(stateArgs, options...)
	}
	return ctx.runTerraformCommand("state", stateArgs...)
}

func (ctx *terraformExecutorContext) RunTerraformImport(importArgs ...string) (output string, err error) {
	return ctx.runTerraformCommand("import", importArgs...)
}

func (ctx *terraformExecutorContext) WriteTerraformVars(obj interface{}) error {
	Logger.Infof("Write tfvars file")
	return WriteTFvarsFile(obj, ctx.grantTFvarsFile())
}

func WriteTFvarsFile(obj interface{}, tfvarsFilePath string) error {
	tfVarsFile, err := os.Create(tfvarsFilePath)
	if err != nil {
		return err
	}
	defer tfVarsFile.Close()
	Logger.Debugf("Recording tfvars file %s", tfvarsFilePath)

	hclFile := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(obj, hclFile.Body())

	var buff bytes.Buffer
	hclFile.WriteTo(&buff)
	Logger.Debugf("Recording tfvars values %v", buff.String())

	_, err = hclFile.WriteTo(tfVarsFile)
	return err
}

func WriteTemporaryTFVarsFile(obj interface{}) (string, error) {
	file, err := createVarsTempFile()
	if err != nil {
		return "", err
	}
	return file, WriteTFvarsFile(obj, file)
}

// Function to read parse tf vars in an object
// See https://hclguide.readthedocs.io/en/latest/go_decoding_gohcl.html
func (ctx *terraformExecutorContext) ReadTerraformVars(obj interface{}) error {
	return ReadTerraformVarsFile(ctx.grantTFvarsFile(), obj)
}

func ReadTerraformVarsFile(filePath string, obj interface{}) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	Logger.Debugf("Reading tfvars file %s", filePath)
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filePath)
	if diags.HasErrors() {
		return errors.Join(diags.Errs()...)
	}

	diags = gohcl.DecodeBody(f.Body, nil, obj)
	if diags.HasErrors() {
		return errors.Join(diags.Errs()...)
	}
	return nil
}

func (ctx *terraformExecutorContext) DeleteTerraformVars() error {
	Logger.Info("Deleting tfvars file")
	return DeleteTFvarsFile(ctx.grantTFvarsFile())
}

func DeleteTFvarsFile(tfVarsFile string) error {
	if _, err := os.Stat(tfVarsFile); err != nil {
		return nil
	}
	Logger.Debugf("Deleting tfvars file %s", tfVarsFile)
	return helper.DeleteFile(tfVarsFile)
}

func createVarsTempFile() (string, error) {
	f, err := os.CreateTemp("", "tfvars-")
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (ctx *terraformExecutorContext) grantTFvarsFile() string {
	// We don't name it `terraform.tfvars` as that one is load automatically
	// See https://developer.hashicorp.com/terraform/language/values/variables#variable-definition-precedence
	// And we don't want to load them when applying new values
	wk := "e2e"
	if ctx.tfWorkspace != "" {
		wk = ctx.tfWorkspace
	}
	return path.Join(ctx.manifestsDir, fmt.Sprintf("terraform.%s.tfvars", wk))
}

func (ctx *terraformExecutorContext) grantTFstateFile() string {
	return path.Join(ctx.manifestsDir, "terraform.tfstate")
}

// Get the resoources state from the terraform.tfstate file by resource type and name
func (ctx *terraformExecutorContext) GetStateResource(resourceType string, resoureName string) (interface{}, error) {
	// Check if there is a terraform.tfstate file in the manifest directory
	stateFile := ctx.grantTFstateFile()
	if _, err := os.Stat(stateFile); err == nil {
		// Read the terraform.tfstate file
		data, err := os.ReadFile(stateFile)
		if err != nil {
			return nil, fmt.Errorf("failed to readFile %s folder,%v", stateFile, err)
		}
		// Unmarshal the data from the terraform.tfstate file
		var state map[string]interface{}
		err = json.Unmarshal(data, &state)
		if err != nil {
			return nil, fmt.Errorf("failed to Unmarshal state %v", err)
		}
		//Find resource by resource type and resource name
		for _, resource := range state["resources"].([]interface{}) {
			if helper.DigString(resource, "type") == resourceType && helper.DigString(resource, "name") == resoureName && resource != nil {
				return resource, err
			}
		}

		return nil, fmt.Errorf("no resource named %s of type %s is found", resoureName, resourceType)

	}
	return nil, fmt.Errorf("terraform.tfstate file doesn't exist in %s folder", ctx.manifestsDir)
}
