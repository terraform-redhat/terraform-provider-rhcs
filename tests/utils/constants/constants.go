package constants

***REMOVED***
***REMOVED***
	"os"
***REMOVED***
	"strings"
***REMOVED***

func initDIR(***REMOVED*** string {
	if os.Getenv("MANIFESTS_DIR"***REMOVED*** != "" {
		return os.Getenv("MANIFESTS_DIR"***REMOVED***
	}
	currentDir, _ := os.Getwd(***REMOVED***
	manifestsDir := path.Join(strings.SplitAfter(currentDir, "tests"***REMOVED***[0], "tf-manifests"***REMOVED***
	if _, err := os.Stat(manifestsDir***REMOVED***; err != nil {
		panic(fmt.Sprintf("Manifests dir %s doesn't exist. Make sure you have the manifests dir in testing repo or set the correct env MANIFESTS_DIR value", manifestsDir***REMOVED******REMOVED***
	}
	return manifestsDir
}

var configrationDir = initDIR(***REMOVED***

// Provider dirs' name definition
const (
	AWSProviderDIR   = "aws"
	AZUREProviderDIR = "azure"
	RHCSProviderDIR  = "rhcs"
***REMOVED***

// Dirs of aws provider
var (
	AccountRolesDir  = path.Join(configrationDir, AWSProviderDIR, "account-roles"***REMOVED***
	OperatorRolesDir = path.Join(configrationDir, AWSProviderDIR, "operator-roles"***REMOVED***
	AWSVPCDir        = path.Join(configrationDir, AWSProviderDIR, "vpc"***REMOVED***
***REMOVED***

// Dirs of rhcs provider
var (
	ClusterDir     = path.Join(configrationDir, RHCSProviderDIR, "clusters"***REMOVED***
	IDPsDir        = path.Join(configrationDir, RHCSProviderDIR, "idps"***REMOVED***
	MachinePoolDir = path.Join(configrationDir, RHCSProviderDIR, "machine-pools"***REMOVED***
***REMOVED***

// Dirs of different types of clusters
var (
	ROSAClassic = path.Join(ClusterDir, "rosa-classic"***REMOVED***
	OSDCCS      = path.Join(ClusterDir, "osd-ccs"***REMOVED***
***REMOVED***

// Supports abs and relatives
func GrantClusterManifestDir(manifestDir string***REMOVED*** string {
	var targetDir string
	if strings.Contains(manifestDir, ClusterDir***REMOVED*** {
		targetDir = manifestDir
	} else {
		targetDir = path.Join(ClusterDir, manifestDir***REMOVED***
	}
	return targetDir
}

// Dirs of azure provider
// Just a placeholder

const (
	DefaultAWSRegion = "us-east-2"
	TokenENVName     = "RHCS_TOKEN"
***REMOVED***
