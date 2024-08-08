package exec

import "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"

type ClusterArgs struct {
	AccountRolePrefix                    *string            `hcl:"account_role_prefix"`
	ClusterName                          *string            `hcl:"cluster_name"`
	OperatorRolePrefix                   *string            `hcl:"operator_role_prefix"`
	OpenshiftVersion                     *string            `hcl:"openshift_version"`
	AWSRegion                            *string            `hcl:"aws_region"`
	AWSAvailabilityZones                 *[]string          `hcl:"aws_availability_zones"`
	Replicas                             *int               `hcl:"replicas"`
	ChannelGroup                         *string            `hcl:"channel_group"`
	Ec2MetadataHttpTokens                *string            `hcl:"ec2_metadata_http_tokens"`
	PrivateLink                          *bool              `hcl:"private_link"`
	Private                              *bool              `hcl:"private"`
	Fips                                 *bool              `hcl:"fips"`
	Tags                                 *map[string]string `hcl:"tags"`
	AuditLogForward                      *bool              `hcl:"audit_log_forward"`
	Autoscaling                          *Autoscaling       `hcl:"autoscaling"`
	Etcd                                 *bool              `hcl:"etcd_encryption"`
	EtcdKmsKeyARN                        *string            `hcl:"etcd_kms_key_arn"`
	KmsKeyARN                            *string            `hcl:"kms_key_arn"`
	AWSSubnetIDs                         *[]string          `hcl:"aws_subnet_ids"`
	ComputeMachineType                   *string            `hcl:"compute_machine_type"`
	DefaultMPLabels                      *map[string]string `hcl:"default_mp_labels"`
	DisableSCPChecks                     *bool              `hcl:"disable_scp_checks"`
	MultiAZ                              *bool              `hcl:"multi_az"`
	CustomProperties                     *map[string]string `hcl:"custom_properties"`
	WorkerDiskSize                       *int               `hcl:"worker_disk_size"`
	AdditionalComputeSecurityGroups      *[]string          `hcl:"additional_compute_security_groups"`
	AdditionalInfraSecurityGroups        *[]string          `hcl:"additional_infra_security_groups"`
	AdditionalControlPlaneSecurityGroups *[]string          `hcl:"additional_control_plane_security_groups"`
	MachineCIDR                          *string            `hcl:"machine_cidr"`
	OIDCConfigID                         *string            `hcl:"oidc_config_id"`
	AdminCredentials                     *map[string]string `hcl:"admin_credentials"`
	DisableUWM                           *bool              `hcl:"disable_workload_monitoring"`
	Proxy                                *Proxy             `hcl:"proxy"`
	UnifiedAccRolesPath                  *string            `hcl:"path"`
	UpgradeAcknowledgementsFor           *string            `hcl:"upgrade_acknowledgements_for"`
	BaseDnsDomain                        *string            `hcl:"base_dns_domain"`
	PrivateHostedZone                    *PrivateHostedZone `hcl:"private_hosted_zone"`
	WaitForCluster                       *bool              `hcl:"wait_for_cluster"`
	DisableClusterWaiter                 *bool              `hcl:"disable_cluster_waiter"`
	DisableWaitingInDestroy              *bool              `hcl:"disable_waiting_in_destroy"`

	AWSAccountID        *string `hcl:"aws_account_id"`
	AWSBillingAccountID *string `hcl:"aws_billing_account_id"`
	HostPrefix          *int    `hcl:"host_prefix"`
	ServiceCIDR         *string `hcl:"service_cidr"`
	PodCIDR             *string `hcl:"pod_cidr"`
	StsInstallerRole    *string `hcl:"installer_role"`
	StsSupportRole      *string `hcl:"support_role"`
	StsWorkerRole       *string `hcl:"worker_role"`

	IncludeCreatorProperty *bool `hcl:"include_creator_property"`

	FullResources *bool `hcl:"full_resources"`
}
type Proxy struct {
	HTTPProxy             *string `cty:"http_proxy"`
	HTTPSProxy            *string `cty:"https_proxy"`
	AdditionalTrustBundle *string `cty:"additional_trust_bundle"`
	NoProxy               *string `cty:"no_proxy"`
}

type PrivateHostedZone struct {
	ID      string `cty:"id"`
	RoleArn string `cty:"role_arn"`
}

type Autoscaling struct {
	AutoscalingEnabled *bool `cty:"autoscaling_enabled"`
	MinReplicas        *int  `cty:"min_replicas"`
	MaxReplicas        *int  `cty:"max_replicas"`
}

// Just a placeholder, not research what to output yet.
type ClusterOutput struct {
	ClusterID                            string            `json:"cluster_id,omitempty"`
	ClusterName                          string            `json:"cluster_name,omitempty"`
	ClusterVersion                       string            `json:"cluster_version,omitempty"`
	AdditionalComputeSecurityGroups      []string          `json:"additional_compute_security_groups,omitempty"`
	AdditionalInfraSecurityGroups        []string          `json:"additional_infra_security_groups,omitempty"`
	AdditionalControlPlaneSecurityGroups []string          `json:"additional_control_plane_security_groups,omitempty"`
	Properties                           map[string]string `json:"properties,omitempty"`
	UserTags                             map[string]string `json:"tags,omitempty"`
}

// ******************************************************
// RHCS test cases used
const (

	// MaxExpiration in unit of hour
	MaxExpiration = 168

	// MaxNodeNumber means max node number per cluster/machinepool
	MaxNodeNumber = 180

	// MaxNameLength means cluster name will be trimed when request certificate
	MaxNameLength = 15

	MaxIngressNumber = 2
)

// version channel_groups
const (
	FastChannel      = "fast"
	StableChannel    = "stable"
	NightlyChannel   = "nightly"
	CandidateChannel = "candidate"
)

type ClusterService interface {
	Init() error
	Plan(args *ClusterArgs) (string, error)
	Apply(args *ClusterArgs) (string, error)
	Output() (*ClusterOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*ClusterArgs, error)
	WriteTFVars(args *ClusterArgs) error
	DeleteTFVars() error
}

type clusterService struct {
	tfExecutor TerraformExecutor
}

func NewClusterService(manifestDir string) (ClusterService, error) {
	manifestsDir := constants.GrantClusterManifestDir(manifestDir)
	svc := &clusterService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *clusterService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *clusterService) Plan(args *ClusterArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *clusterService) Apply(args *ClusterArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *clusterService) Output() (*ClusterOutput, error) {
	var output ClusterOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *clusterService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *clusterService) ReadTFVars() (*ClusterArgs, error) {
	args := &ClusterArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *clusterService) WriteTFVars(args *ClusterArgs) error {
	err := svc.tfExecutor.WriteTerraformVars(args)
	return err
}

func (svc *clusterService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
