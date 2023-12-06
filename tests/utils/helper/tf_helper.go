package helper

import (
	"encoding/json"
	"fmt"
	"os"

	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

// Get the resoources state from the terraform.tfstate file by resource type and name
func GetResource(manifestDir string, resourceType string, resoureName string) (interface{}, error) {
	// Check if there is a terraform.tfstate file in the manifest directory
	if _, err := os.Stat(manifestDir + "/terraform.tfstate"); err == nil {
		// Read the terraform.tfstate file
		data, err := os.ReadFile(manifestDir + "/terraform.tfstate")
		if err != nil {
			return nil, fmt.Errorf("failed to readFile %s folder,%v", manifestDir+"/terraform.tfstate", err)
		}
		// Unmarshal the data from the terraform.tfstate file
		var state map[string]interface{}
		err = json.Unmarshal(data, &state)
		if err != nil {
			return nil, fmt.Errorf("failed to Unmarshal state %v", err)
		}
		//Find resource by resource type and resource name
		for _, resource := range state["resources"].([]interface{}) {
			if DigString(resource, "type") == resourceType && DigString(resource, "name") == resoureName {
				return resource, err
			}
		}

		return nil, fmt.Errorf("no resource named %s of type %s is found", resoureName, resourceType)

	}
	return nil, fmt.Errorf("terraform.tfstate file doesn't exist in %s folder", manifestDir)
}

// Make sure the default machinepool imported by checking if there is terraform.tfstate in DefaultMachinePoolDir
func MakeSureDefaultMachinePoolImported() (imported bool, err error) {
	_, err = GetResource(con.DefaultMachinePoolDir, "rhcs_machine_pool", "dmp")
	if err != nil {
		return false, err
	}
	return true, nil
}
