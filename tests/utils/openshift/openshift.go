package openshift

***REMOVED***
***REMOVED***
	"strings"
	"time"

***REMOVED***
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
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
	Logger.Infof("Retrying command %s in %d mins", cmd, timeout***REMOVED***
	for time.Now(***REMOVED***.Before(now.Add(timeout * time.Minute***REMOVED******REMOVED*** {
		stdout, stderr, err = h.RunCMD(cmd***REMOVED***
		if err == nil {
			Logger.Debugf("Run command %s successffly", cmd***REMOVED***
			return stdout, nil
***REMOVED***
		err = fmt.Errorf(stdout + stderr***REMOVED***
		time.Sleep(time.Minute***REMOVED***
	}
	return "", fmt.Errorf("timeout %d mins for command run %s with error: %s", timeout, cmd, err.Error(***REMOVED******REMOVED***
}

func OcLogin(ocLoginAtter OcAttributes***REMOVED*** (string, error***REMOVED*** {
	cmd := GenerateOCLoginCMD(ocLoginAtter.Server,
		ocLoginAtter.Username,
		ocLoginAtter.Password,
		ocLoginAtter.ClusterID,
		ocLoginAtter.AdditioanlFlags...***REMOVED***

	output, err := RetryCMDRun(cmd, ocLoginAtter.Timeout***REMOVED***
	return output, err

}
