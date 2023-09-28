package exec

import (
	"fmt"
	"strings"
	"time"

	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
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
	for time.Now().Before(now.Add(timeout * time.Minute)) {
		stdout, stderr, err = h.RunCMD(cmd)
		if err == nil {
			return stdout, nil
		}
		err = fmt.Errorf(stdout + stderr)
		time.Sleep(time.Minute)
	}
	return "", fmt.Errorf("timeout %d mins for command run %s with error: %s", timeout, cmd, err.Error())
}

func OcLogin(ocLoginAtter OcAttributes) error {
	cmd := GenerateOCLoginCMD(ocLoginAtter.Server,
		ocLoginAtter.Username,
		ocLoginAtter.Password,
		ocLoginAtter.ClusterID,
		ocLoginAtter.AdditioanlFlags...)

	errMsg, errStatus := RetryCMDRun(cmd, ocLoginAtter.Timeout)
	if errMsg != "" {
		fmt.Errorf("timeout %d mins for command run %s with error: %s", ocLoginAtter.Timeout, cmd, errStatus.Error())
	}
	return errStatus
}
