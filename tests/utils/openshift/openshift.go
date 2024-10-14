package openshift

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	client "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type OcAttributes struct {
	Server          string
	Username        string
	Password        string
	ClusterID       string
	AdditionalFlags []string
	Timeout         time.Duration
}

// Pod struct is struct that contains info
type Pod struct {
	Name      string `json:"name,omitempty"`
	IP        string `json:"ip,omitempty"`
	Status    string `json:"status,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	HostIP    string `json:"hostIP,omitempty"`
}

func GenerateOCLoginCMD(server string, username string, password string, clusterid string, additioanlFlags ...string) string {
	cmd := fmt.Sprintf("oc login %s --username %s --password %s",
		server, username, password)
	if len(additioanlFlags) != 0 {
		cmd = cmd + " " + strings.Join(additioanlFlags, " ")
	}
	return cmd
}

func RetryCMDRun(cmd string, timeout time.Duration) (string, error) {
	now := time.Now()
	var stdout string
	var stderr string
	var err error
	Logger.Infof("Retrying command %s in %d mins", cmd, timeout)
	for time.Now().Before(now.Add(timeout * time.Minute)) {
		stdout, stderr, err = helper.RunCMD(cmd)
		if err == nil {
			Logger.Debugf("Run command %s successffly", cmd)
			return stdout, nil
		}
		err = fmt.Errorf(stdout + stderr)
		time.Sleep(time.Minute)
	}
	return "", fmt.Errorf("timeout %d mins for command run %s with error: %s", timeout, cmd, err.Error())
}

func OcLogin(ocLoginAtter OcAttributes) (string, error) {
	cmd := GenerateOCLoginCMD(ocLoginAtter.Server,
		ocLoginAtter.Username,
		ocLoginAtter.Password,
		ocLoginAtter.ClusterID,
		ocLoginAtter.AdditionalFlags...)

	output, err := RetryCMDRun(cmd, ocLoginAtter.Timeout)
	return output, err

}

func WaitForOperatorsToBeReady(connection *client.Connection, clusterID string, timeout int) error {
	// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
	timeoutMin := time.Duration(timeout)
	console, err := NewConsole(clusterID, connection)
	if err != nil {
		Logger.Warnf("Got error %s when config the openshift console. Return without waiting for operators ready", err.Error())
		return err
	}
	_, err = RetryCMDRun(fmt.Sprintf("oc wait clusteroperators --all --for=condition=Progressing=false --kubeconfig %s --timeout %dm", console.KubePath, timeout), timeoutMin)
	return err
}

func RestartMUOPods(connection *client.Connection, clusterID string) error {
	MUONamespace := "openshift-managed-upgrade-operator"
	console, err := NewConsole(clusterID, connection)
	if err != nil {
		return err
	}
	pods, err := console.GetPods(MUONamespace)
	for _, pod := range pods {
		cmd := fmt.Sprintf("oc delete pod/%s -n %s --kubeconfig %s", pod.Name, MUONamespace, console.KubePath)
		_, _, err = helper.RunCMD(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// WaitForUpgradePolicyToState will time out after <timeout> minutes
// Be careful for state completed. Make sure the automatic policy is in status of other status rather than pending
func WaitForUpgradePolicyToState(connection *client.Connection, clusterID string, policyID string, state string, timeout int) error {
	fmt.Println("Going to wait upgrade to status ", state)
	startTime := time.Now()
	resp, err := cms.RetrieveUpgradePolicies(connection, clusterID, policyID)
	if err != nil {
		return err
	}
	if resp.Status() != http.StatusOK {
		return fmt.Errorf(">>> Error happened when retrieve policy detail: %s", resp.Error())
	}
	scheduleType := resp.Body().ScheduleType()

	for time.Now().Before(startTime.Add(time.Duration(timeout) * time.Minute)) {
		stateResp, _ := cms.GetUpgradePolicyState(connection, clusterID, policyID)

		switch state {
		case constants.Completed:
			if scheduleType == constants.ManualScheduleType {
				if stateResp.Status() == http.StatusNotFound {
					return nil
				} else if resp.Status() != http.StatusOK {
					return fmt.Errorf(">>> Got response %s when retrieve the policy state: %s", resp.Error(), state)
				}
			} else {
				if stateResp.Status() != http.StatusOK {
					return fmt.Errorf(">>> Got response %s when retrieve the policy state: %s", resp.Error(), state)
				}
				if stateResp.Body().Value() == constants.Pending {
					return nil
				}
			}

		default:
			if resp.Status() != http.StatusOK {
				return fmt.Errorf(">>> Got response %s when retrieve the policy state: %s", resp.Error(), state)
			}
			if string(stateResp.Body().Value()) == state {
				return nil
			}

		}

		time.Sleep(1 * time.Minute)

	}
	return fmt.Errorf("ERROR!Timeout after %d minutes to wait for the policy %s into status %s of cluster %s",
		timeout, policyID, state, clusterID)

}

// WaitForControlPlaneUpgradePolicyToState will time out after <timeout> minutes
// Be careful for state completed. Make sure the automatic policy is in status of other status rather than pending
func WaitForControlPlaneUpgradePolicyToState(connection *client.Connection, clusterID string, policyID string, state v1.UpgradePolicyStateValue, timeout int) error {
	fmt.Println("Going to wait upgrade to status ", state)
	startTime := time.Now()
	resp, err := cms.RetrieveControlPlaneUpgradePolicy(connection, clusterID, policyID)
	if err != nil {
		return err
	}
	if resp.Status() != http.StatusOK {
		return fmt.Errorf(">>> Error happened when retrieve policy detail: %s", resp.Error())
	}
	scheduleType := resp.Body().ScheduleType()

	for time.Now().Before(startTime.Add(time.Duration(timeout) * time.Minute)) {
		resp, _ := cms.RetrieveControlPlaneUpgradePolicy(connection, clusterID, policyID)

		switch state {
		case v1.UpgradePolicyStateValueCompleted:
			if scheduleType == constants.ManualScheduleType {
				if resp.Status() == http.StatusNotFound {
					return nil
				} else if resp.Status() != http.StatusOK {
					return fmt.Errorf(">>> Got response %s when retrieve the policy state: %s", resp.Error(), state)
				}
			} else {
				if resp.Status() != http.StatusOK {
					return fmt.Errorf(">>> Got response %s when retrieve the policy state: %s", resp.Error(), state)
				}
				if resp.Body().State().Value() == v1.UpgradePolicyStateValuePending {
					return nil
				}
			}

		default:
			if resp.Status() != http.StatusOK {
				return fmt.Errorf(">>> Got response %s when retrieve the policy state: %s", resp.Error(), state)
			}
			if resp.Body().State().Value() == state {
				return nil
			}

		}

		time.Sleep(1 * time.Minute)

	}
	return fmt.Errorf("ERROR!Timeout after %d minutes to wait for the policy %s into status %s of cluster %s",
		timeout, policyID, state, clusterID)

}

func WaitClassicClusterUpgradeFinished(connection *client.Connection, clusterID string) error {
	Logger.Infof("Get the automatic policy created for the cluster upgrade")
	policyIDs, err := cms.ListUpgradePolicies(cms.RHCSConnection, clusterID)
	if err != nil {
		return err
	}
	policyID := policyIDs.Items().Get(0).ID()

	Logger.Infof("Wait the policy to be scheduled")
	err = WaitForUpgradePolicyToState(cms.RHCSConnection, clusterID, policyID, constants.Scheduled, 4)
	if err != nil {
		return fmt.Errorf("Policy %s not moved to state %s in 2 minutes with the error: %s", constants.Scheduled, policyID, err.Error())
	}

	Logger.Infof("Restart the MUO operator pod to make the policy synced")
	err = RestartMUOPods(cms.RHCSConnection, clusterID)
	if err != nil {
		return err
	}
	Logger.Infof("Watch for the upgrade Started in 1 hour")
	err = WaitForUpgradePolicyToState(cms.RHCSConnection, clusterID, policyID, constants.Started, 60)
	if err != nil {
		return fmt.Errorf("Policy %s not moved to state %s in 1 hour with the error: %s", constants.Started, policyID, err.Error())
	}
	Logger.Infof("Watch for the upgrade finished in 2 hours")
	err = WaitForUpgradePolicyToState(cms.RHCSConnection, clusterID, policyID, constants.Completed, 2*60)
	if err != nil {
		return fmt.Errorf("Policy %s not moved to state %s in 2 hour with the error: %s", constants.Completed, policyID, err.Error())
	}
	return nil
}

func WaitHCPClusterUpgradeFinished(connection *client.Connection, clusterID string) error {
	Logger.Infof("Get the automatic policy created for the cluster upgrade")
	policyIDs, err := cms.ListControlPlaneUpgradePolicies(cms.RHCSConnection, clusterID)
	if err != nil {
		return err
	}
	policyID := policyIDs.Items().Get(0).ID()
	Logger.Infof("Got policy ID %s", policyID)

	Logger.Infof("Wait the policy to be scheduled")
	err = WaitForControlPlaneUpgradePolicyToState(cms.RHCSConnection, clusterID, policyID, v1.UpgradePolicyStateValueScheduled, 4)
	if err != nil {
		return fmt.Errorf("Policy %s not moved to state %s in 2 minutes with the error: %s", v1.UpgradePolicyStateValueScheduled, policyID, err.Error())
	}

	Logger.Infof("Watch for the upgrade Started in 1 hour")
	err = WaitForControlPlaneUpgradePolicyToState(cms.RHCSConnection, clusterID, policyID, v1.UpgradePolicyStateValueStarted, 60)
	if err != nil {
		return fmt.Errorf("Policy %s not moved to state %s in 1 hour with the error: %s", v1.UpgradePolicyStateValueStarted, policyID, err.Error())
	}
	Logger.Infof("Watch for the upgrade finished in 2 hours")
	err = WaitForControlPlaneUpgradePolicyToState(cms.RHCSConnection, clusterID, policyID, v1.UpgradePolicyStateValueCompleted, 2*60)
	if err != nil {
		return fmt.Errorf("Policy %s not moved to state %s in 2 hour with the error: %s", v1.UpgradePolicyStateValueCompleted, policyID, err.Error())
	}
	return nil
}

// will return [map["NAME":"ip-10-0-130-210.us-east-2.compute.internal","STATUS":"Ready","ROLES":"worker"...]]
func FigureStdout(stdout string, columns map[string][]interface{}) (result []map[string]interface{}, err error) {
	items := helper.DigArray(helper.Parse([]byte(stdout)), "items")
	for _, item := range items {
		newMap := map[string]interface{}{}
		for key, pattern := range columns {
			newMap[key] = helper.Dig(item, pattern)
		}
		result = append(result, newMap)
	}
	return
}
