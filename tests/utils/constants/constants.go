package constants

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	X = "x"
	Y = "y"
	Z = "z"

	UnderscoreConnector string = "_"
	DotConnector        string = "."
	HyphenConnector     string = "-"
)

var (
	TokenENVName              = "RHCS_TOKEN"
	ClusterIDEnv              = "CLUSTER_ID"
	OCMEnv                    = "OCM_ENV"
	RhcsClusterProfileENV     = "CLUSTER_PROFILE"
	ClusterTypeManifestDirEnv = "CLUSTER_ROSA_TYPE"
	MajorVersion              = "MAJOR_VERSION_ENV"
	ManifestsDirENV           = os.Getenv("MANIFESTS_DIR")
)

var (
	DefaultMajorVersion = "4.13"
	CharsBytes          = "abcdefghijklmnopqrstuvwxyz123456789"
	WorkSpace           = "WORKSPACE"
	RHCSPrefix          = "rhcs"
	TFYAMLProfile       = "tf_cluster_profile.yml"
)

const (
	DefaultAWSRegion = "us-east-2"
)

func initDIR() string {
	if ManifestsDirENV != "" {
		return ManifestsDirENV
	}
	currentDir, _ := os.Getwd()
	manifestsDir := path.Join(strings.SplitAfter(currentDir, "tests")[0], "tf-manifests")
	if _, err := os.Stat(manifestsDir); err != nil {
		panic(fmt.Sprintf("Manifests dir %s doesn't exist. Make sure you have the manifests dir in testing repo or set the correct env MANIFESTS_DIR value", manifestsDir))
	}
	return manifestsDir
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
	AccountRolesDir                      = path.Join(configrationDir, AWSProviderDIR, "account-roles")
	OIDCProviderOperatorRolesManifestDir = path.Join(configrationDir, AWSProviderDIR, "oidc-provider-operator-roles")
	AWSVPCDir                            = path.Join(configrationDir, AWSProviderDIR, "vpc")
)

// Dirs of rhcs provider
var (
	ClusterDir     = path.Join(configrationDir, RHCSProviderDIR, "clusters")
	IDPsDir        = path.Join(configrationDir, RHCSProviderDIR, "idps")
	MachinePoolDir = path.Join(configrationDir, RHCSProviderDIR, "machine-pools")
)

// Dirs of different types of clusters
var (
	ROSAClassic = path.Join(ClusterDir, "rosa-classic")
	OSDCCS      = path.Join(ClusterDir, "osd-ccs")
)

// Dirs of identity providers
var (
	HtpasswdDir = path.Join(IDPsDir, "htpasswd")
	GitlabDir   = path.Join(IDPsDir, "gitlab")
	GithubDir   = path.Join(IDPsDir, "github")
	LdapDir     = path.Join(IDPsDir, "ldap")
	OpenidDir   = path.Join(IDPsDir, "openid")
	GoogleDir   = path.Join(IDPsDir, "google")
)

// Supports abs and relatives
func GrantClusterManifestDir(manifestDir string) string {
	var targetDir string
	if strings.Contains(manifestDir, ClusterDir) {
		targetDir = manifestDir
	} else {
		targetDir = path.Join(ClusterDir, manifestDir)
	}
	return targetDir
}
