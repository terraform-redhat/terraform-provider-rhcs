package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// ************************ TF CMD***********************************

func runTerraformInit(ctx context.Context, dir string) error {
	Logger.Infof("Running terraform init against the dir %s", dir)
	terraformInitCmd := exec.Command("terraform", "init", "-no-color")
	terraformInitCmd.Dir = dir
	var stdoutput bytes.Buffer
	terraformInitCmd.Stdout = &stdoutput
	terraformInitCmd.Stderr = &stdoutput
	err := terraformInitCmd.Run()
	output := h.Strip(stdoutput.String(), "\n")
	Logger.Debugf(output)
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("terraform init failed %s: %s", err.Error(), output)
	}

	return err
}

func runTerraformApplyWithArgs(ctx context.Context, dir string, terraformArgs []string) (output string, err error) {
	applyArgs := append([]string{"apply", "-auto-approve", "-no-color"}, terraformArgs...)
	Logger.Infof("Running terraform apply against the dir: %s", dir)
	Logger.Debugf("Running terraform apply against the dir: %s with args %v", dir, terraformArgs)
	terraformApply := exec.Command("terraform", applyArgs...)
	terraformApply.Dir = dir
	var stdoutput bytes.Buffer

	terraformApply.Stdout = &stdoutput
	terraformApply.Stderr = &stdoutput
	err = terraformApply.Run()
	output = h.Strip(stdoutput.String(), "\n")
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return
	}
	Logger.Debugf(output)
	return
}

func runTerraformPlanWithArgs(ctx context.Context, dir string, terraformArgs []string) (output string, err error) {
	planArgs := append([]string{"plan", "-no-color"}, terraformArgs...)
	Logger.Infof("Running terraform plan against the dir: %s ", dir)
	Logger.Debugf("Running terraform plan against the dir: %s with args %v", dir, terraformArgs)
	terraformPlan := exec.Command("terraform", planArgs...)
	terraformPlan.Dir = dir
	var stdoutput bytes.Buffer

	terraformPlan.Stdout = &stdoutput
	terraformPlan.Stderr = &stdoutput
	err = terraformPlan.Run()
	output = h.Strip(stdoutput.String(), "\n")
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return
	}
	Logger.Debugf(output)
	return
}

func runTerraformDestroyWithArgs(ctx context.Context, dir string, terraformArgs []string) (output string, err error) {
	destroyArgs := append([]string{"destroy", "-auto-approve", "-no-color"}, terraformArgs...)
	Logger.Infof("Running terraform destroy against the dir: %s", dir)
	terraformDestroy := exec.Command("terraform", destroyArgs...)
	terraformDestroy.Dir = dir
	var stdoutput bytes.Buffer
	terraformDestroy.Stdout = &stdoutput
	terraformDestroy.Stderr = os.Stderr
	err = terraformDestroy.Run()
	output = h.Strip(stdoutput.String(), "\n")
	if err != nil {
		Logger.Errorf(output)
		err = fmt.Errorf("%s: %s", err.Error(), output)
		return
	}
	Logger.Debugf(output)
	return
}

func runTerraformOutput(ctx context.Context, dir string) (map[string]interface{}, error) {
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

func runTerraformState(dir, subcommand, terraformArgs string) (string, error) {
	stateArgs := []string{"state", subcommand, terraformArgs}
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

func runTerraformImportWithArgs(ctx context.Context, dir string, terraformArgs []string) (output string, err error) {
	Logger.Infof("Running terraform import against the dir: %s", dir)
	Logger.Debugf("Running terraform import against the dir: %s with args %v", dir, terraformArgs)

	terraformImport := exec.CommandContext(ctx, "terraform", append([]string{"import"}, terraformArgs...)...)
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
