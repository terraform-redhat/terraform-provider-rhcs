package exec

***REMOVED***
	"bytes"
	"context"
	"encoding/json"
***REMOVED***
	"os"
	"os/exec"
***REMOVED***

	logging "github.com/sirupsen/logrus"
***REMOVED***
***REMOVED***

// ************************ TF CMD***********************************
func GetLogger(***REMOVED*** *logging.Logger {
	// Create the logger:
	logger := logging.New(***REMOVED***
	return logger
}

var logger *logging.Logger = GetLogger(***REMOVED***

func runTerraformInit(ctx context.Context, dir string***REMOVED*** error {
	logger.Infof("Running terraform init against the dir %s", dir***REMOVED***
	terraformInitCmd := exec.Command("terraform", "init", "-no-color"***REMOVED***
	terraformInitCmd.Dir = dir
	var stdoutput bytes.Buffer
	terraformInitCmd.Stdout = &stdoutput
	terraformInitCmd.Stderr = &stdoutput
	err := terraformInitCmd.Run(***REMOVED***
	output := h.Strip(stdoutput.String(***REMOVED***, "\n"***REMOVED***
	logger.Infof(output***REMOVED***
	if err != nil {
		logger.Errorf(output***REMOVED***
		err = fmt.Errorf("terraform init failed %s: %s", err.Error(***REMOVED***, output***REMOVED***
	}

	return err
}

func runTerraformApplyWithArgs(ctx context.Context, dir string, terraformArgs []string***REMOVED*** (output string, err error***REMOVED*** {
	applyArgs := append([]string{"apply", "-auto-approve", "-no-color"}, terraformArgs...***REMOVED***
	logger.Infof("Running terraform apply against the dir: %s with args %v", dir, terraformArgs***REMOVED***
	terraformApply := exec.Command("terraform", applyArgs...***REMOVED***
	terraformApply.Dir = dir
	var stdoutput bytes.Buffer
	// var stderroutput bytes.Buffer
	terraformApply.Stdout = &stdoutput
	terraformApply.Stderr = &stdoutput
	err = terraformApply.Run(***REMOVED***
	output = h.Strip(stdoutput.String(***REMOVED***, "\n"***REMOVED***
	if err != nil {
		logger.Errorf(output***REMOVED***
		err = fmt.Errorf("%s: %s", err.Error(***REMOVED***, output***REMOVED***
		return
	}
	logger.Infof(output***REMOVED***
	return
}
func runTerraformDestroyWithArgs(ctx context.Context, dir string, terraformArgs []string***REMOVED*** error {
	destroyArgs := append([]string{"destroy", "-auto-approve", "-no-color"}, terraformArgs...***REMOVED***
	logger.Infof("Running terraform destroy against the dir: %s", dir***REMOVED***
	terraformApply := exec.Command("terraform", destroyArgs...***REMOVED***
	terraformApply.Dir = dir
	terraformApply.Stdout = os.Stdout
	terraformApply.Stderr = os.Stderr
	err := terraformApply.Run(***REMOVED***
	return err
}
func runTerraformOutput(ctx context.Context, dir string***REMOVED*** (map[string]interface{}, error***REMOVED*** {
	outputArgs := []string{"output", "-json"}
	logger.Infof("Running terraform output against the dir: %s", dir***REMOVED***
	terraformOutput := exec.Command("terraform", outputArgs...***REMOVED***
	terraformOutput.Dir = dir
	output, err := terraformOutput.Output(***REMOVED***
	if err != nil {
		return nil, err
	}
	parsedResult := h.Parse(output***REMOVED***
	return parsedResult, err
}

func combineArgs(varAgrs map[string]interface{}, abArgs ...string***REMOVED*** []string {

	args := []string{}
	for k, v := range varAgrs {
		var argV interface{}
		switch v.(type***REMOVED*** {
		case string:
			argV = v
		case int:
			argV = v
		case float64:
			argV = v
		default:
			mv, _ := json.Marshal(v***REMOVED***
			argV = string(mv***REMOVED***

***REMOVED***
		arg := fmt.Sprintf("%s=%v", k, argV***REMOVED***
		args = append(args, "-var"***REMOVED***
		args = append(args, arg***REMOVED***
	}
	args = append(args, abArgs...***REMOVED***
	return args
}

func combineStructArgs(argObj interface{}, abArgs ...string***REMOVED*** []string {
	parambytes, _ := json.Marshal(argObj***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	return combineArgs(args, abArgs...***REMOVED***
}

func CleanTFTempFiles(providerDir string***REMOVED*** error {
	tempList := []string{}
	for _, temp := range tempList {
		tempPath := path.Join(providerDir, temp***REMOVED***
		err := os.RemoveAll(tempPath***REMOVED***
		if err != nil {
			return err
***REMOVED***
	}
	return nil
}
