package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"

	logging "github.com/sirupsen/logrus"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

// ************************ TF CMD***********************************
func GetLogger() *logging.Logger {
	// Create the logger:
	logger := logging.New()
	return logger
}

var logger *logging.Logger = GetLogger()

func runTerraformInit(ctx context.Context, dir string) error {
	logger.Infof("Running terraform init against the dir %s", dir)
	terraformInitCmd := exec.Command("terraform", "init", "-no-color")
	terraformInitCmd.Dir = dir
	var stdoutput bytes.Buffer
	terraformInitCmd.Stdout = &stdoutput
	terraformInitCmd.Stderr = &stdoutput
	err := terraformInitCmd.Run()
	output := h.Strip(stdoutput.String(), "\n")
	logger.Infof(output)
	if err != nil {
		logger.Errorf(output)
		err = fmt.Errorf("terraform init failed %s: %s", err.Error(), output)
	}

	return err
}

func runTerraformApplyWithArgs(ctx context.Context, dir string, terraformArgs []string) (output string, err error) {
	applyArgs := append([]string{"apply", "-auto-approve", "-no-color"}, terraformArgs...)
	logger.Infof("Running terraform apply against the dir: %s with args %v", dir, terraformArgs)
	terraformApply := exec.Command("terraform", applyArgs...)
	terraformApply.Dir = dir
	var stdoutput bytes.Buffer
	// var stderroutput bytes.Buffer
	terraformApply.Stdout = &stdoutput
	terraformApply.Stderr = &stdoutput
	err = terraformApply.Run()
	output = h.Strip(stdoutput.String(), "\n")
	if err != nil {
		logger.Errorf(output)
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return
	}
	logger.Infof(output)
	return
}
func runTerraformDestroyWithArgs(ctx context.Context, dir string, terraformArgs []string) error {
	destroyArgs := append([]string{"destroy", "-auto-approve", "-no-color"}, terraformArgs...)
	logger.Infof("Running terraform destroy against the dir: %s", dir)
	terraformApply := exec.Command("terraform", destroyArgs...)
	terraformApply.Dir = dir
	terraformApply.Stdout = os.Stdout
	terraformApply.Stderr = os.Stderr
	err := terraformApply.Run()
	return err
}
func runTerraformOutput(ctx context.Context, dir string) (map[string]interface{}, error) {
	outputArgs := []string{"output", "-json"}
	logger.Infof("Running terraform output against the dir: %s", dir)
	terraformOutput := exec.Command("terraform", outputArgs...)
	terraformOutput.Dir = dir
	output, err := terraformOutput.Output()
	if err != nil {
		return nil, err
	}
	parsedResult := h.Parse(output)
	return parsedResult, err
}

func combineArgs(varAgrs map[string]interface{}, abArgs ...string) []string {

	args := []string{}
	for k, v := range varAgrs {
		var argV interface{}
		switch v.(type) {
		case string:
			argV = v
		case int:
			argV = v
		case float64:
			argV = v
		default:
			mv, _ := json.Marshal(v)
			argV = string(mv)

		}
		arg := fmt.Sprintf("%s=%v", k, argV)
		args = append(args, "-var")
		args = append(args, arg)
	}
	args = append(args, abArgs...)
	return args
}

func combineStructArgs(argObj interface{}, abArgs ...string) []string {
	parambytes, _ := json.Marshal(argObj)
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args)
	return combineArgs(args, abArgs...)
}

func CleanTFTempFiles(providerDir string) error {
	tempList := []string{}
	for _, temp := range tempList {
		tempPath := path.Join(providerDir, temp)
		err := os.RemoveAll(tempPath)
		if err != nil {
			return err
		}
	}
	return nil
}
