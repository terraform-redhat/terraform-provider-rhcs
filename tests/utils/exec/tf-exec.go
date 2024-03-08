package exec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type TerraformExec interface {
	RunTerraformInit(dir string) error
	RunTerraformPlan(dir string, tfVars *TFVars) (string, error)
	RunTerraformApply(dir string, tfVars *TFVars) (string, error)
	RunTerraformDestroy(dir string, deleteTFVars bool) (string, error)
	RunTerraformOutput(dir string) (map[string]interface{}, error)
	RunTerraformState(dir string, subcommand string, tfArgs ...string) (string, error)
	RunTerraformImport(dir string, tfArgs ...string) (string, error)

	ReadTerraformTFVars(dir string) (map[string]string, error)
}

func NewTerraformExec(tfWorkspace string) (TerraformExec, error) {
	return &terraformExec{
		tfWorkspace: tfWorkspace,
	}, nil
}

type terraformExec struct {
	tfWorkspace string
}

func (tfExec *terraformExec) getFullManifestsDir(dir string) string {
	return path.Join(defaultManifestsConfigurationDir, dir)
}

// ************************ TF CMD***********************************

func (tfExec *terraformExec) RunTerraformInit(dir string) (err error) {
	var stdOutput bytes.Buffer
	stdOutput, err = tfExec.runTerraformCommand(dir, "init", "-no-color")
	output := h.Strip(stdOutput.String(), "\n")
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("terraform init failed %s: %s", err.Error(), output)
		return
	}
	Logger.Debugf(output)
	return err
}

func (tfExec *terraformExec) RunTerraformApply(dir string, tfVars *TFVars) (output string, err error) {
	var tfVarsFile string
	tfVarsFile, err = tfExec.RecordTFvarsFile(dir, tfVars)
	if err != nil {
		return
	}

	applyArgs := []string{"apply", "-no-color", "-auto-approve", "-var-file", tfVarsFile}
	var stdOutput bytes.Buffer
	stdOutput, err = tfExec.runTerraformCommand(dir, applyArgs...)
	output = h.Strip(stdOutput.String(), "\n")
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return
	}
	Logger.Debugf(output)
	return
}

func (tfExec *terraformExec) RunTerraformPlan(dir string, tfVars *TFVars) (output string, err error) {
	planArgs := append([]string{"plan", "-no-color"})
	var stdOutput bytes.Buffer
	stdOutput, err = tfExec.runTerraformCommand(dir, planArgs...)
	output = h.Strip(stdOutput.String(), "\n")
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return
	}
	Logger.Debugf(output)
	return
}

func (tfExec *terraformExec) RunTerraformDestroy(dir string, deleteTFVars bool) (output string, err error) {
	tfVarsFile := GrantTFvarsFile(tfExec.getFullManifestsDir(dir))

	destroyArgs := append([]string{"destroy", "-auto-approve", "-no-color", "-var-file", tfVarsFile})
	var stdOutput bytes.Buffer
	stdOutput, err = tfExec.runTerraformCommand(dir, destroyArgs...)
	output = h.Strip(stdOutput.String(), "\n")
	if err != nil {
		err = fmt.Errorf("%s: %s", err.Error(), output)
		Logger.Errorf(err.Error())
		return
	}
	if deleteTFVars {
		tfExec.DeleteTFvarsFile(dir)
	}
	Logger.Debugf(output)
	return
}

func (tfExec *terraformExec) RunTerraformOutput(dir string) (map[string]interface{}, error) {
	outputArgs := []string{"output", "-json"}

	Logger.Infof("Running terraform output against the dir: %s", dir)
	terraformOutput := exec.Command("terraform", outputArgs...)
	terraformOutput.Dir = dir
	output, err := terraformOutput.Output()
	if err != nil {
		return nil, err
	}
	parsedResult := h.Parse(output)
	if err != nil {
		Logger.Errorf(string(output))
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return nil, err
	}
	Logger.Debugf(string(output))
	return parsedResult, err
}

func (tfExec *terraformExec) RunTerraformState(dir, subcommand string, tfArgs ...string) (string, error) {
	stateArgs := append([]string{"state", subcommand}, tfArgs...)
	Logger.Infof("Running terraform state %s against the dir: %s", subcommand, dir)

	terraformState := exec.Command("terraform", stateArgs...)
	terraformState.Dir = dir

	var stdOutput bytes.Buffer
	terraformState.Stdout = &stdOutput
	terraformState.Stderr = &stdOutput

	if err := terraformState.Run(); err != nil {
		errMsg := fmt.Sprintf("%s: %s", err, stdOutput.String())
		Logger.Errorf(errMsg)
		return "", fmt.Errorf("terraform state %s failed: %w", subcommand, errors.New(errMsg))
	}

	output := h.Strip(stdOutput.String(), "\n")
	Logger.Debugf(output)

	return output, nil
}

