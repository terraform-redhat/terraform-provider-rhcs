package openshift

import (
	"fmt"
	"strings"
	"time"

	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type OcAttributes struct {
	Server          string
	Username        string
	Password        string
	ClusterID       string
	AdditioanlFlags []string
	Timeout         time.Duration
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
		stdout, stderr, err = h.RunCMD(cmd)
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
		ocLoginAtter.AdditioanlFlags...)

	output, err := RetryCMDRun(cmd, ocLoginAtter.Timeout)
	return output, err

}
