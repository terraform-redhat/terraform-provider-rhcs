package constants

import (
	"os"
	"path"
)

func initDIR() string {
	currentDir, _ := os.Getwd()
	return path.Join(currentDir, "tf-manifests")
}

var configrationDir = initDIR()

// Provider dirs' name definition
const (
	AWSProviderDIR   = "aws"
	AZUREProviderDIR = "azure"
	RHCSProviderDIR  = "rhcs"
)

// Dirs of aws provider
var (
	AccountRolesDir  = path.Join(configrationDir, AWSProviderDIR, "account-roles")
	OperatorRolesDir = path.Join(configrationDir, AWSProviderDIR, "operator-roles")
	AWSVPCDir        = path.Join(configrationDir, AWSProviderDIR, "vpc")
)

// Dirs of rhcs provider
var (
	ClusterDir     = path.Join(configrationDir, RHCSProviderDIR, "clusters")
	IDPsDir        = path.Join(configrationDir, RHCSProviderDIR, "idps")
	MachinePoolDir = path.Join(configrationDir, RHCSProviderDIR, "machine-pools")
)

// Dirs of different types of clusters
var (
	ROSAclassic = path.Join(ClusterDir, "rosa-classic")
	OSDccs      = path.Join(ClusterDir, "osd-ccs")
)

// Dirs of azure provider
