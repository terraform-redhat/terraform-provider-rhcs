package constants

***REMOVED***
***REMOVED***
	"os"
***REMOVED***
	"strings"
***REMOVED***

const (
	X = "x"
	Y = "y"
	Z = "z"

	UnderscoreConnector string = "_"
	DotConnector        string = "."
	HyphenConnector     string = "-"
***REMOVED***

var (
	TokenENVName              = "RHCS_TOKEN"
	ClusterIDEnv              = "CLUSTER_ID"
	RHCSENV                   = "RHCS_ENV"
	RhcsClusterProfileENV     = "CLUSTER_PROFILE"
	ClusterTypeManifestDirEnv = "CLUSTER_ROSA_TYPE"
	MajorVersion              = "MAJOR_VERSION_ENV"
	ManifestsDirENV           = os.Getenv("MANIFESTS_FOLDER"***REMOVED***
***REMOVED***

var (
	DefaultMajorVersion = "4.13"
	CharsBytes          = "abcdefghijklmnopqrstuvwxyz123456789"
	WorkSpace           = "WORKSPACE"
	RHCSPrefix          = "rhcs"
	TFYAMLProfile       = "tf_cluster_profile.yml"
	LdapURL             = "ldap://ldap.forumsys.com/dc=example,dc=com?uid"
	GitLabURL           = "https://gitlab.cee.redhat.com"
	Organizations       = []string{"openshift"}
	HostedDomain        = "redhat.com"
	DefaultMPLabels     = map[string]string{
		"test1": "testdata1",
	}
***REMOVED***

const (
	DefaultAWSRegion = "us-east-2"
***REMOVED***

func initDIR(***REMOVED*** string {
	if ManifestsDirENV != "" {
		return ManifestsDirENV
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
	AccountRolesDir                      = path.Join(configrationDir, AWSProviderDIR, "account-roles"***REMOVED***
	OIDCProviderOperatorRolesManifestDir = path.Join(configrationDir, AWSProviderDIR, "oidc-provider-operator-roles"***REMOVED***
	AWSVPCDir                            = path.Join(configrationDir, AWSProviderDIR, "vpc"***REMOVED***
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

// Dirs of identity providers
var (
	HtpasswdDir = path.Join(IDPsDir, "htpasswd"***REMOVED***
	GitlabDir   = path.Join(IDPsDir, "gitlab"***REMOVED***
	GithubDir   = path.Join(IDPsDir, "github"***REMOVED***
	LdapDir     = path.Join(IDPsDir, "ldap"***REMOVED***
	OpenidDir   = path.Join(IDPsDir, "openid"***REMOVED***
	GoogleDir   = path.Join(IDPsDir, "google"***REMOVED***
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

// Machine pool taints effect
const (
	NoExecute        = "NoExecute"
	NoSchedule       = "NoSchedule"
	PreferNoSchedule = "PreferNoSchedule"
***REMOVED***
