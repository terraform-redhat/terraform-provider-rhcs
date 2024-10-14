package constants

import (
	"fmt"

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

const (
	TokenURL       = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	ClientID       = "cloud-services"
	ClientSecret   = ""
	SkipAuth       = true
	Integration    = false
	HealthcheckURL = "http://localhost:8083"
)

var (
	CharsBytes                = "abcdefghijklmnopqrstuvwxyz123456789"
	RHCSPrefix                = "rhcs"
	ConfigSuffix              = "kubeconfig"
	DefaultAccountRolesPrefix = "account-role-"
)

var (
	DefaultAWSRegion = "us-east-2"
	DefaultRHCSURL   = "https://api.openshift.com"
)

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

type IDPType string

const (
	IDPHTPassword IDPType = "htpasswd"
	IDPGitlab     IDPType = "gitlab"
	IDPGithub     IDPType = "github"
	IDPGoogle     IDPType = "google"
	IDPLDAP       IDPType = "ldap"
	IDPOpenID     IDPType = "openid"
	IDPMulti      IDPType = "multi-idp"
)

const (

	// MaxExpiration in unit of hour
	ClusterMaxExpiration = 168

	// MaxNodeNumber means max node number per cluster/machinepool
	ClusterMaxNodeNumber = 180

	// MaxNameLength means cluster name will be trimed when request certificate
	ClusterMaxNameLength = 15

	ClusterMaxIngressNumber = 2
)

// disk size
const (
	MinClassicDiskSize = 128
	MinHCPDiskSize     = 75
	MaxDiskSize        = 16384
)
