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
	RHCSURL                   = "RHCS_URL"
	RhcsClusterProfileENV     = "CLUSTER_PROFILE"
	QEUsage                   = "QE_USAGE"
	ClusterTypeManifestDirEnv = "CLUSTER_ROSA_TYPE"
	MajorVersion              = "MAJOR_VERSION_ENV"
	RHCSVersion               = "RHCS_VERSION"
	RHCSSource                = "RHCS_SOURCE"
	// Set this to update the version for the terraform-redhat/rosa-classic/rhcs module
	ModuleVersion = "MODULE_VERSION"
	// Set this to update the source for the terraform-redhat/rosa-classic/rhcs module
	ModuleSource = "MODULE_SOURCE"
	// Set this to any value if `MODULE_SOURCE` refers to a local path, in which case `MODULE_VERSION` will be ignored
	ModuleSourceLocal        = "MODULE_SOURCE_LOCAL"
	WaitOperators            = "WAIT_OPERATORS"
	RHCS_CLUSTER_NAME        = "RHCS_CLUSTER_NAME"
	RHCS_CLUSTER_NAME_PREFIX = "RHCS_CLUSTER_NAME_PREFIX"
	RHCS_CLUSTER_NAME_SUFFIX = "RHCS_CLUSTER_NAME_SUFFIX"
	COMPUTE_MACHINE_TYPE     = "COMPUTE_MACHINE_TYPE"
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

func GrantTFstateFile(manifestDir string) string {
	return path.Join(manifestDir, "terraform.tfstate")
}

// Machine pool taints effect
const (
	NoExecute        = "NoExecute"
	NoSchedule       = "NoSchedule"
	PreferNoSchedule = "PreferNoSchedule"
)

// Machine pool
const (
	DefaultMachinePoolName = "worker"
	DefaultNodePoolName    = "workers"
)

// Ec2MetadataHttpTokens for hcp cluster
const (
	DefaultEc2MetadataHttpTokens  = "optional"
	RequiredEc2MetadataHttpTokens = "required"
	OptionalEc2MetadataHttpTokens = "optional"
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

func GetHCPClusterTypes() (types []ClusterType) {
	for _, ct := range allClusterTypes {
		if ct.HCP {
			types = append(types, ct)
		}
	}
	return
}

func (ct *ClusterType) String() string {
	return ct.Name
}