func (tfExec *terraformExec) RunTerraformImport(dir string, tfArgs ...string) (output string, err error) {
	terraformImport := exec.Command("terraform", append([]string{"import"}, tfArgs...)...)
	terraformImport.Dir = dir

	var stdOutput bytes.Buffer
	terraformImport.Stdout = &stdOutput
	terraformImport.Stderr = &stdOutput

	if err := terraformImport.Run(); err != nil {
		errMsg := fmt.Sprintf("%s: %s", err.Error(), stdOutput.String())
		Logger.Errorf(errMsg)
		return "", errors.New(errMsg)
	}

	output = h.Strip(stdOutput.String(), "\n")
	Logger.Debugf(output)

	return output, nil
}

func (tfExec *terraformExec) runTerraformCommand(dir string, tfArgs ...string) (bytes.Buffer, error) {
	Logger.Infof("Running terraform command against the dir: %s with args %v", dir, tfArgs)

	tfCmd := exec.Command("terraform", tfArgs...)
	tfCmd.Env = append(tfCmd.Env, fmt.Sprintf("TF_WORKSPACE=%s", tfExec.tfWorkspace))
	tfCmd.Dir = tfExec.getFullManifestsDir(dir)

	var stdOutput bytes.Buffer
	tfCmd.Stdout = &stdOutput
	tfCmd.Stderr = &stdOutput

	err := tfCmd.Run()
	return stdOutput, err
}

func (tfExec *terraformExec) RecordTFvarsFile(dir string, tfvars *TFVars) (string, error) {
	tfvarsFile := GrantTFvarsFile(tfExec.getFullManifestsDir(dir))
	iniConn, err := h.IniConnection(tfvarsFile)
	Logger.Infof("Recording tfvars file %s", tfvarsFile)

	if err != nil {
		return "", err
	}
	defer iniConn.SaveTo(tfvarsFile)
	section, err := iniConn.GetSection("")
	if err != nil {
		return "", err
	}
	for k, v := range tfvars.vars {
		section.Key(k).SetValue(v)

	}
	return tfvarsFile, iniConn.SaveTo(tfvarsFile)
}

// delete the recorded TFvarsFile named terraform.tfvars
func (tfExec *terraformExec) DeleteTFvarsFile(dir string) error {
	tfVarsFile := GrantTFvarsFile(tfExec.getFullManifestsDir(dir))
	if _, err := os.Stat(tfVarsFile); err != nil {
		return nil
	}
	Logger.Infof("Deleting tfvars file %s", tfVarsFile)
	return h.DeleteFile(tfVarsFile)
}

// Function to read terraform.tfvars file and return its content as a map
func (tfExec *terraformExec) ReadTerraformTFVars(dir string) (map[string]string, error) {
	filePath := GrantTFvarsFile(tfExec.getFullManifestsDir(dir))
	content, err := os.ReadFile(filePath)
	if err != nil {
		Logger.Errorf("Can't read file %s - not found or could not be fetched", filePath)
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	properties := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove quotes if present
				value = strings.Trim(value, `"`)
				properties[key] = value
			}
		}
	}

	return properties, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
// Args management for TF

type TFVars struct {
	original interface{}
	vars     map[string]string
	// extraCmdLineArgs []string
}

// func NewTFVars(cmdLineArgs ...string) (tfVars *TFVars) {
// 	return &TFVars{
// 		extraCmdLineArgs: cmdLineArgs,
// 	}
// }

func NewTFArgs(obj interface{}) (tfVars *TFVars, err error) {
	tfVars = &TFVars{
		original: obj,
		vars:     make(map[string]string),
	}

	var jsonStruct []byte
	jsonStruct, err = json.Marshal(obj)
	if err != nil {
		return
	}
	parsedMap := map[string]interface{}{}
	err = json.Unmarshal(jsonStruct, &parsedMap)
	if err != nil {
		return
	}

	for k, v := range parsedMap {
		var tfvarV string
		switch v := v.(type) {
		case string:
			tfvarV = fmt.Sprintf(`"%s"`, v)
		case int:
			tfvarV = strconv.Itoa(v)
		case float64:
			tfvarV = fmt.Sprintf("%v", v)
		default:
			var mv []byte
			mv, err = json.Marshal(v)
			if err != nil {
				return
			}
			tfvarV = string(mv)
		}
		tfVars.vars[k] = tfvarV
	}
	return
}

func (args *TFVars) GetVarsAsCmdLineArgs() []string {
	var cmdLineArgs []string
	for k, v := range args.vars {
		arg := fmt.Sprintf("%s=%v", k, v)
		cmdLineArgs = append(cmdLineArgs, "-var", arg)
	}
	return cmdLineArgs
}

// func (args *TFVars) AppendExtraCmdLineArgs(addedArgs ...string) {
// 	args.extraCmdLineArgs = append(args.extraCmdLineArgs, addedArgs...)
// }

// func (args *TFVars) GetExtraCmdLineArgs() []string {
// 	return args.extraCmdLineArgs
// }
