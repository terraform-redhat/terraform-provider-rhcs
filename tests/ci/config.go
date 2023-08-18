package ci

***REMOVED***
***REMOVED***
	"os"
***REMOVED***
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***

const (
	tokenURL       = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	clientID       = "cloud-services"
	clientSecret   = ""
	skipAuth       = true
	integration    = false
	healthcheckURL = "http://localhost:8083"
***REMOVED***

var RHCS = new(RHCSconfig***REMOVED***

// RHCSConfig contains platforms info for the RHCS testing
type RHCSconfig struct {
	// Env is the OpenShift Cluster Management environment used to provision clusters.
	RHCSEnv           string `env:"RHCS_ENV" default:"staging" yaml:"env"`
	ClusterProfile    string `env:"CLUSTER_PROFILE" yaml:"clusterProfile,omitempty"`
	ClusterProfileDir string `env:"CLUSTER_PROFILE_DIR" yaml:"clusterProfileDir,omitempty"`

	YAMLProfilesDir string
}

func init(***REMOVED*** {
	currentDir, _ := os.Getwd(***REMOVED***
	project := "terraform-provider-rhcs"
	rootDir := GetEnvWithDefault(CON.WorkSpace, strings.SplitAfter(currentDir, project***REMOVED***[0]***REMOVED***

	// defaulted to staging
	RHCS.RHCSEnv = GetEnvWithDefault(CON.OCMEnv, RHCS.RHCSEnv***REMOVED***

	if os.Getenv("CLUSTER_PROFILE"***REMOVED*** != "" {
		RHCS.ClusterProfile = os.Getenv("CLUSTER_PROFILE"***REMOVED***
	}
	if os.Getenv("CLUSTER_PROFILE_DIR"***REMOVED*** != "" {
		RHCS.ClusterProfileDir = os.Getenv("CLUSTER_PROFILE_DIR"***REMOVED***
	}

	RHCS.YAMLProfilesDir = path.Join(rootDir, "tests", "ci", "profiles"***REMOVED***
}

func gatewayURL(***REMOVED*** (url string***REMOVED*** {
	url = GetEnvWithDefault("OCM_BASE_URL", ""***REMOVED***

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

	url = fmt.Sprintf("https://%s", url***REMOVED***
	return url
}

func GateWayURL(***REMOVED*** string {
	return gatewayURL(***REMOVED***
}

func GetEnvWithDefault(key string, defaultValue string***REMOVED*** string {
	if value, ok := os.LookupEnv(key***REMOVED***; ok {
		return value
	} else {
		if key == CON.TokenENVName {
			panic(fmt.Errorf("ENV Variable RHCS_TOKEN is empty, please make sure you set the env value"***REMOVED******REMOVED***
***REMOVED***
	}
	return defaultValue
}
