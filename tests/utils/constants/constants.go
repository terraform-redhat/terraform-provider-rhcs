package constants

import (
	"fmt"
	"os"

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

const (
	DefaultAWSRegion = "us-east-2"
)

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

type IDPType string

const (
	IDPHtpasswd IDPType = "htpasswd"
	IDPGitlab   IDPType = "gitlab"
	IDPGithub   IDPType = "github"
	IDPLdap     IDPType = "ldap"
	IDPOpenid   IDPType = "openid"
	IDPGoogle   IDPType = "google"
	IDPMultiIDP IDPType = "multi-idp"
)
