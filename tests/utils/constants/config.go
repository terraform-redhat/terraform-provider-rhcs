package constants

***REMOVED***
***REMOVED***
	"os"
***REMOVED***
	"strings"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
***REMOVED***

const (
	TokenURL       = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	ClientID       = "cloud-services"
	ClientSecret   = ""
	SkipAuth       = true
	Integration    = false
	HealthcheckURL = "http://localhost:8083"
***REMOVED***

var RHCS = new(RHCSconfig***REMOVED***

// RHCSConfig contains platforms info for the RHCS testing
type RHCSconfig struct {
	// Env is the OpenShift Cluster Management environment used to provision clusters.
	RHCSEnv           string `env:"RHCS_ENV" default:"staging" yaml:"env"`
	ClusterProfile    string `env:"CLUSTER_PROFILE" yaml:"clusterProfile,omitempty"`
	ClusterProfileDir string `env:"CLUSTER_PROFILE_DIR" yaml:"clusterProfileDir,omitempty"`
	RhcsOutputDir     string
	YAMLProfilesDir   string
	RootDir           string
}

func init(***REMOVED*** {
	currentDir, _ := os.Getwd(***REMOVED***
	project := "terraform-provider-rhcs"
	RHCS.RootDir = GetEnvWithDefault(WorkSpace, strings.SplitAfter(currentDir, project***REMOVED***[0]***REMOVED***

	// defaulted to staging
	RHCS.RHCSEnv = GetEnvWithDefault(RHCSENV, RHCS.RHCSEnv***REMOVED***

	if os.Getenv("CLUSTER_PROFILE"***REMOVED*** != "" {
		RHCS.ClusterProfile = os.Getenv("CLUSTER_PROFILE"***REMOVED***
	}
	if os.Getenv("CLUSTER_PROFILE_DIR"***REMOVED*** != "" {
		RHCS.ClusterProfileDir = os.Getenv("CLUSTER_PROFILE_DIR"***REMOVED***
	}

	RHCS.RhcsOutputDir = GetRhcsOutputDir(***REMOVED***
	RHCS.YAMLProfilesDir = path.Join(RHCS.RootDir, "tests", "ci", "profiles"***REMOVED***
}

// gatewayURL is used to get the global env of default gateway for testing
// GATEWAY_URL can be set directly in case anybody run on a sandbox url
// Otherwize can use export RHCS_ENV=<staging, production, integration> to set
// Default value is https://api.stage.openshift.com which is staging env
func gatewayURL(***REMOVED*** (url string, ocmENV string***REMOVED*** {
	url = GetEnvWithDefault("GATEWAY_URL", ""***REMOVED***
	if url != "" {
		return url, ""
	} else {
		switch os.Getenv(RHCSENV***REMOVED*** {
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
***REMOVED***
	}
	url = fmt.Sprintf("https://%s", url***REMOVED***
	Logger.Infof("Running against env %s with gateway url %s", ocmENV, url***REMOVED***
	return url, ocmENV
}

var GateWayURL, OCMENV = gatewayURL(***REMOVED***

func GetEnvWithDefault(key string, defaultValue string***REMOVED*** string {
	if value, ok := os.LookupEnv(key***REMOVED***; ok {
		return value
	} else {
		if key == TokenENVName {
			panic(fmt.Errorf("ENV Variable RHCS_TOKEN is empty, please make sure you set the env value"***REMOVED******REMOVED***
***REMOVED***
	}
	return defaultValue
}

func GetRhcsOutputDir(***REMOVED*** string {
	var rhcsNewOutPath string

	if GetEnvWithDefault("RHCS_OUTPUT", ""***REMOVED*** != "" {
		rhcsNewOutPath = os.Getenv("RHCS_OUTPUT"***REMOVED***
		return rhcsNewOutPath
	}
	rhcsNewOutPath = path.Join(RHCS.RootDir, "tests", "rhcs_output"***REMOVED***
	os.MkdirAll(rhcsNewOutPath, 0777***REMOVED***
	return rhcsNewOutPath
}
