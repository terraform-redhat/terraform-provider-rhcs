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

// Below constants is the env variable name defined to run on different testing requirements
const (
	TokenENVName              = "RHCS_TOKEN"
	ClusterIDEnv              = "CLUSTER_ID"
	RHCSENV                   = "RHCS_ENV"
	RhcsClusterProfileENV     = "CLUSTER_PROFILE"
	QEUsage                   = "QE_USAGE"
	ClusterTypeManifestDirEnv = "CLUSTER_ROSA_TYPE"
	MajorVersion              = "MAJOR_VERSION_ENV"
	RHCSVersion               = "RHCS_VERSION"
	RHCSSource                = "RHCS_SOURCE"
)

var (
	DefaultMajorVersion       = "4.14"
	CharsBytes                = "abcdefghijklmnopqrstuvwxyz123456789"
	WorkSpace                 = "WORKSPACE"
	RHCSPrefix                = "rhcs"
	TFYAMLProfile             = "tf_cluster_profile.yml"
	ConfigSuffix              = "kubeconfig"
	DefaultAccountRolesPrefix = "account-role-"
	ManifestsDirENV           = os.Getenv("MANIFESTS_FOLDER")
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

var ConfigrationDir = initDIR()

// Provider dirs' name definition
const (
	AWSProviderDIR   = "aws"
	AZUREProviderDIR = "azure"
	RHCSProviderDIR  = "rhcs"
)

// Dirs of aws provider
var (
	AccountRolesDir                      = path.Join(ConfigrationDir, AWSProviderDIR, "account-roles")
	AddAccountRolesDir                   = path.Join(ConfigrationDir, AWSProviderDIR, "add-account-roles")
	OIDCProviderOperatorRolesManifestDir = path.Join(ConfigrationDir, AWSProviderDIR, "oidc-provider-operator-roles")
	AWSVPCDir                            = path.Join(ConfigrationDir, AWSProviderDIR, "vpc")
	AWSVPCTagDir                         = path.Join(ConfigrationDir, AWSProviderDIR, "vpc-tags")
	AWSSecurityGroupDir                  = path.Join(ConfigrationDir, AWSProviderDIR, "security-groups")
)

// Dirs of rhcs provider
var (
	ClusterDir            = path.Join(ConfigrationDir, RHCSProviderDIR, "clusters")
	ImportResourceDir     = path.Join(ConfigrationDir, RHCSProviderDIR, "resource-import")
	IDPsDir               = path.Join(ConfigrationDir, RHCSProviderDIR, "idps")
	MachinePoolDir        = path.Join(ConfigrationDir, RHCSProviderDIR, "machine-pools")
	DNSDir                = path.Join(ConfigrationDir, RHCSProviderDIR, "dns")
	RhcsInfoDir           = path.Join(ConfigrationDir, RHCSProviderDIR, "rhcs-info")
	DefaultMachinePoolDir = path.Join(ConfigrationDir, RHCSProviderDIR, "default-machine-pool")
	KubeletConfigDir      = path.Join(ConfigrationDir, RHCSProviderDIR, "kubelet-config")
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
	MultiIDPDir = path.Join(IDPsDir, "multi-idp")
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

func GrantTFvarsFile(manifestDir string) string {
	return path.Join(manifestDir, "terraform.tfvars")
}

func GrantTFstateFile(manifestDir string) string {
	return path.Join(manifestDir, "terraform.tfstate")
}

// Machine pool taints effect
const (
	NoExecute        = "NoExecute"
	NoSchedule       = "NoSchedule"
	PreferNoSchedule = "PreferNoSchedule"
)
