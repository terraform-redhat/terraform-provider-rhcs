package constants

***REMOVED***
	"os"
***REMOVED***
***REMOVED***

func initDIR(***REMOVED*** string {
	currentDir, _ := os.Getwd(***REMOVED***
	return path.Join(currentDir, "tf-manifests"***REMOVED***
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
	ROSAclassic = path.Join(ClusterDir, "rosa-classic"***REMOVED***
	OSDccs      = path.Join(ClusterDir, "osd-ccs"***REMOVED***
***REMOVED***

// Dirs of azure provider
