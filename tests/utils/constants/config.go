package constants

import (
	"fmt"
	"os"
	"path"
	"strings"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

const (
	TokenURL       = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	ClientID       = "cloud-services"
	ClientSecret   = ""
	SkipAuth       = true
	Integration    = false
	HealthcheckURL = "http://localhost:8083"
)

var RHCS = new(RHCSconfig)

// RHCSConfig contains platforms info for the RHCS testing
type RHCSconfig struct {
	// Env is the OpenShift Cluster Management environment used to provision clusters.
	RHCSEnv           string `env:"RHCS_ENV" default:"staging" yaml:"env"`
	ClusterProfile    string `env:"CLUSTER_PROFILE" yaml:"clusterProfile,omitempty"`
	ClusterProfileDir string `env:"CLUSTER_PROFILE_DIR" yaml:"clusterProfileDir,omitempty"`
	RhcsOutputDir     string
	YAMLProfilesDir   string
	RootDir           string
	KubeConfigDir     string
}

func init() {
	currentDir, _ := os.Getwd()
	project := "terraform-provider-rhcs"
	RHCS.RootDir = GetEnvWithDefault(WorkSpace, strings.SplitAfter(currentDir, project)[0])

	// defaulted to staging
	RHCS.RHCSEnv = GetEnvWithDefault(RHCSENV, RHCS.RHCSEnv)

	if os.Getenv("CLUSTER_PROFILE") != "" {
		RHCS.ClusterProfile = os.Getenv("CLUSTER_PROFILE")
	}
	if os.Getenv("CLUSTER_PROFILE_DIR") != "" {
		RHCS.ClusterProfileDir = os.Getenv("CLUSTER_PROFILE_DIR")
	}

	RHCS.RhcsOutputDir = GetRHCSOutputDir()
	RHCS.KubeConfigDir = GetKubeConfigDir()
	RHCS.YAMLProfilesDir = path.Join(RHCS.RootDir, "tests", "ci", "profiles")
}

// gatewayURL is used to get the global env of default gateway for testing
// GATEWAY_URL can be set directly in case anybody run on a sandbox url
// Otherwize can use export RHCS_ENV=<staging, production, integration> to set
// Default value is https://api.stage.openshift.com which is staging env
func gatewayURL() (url string, ocmENV string) {
	url = GetEnvWithDefault("GATEWAY_URL", "")
	if url != "" {
		return url, ""
	} else {
		switch os.Getenv(RHCSENV) {
		case "production", "prod":
			url = "api.openshift.com"
			ocmENV = "production"
		case "staging", "stage":
			url = "api.stage.openshift.com"
			ocmENV = "staging"
		case "integration", "int":
			url = "api.integration.openshift.com"
			ocmENV = "integration"
		case "local":
			url = ""
			ocmENV = "local"
		default:
			url = "api.stage.openshift.com"
			ocmENV = "staging"
		}
	}
	url = fmt.Sprintf("https://%s", url)
	Logger.Infof("Running against env %s with gateway url %s", ocmENV, url)
	return url, ocmENV
}

var GateWayURL, OCMENV = gatewayURL()

func GetEnvWithDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		if key == TokenENVName {
			panic(fmt.Errorf("ENV Variable RHCS_TOKEN is empty, please make sure you set the env value"))
		}
	}
	return defaultValue
}

func GetRHCSOutputDir() string {
	var rhcsNewOutPath string

	if GetEnvWithDefault("RHCS_OUTPUT", "") != "" {
		rhcsNewOutPath = os.Getenv("RHCS_OUTPUT")
		return rhcsNewOutPath
	}
	rhcsNewOutPath = path.Join(RHCS.RootDir, "tests", "rhcs_output")
	os.MkdirAll(rhcsNewOutPath, 0777)
	return rhcsNewOutPath
}

func GetKubeConfigDir() string {
	outputDIR := GetRHCSOutputDir()
	configDir := path.Join(outputDIR, "kubeconfig")
	if _, err := os.Stat(configDir); err != nil {
		os.MkdirAll(configDir, 0777)
	}
	return configDir
}
