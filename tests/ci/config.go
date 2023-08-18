package ci

import (
	"fmt"
	"os"
	"path"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

const (
	tokenURL       = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	clientID       = "cloud-services"
	clientSecret   = ""
	skipAuth       = true
	integration    = false
	healthcheckURL = "http://localhost:8083"
)

var RHCS = new(RHCSconfig)

// RHCSConfig contains platforms info for the RHCS testing
type RHCSconfig struct {
	// Env is the OpenShift Cluster Management environment used to provision clusters.
	RHCSEnv           string `env:"RHCS_ENV" default:"staging" yaml:"env"`
	ClusterProfile    string `env:"CLUSTER_PROFILE" yaml:"clusterProfile,omitempty"`
	ClusterProfileDir string `env:"CLUSTER_PROFILE_DIR" yaml:"clusterProfileDir,omitempty"`

	YAMLProfilesDir string
}

func init() {
	currentDir, _ := os.Getwd()
	project := "terraform-provider-rhcs"
	rootDir := GetEnvWithDefault(CON.WorkSpace, strings.SplitAfter(currentDir, project)[0])

	// defaulted to staging
	RHCS.RHCSEnv = GetEnvWithDefault(CON.OCMEnv, RHCS.RHCSEnv)

	if os.Getenv("CLUSTER_PROFILE") != "" {
		RHCS.ClusterProfile = os.Getenv("CLUSTER_PROFILE")
	}
	if os.Getenv("CLUSTER_PROFILE_DIR") != "" {
		RHCS.ClusterProfileDir = os.Getenv("CLUSTER_PROFILE_DIR")
	}

	RHCS.YAMLProfilesDir = path.Join(rootDir, "tests", "ci", "profiles")
}

func gatewayURL() (url string) {
	url = GetEnvWithDefault("OCM_BASE_URL", "")

	switch RHCS.RHCSEnv {
	case "production", "prod":
		url = "api.openshift.com"
	case "staging", "stage":
		url = "api.stage.openshift.com"
	case "integration", "int":
		url = "api.integration.openshift.com"
	default:
		url = "api.stage.openshift.com"
	}

	url = fmt.Sprintf("https://%s", url)
	return url
}

func GateWayURL() string {
	return gatewayURL()
}

func GetEnvWithDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		if key == CON.TokenENVName {
			panic(fmt.Errorf("ENV Variable RHCS_TOKEN is empty, please make sure you set the env value"))
		}
	}
	return defaultValue
}
