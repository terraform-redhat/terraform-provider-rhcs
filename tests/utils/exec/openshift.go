package exec

***REMOVED***
***REMOVED***
	"strings"
	"time"

***REMOVED***
***REMOVED***

type OcAttributes struct {
	Server          string
	Username        string
	Password        string
	ClusterID       string
	AdditioanlFlags []string
	Timeout         time.Duration
}

func GenerateOCLoginCMD(server string, username string, password string, clusterid string, additioanlFlags ...string***REMOVED*** string {
	cmd := fmt.Sprintf("oc login %s --username %s --password %s",
		server, username, password***REMOVED***
	if len(additioanlFlags***REMOVED*** != 0 {
		cmd = cmd + " " + strings.Join(additioanlFlags, " "***REMOVED***
	}
	return cmd
}

func RetryCMDRun(cmd string, timeout time.Duration***REMOVED*** (string, error***REMOVED*** {
	now := time.Now(***REMOVED***
	var stdout string
	var stderr string
	var err error
	for time.Now(***REMOVED***.Before(now.Add(timeout * time.Minute***REMOVED******REMOVED*** {
		stdout, stderr, err = h.RunCMD(cmd***REMOVED***
		if err == nil {
			return stdout, nil
***REMOVED***
		err = fmt.Errorf(stdout + stderr***REMOVED***
		time.Sleep(time.Minute***REMOVED***
	}
	return "", fmt.Errorf("timeout %d mins for command run %s with error: %s", timeout, cmd, err.Error(***REMOVED******REMOVED***
}

func OcLogin(ocLoginAtter OcAttributes***REMOVED*** error {
	cmd := GenerateOCLoginCMD(ocLoginAtter.Server,
		ocLoginAtter.Username,
		ocLoginAtter.Password,
		ocLoginAtter.ClusterID,
		ocLoginAtter.AdditioanlFlags...***REMOVED***

	errMsg, errStatus := RetryCMDRun(cmd, ocLoginAtter.Timeout***REMOVED***
	if errMsg != "" {
		fmt.Errorf("timeout %d mins for command run %s with error: %s", ocLoginAtter.Timeout, cmd, errStatus.Error(***REMOVED******REMOVED***
	}
	return errStatus
}
