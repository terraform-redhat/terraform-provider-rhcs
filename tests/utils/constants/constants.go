package constants

import (
	"fmt"
	"os"
	"path"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// OCP version streams
const (
	X = "x"
	Y = "y"
	Z = "z"
)

const (
	UnderscoreConnector string = "_"
	DotConnector        string = "."
	HyphenConnector     string = "-"
)

// Upgrade Policy States
const (
	Pending   = "pending"
	Scheduled = "scheduled"
	Started   = "started"
	Completed = "completed"
	Delayed   = "delayed"
	Failed    = "failed"
	Cancelled = "cancelled"
	Waiting   = "waiting"
)

// Cluster state
const (
	Ready = "ready"
)

var (
	AutomaticScheduleType cmv1.ScheduleType = "automatic"
	ManualScheduleType    cmv1.ScheduleType = "manual"
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
	WaitOperators             = "WAIT_OPERATORS"
	RHCS_CLUSTER_NAME         = "RHCS_CLUSTER_NAME"
	RHCS_CLUSTER_NAME_PREFIX  = "RHCS_CLUSTER_NAME_PREFIX"
	RHCS_CLUSTER_NAME_SUFFIX  = "RHCS_CLUSTER_NAME_SUFFIX"
)

var (
	DefaultMajorVersion                  = "4.14"
	CharsBytes                           = "abcdefghijklmnopqrstuvwxyz123456789"
	WorkSpace                            = "WORKSPACE"
	RHCSPrefix                           = "rhcs"
	ConfigSuffix                         = "kubeconfig"
	DefaultAccountRolesPrefix            = "account-role-"
	ManifestsDirENV                      = os.Getenv("MANIFESTS_FOLDER")
	SharedVpcAWSSharedCredentialsFileENV = os.Getenv("SHARED_VPC_AWS_SHARED_CREDENTIALS_FILE")
)

var (
	DefaultAWSRegion = "us-east-2"
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

type ClusterType struct {
	Name string
	HCP  bool
}

var (
	ROSA_CLASSIC = ClusterType{Name: "rosa-classic"}
	ROSA_HCP     = ClusterType{Name: "rosa-hcp", HCP: true}

	allClusterTypes = []ClusterType{
		ROSA_CLASSIC,
		ROSA_HCP,
	}
)

func FindClusterType(clusterTypeName string) ClusterType {
	for _, clusterType := range allClusterTypes {
		if clusterType.String() == clusterTypeName {
			return clusterType
		}
	}
	panic(fmt.Sprintf("Unknown cluster type %s", clusterTypeName))
}

func (ct *ClusterType) String() string {
	return ct.Name
}
