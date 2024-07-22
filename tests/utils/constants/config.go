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
	RHCSURL               string `env:"RHCS_URL" yaml:"rhcsURL,omitempty"`
	OCMEnv                string `env:"OCM_ENV" default:"staging" yaml:"ocmENV,omitempty"`
	ClusterProfile        string `env:"CLUSTER_PROFILE" yaml:"clusterProfile,omitempty"`
	ClusterProfileDir     string `env:"CLUSTER_PROFILE_DIR" yaml:"clusterProfileDir,omitempty"`
	RhcsOutputDir         string
	YAMLProfilesDir       string
	RootDir               string
	KubeConfigDir         string
	RHCSSource            string `env:"RHCS_SOURCE" default:"staging" yaml:"rhcsSource,omitempty"`
	RHCSVersion           string `env:"RHCS_VERSION" default:"staging" yaml:"rhcsVersion,omitempty"`
	RHCSClusterName       string `env:"RHCS_CLUSTER_NAME" yaml:"clusterName,omitempty"`
	RHCSClusterNamePrefix string `env:"RHCS_CLUSTER_NAME_PREFIX" yaml:"clusterNamePrefix,omitempty"`
	RHCSClusterNameSuffix string `env:"RHCS_CLUSTER_NAME_SUFFIX" yaml:"clusterNameSuffix,omitempty"`
	ComputeMachineType    string `env:"COMPUTE_MACHINE_TYPE" yaml:"computeMachineType,omitempty"`
}

func init() {
	currentDir, _ := os.Getwd()
	project := "terraform-provider-rhcs"
	RHCS.RootDir = GetEnvWithDefault(WorkSpace, strings.SplitAfter(currentDir, project)[0])

	// defaulted to staging
	RHCS.RHCSURL = GetEnvWithDefault(RHCSURL, "https://api.openshift.com")
	Logger.Infof("Running against RHCS URL: %s", RHCS.RHCSURL)
	RHCS.OCMEnv = ocmEnv(RHCS.RHCSURL)

	RHCS.RHCSClusterName = GetEnvWithDefault(RHCS_CLUSTER_NAME, RHCS.RHCSClusterName)
	RHCS.RHCSClusterNamePrefix = GetEnvWithDefault(RHCS_CLUSTER_NAME_PREFIX, RHCS.RHCSClusterNamePrefix)
	RHCS.RHCSClusterNameSuffix = GetEnvWithDefault(RHCS_CLUSTER_NAME_SUFFIX, RHCS.RHCSClusterNameSuffix)

	RHCS.ComputeMachineType = GetEnvWithDefault(COMPUTE_MACHINE_TYPE, RHCS.ComputeMachineType)

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

// ocmEnv retrieve the env name based on
// Values: production, staging, integration, local
func ocmEnv(rhcsURL string) (ocmENV string) {
	switch rhcsURL {
	case "https://api.openshift.com", "api.openshift.com":
		return "production"
	case "https://api.stage.openshift.com", "api.stage.openshift.com":
		return "staging"
	case "https://api.integration.openshift.com", "api.integration.openshift.com":
		return "integration"
	case "":
		return "local"
	default:
		return "staging"
	}
}

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
