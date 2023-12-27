package exec

import (
	"context"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type ClusterCreationArgs struct {
	AccountRolePrefix                    string            `json:"account_role_prefix,omitempty"`
	OCMENV                               string            `json:"rhcs_environment,omitempty"`
	ClusterName                          string            `json:"cluster_name,omitempty"`
	OperatorRolePrefix                   string            `json:"operator_role_prefix,omitempty"`
	OpenshiftVersion                     string            `json:"openshift_version,omitempty"`
	URL                                  string            `json:"url,omitempty"`
	AWSRegion                            string            `json:"aws_region,omitempty"`
	AWSAvailabilityZones                 []string          `json:"aws_availability_zones,omitempty"`
	Replicas                             int               `json:"replicas,omitempty"`
	ChannelGroup                         string            `json:"channel_group,omitempty"`
	AWSHttpTokensState                   string            `json:"aws_http_tokens_state,omitempty"`
	PrivateLink                          bool              `json:"private_link,omitempty"`
	Private                              bool              `json:"private,omitempty"`
	Fips                                 bool              `json:"fips,omitempty"`
	Tagging                              map[string]string `json:"tags,omitempty"`
	AuditLogForward                      bool              `json:"audit_log_forward,omitempty"`
	Autoscale                            bool              `json:"autoscaling_enabled,omitempty"`
	Etcd                                 bool              `json:"etcd_encryption,omitempty"`
	AWSSubnetIDs                         []string          `json:"aws_subnet_ids,omitempty"`
	ComputeMachineType                   string            `json:"compute_machine_type,omitempty"`
	DefaultMPLabels                      map[string]string `json:"default_mp_labels,omitempty"`
	DisableSCPChecks                     bool              `json:"disable_scp_checks,omitempty"`
	MultiAZ                              bool              `json:"multi_az,omitempty"`
	CustomProperties                     map[string]string `json:"custom_properties,omitempty"`
	WorkerDiskSize                       int               `json:"worker_disk_size,omitempty"`
	AdditionalComputeSecurityGroups      []string          `json:"additional_compute_security_groups,omitempty"`
	AdditionalInfraSecurityGroups        []string          `json:"additional_infra_security_groups,omitempty"`
	AdditionalControlPlaneSecurityGroups []string          `json:"additional_control_plane_security_groups,omitempty"`
	MachineCIDR                          string            `json:"machine_cidr,omitempty"`
	OIDCConfigID                         string            `json:"oidc_config_id,omitempty"`
	AdminCredentials                     map[string]string `json:"admin_credentials,omitempty"`
	DisableUWM                           bool              `json:"disable_workload_monitoring,omitempty"`
	Proxy                                *Proxy            `json:"proxy,omitempty"`
	UnifiedAccRolesPath                  string            `json:"path,omitempty"`
}
type Proxy struct {
	HTTPProxy             string `json:"http_proxy,omitempty"`
	HTTPSProxy            string `json:"https_proxy,omitempty"`
	AdditionalTrustBundle string `json:"additional_trust_bundle,omitempty"`
	NoProxy               string `json:"no_proxy,omitempty"`
}

// Just a placeholder, not research what to output yet.
type ClusterOutput struct {
	ClusterID                            string   `json:"cluster_id,omitempty"`
	AdditionalComputeSecurityGroups      []string `json:"additional_compute_security_groups,omitempty"`
	AdditionalInfraSecurityGroups        []string `json:"additional_infra_security_groups,omitempty"`
	AdditionalControlPlaneSecurityGroups []string `json:"additional_control_plane_security_groups,omitempty"`
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

type ClusterService struct {
	CreationArgs *ClusterCreationArgs
	ManifestDir  string
	Context      context.Context
}

func (creator *ClusterService) Init(manifestDir string) error {
	creator.ManifestDir = CON.GrantClusterManifestDir(manifestDir)
	ctx := context.TODO()
	creator.Context = ctx
	err := runTerraformInit(ctx, creator.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (creator *ClusterService) Apply(createArgs *ClusterCreationArgs, recordtfvars bool, tfvarsDeletion bool, extraArgs ...string) error {
	createArgs.URL = CON.GateWayURL
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	if recordtfvars {
		recordTFvarsFile(creator.ManifestDir, tfvars) // Record the tfvars before apply in case cluster creation error and we need clean
	}

	_, err := runTerraformApplyWithArgs(creator.Context, creator.ManifestDir, args)
	if err != nil {
		if tfvarsDeletion {
			deleteTFvarsFile(creator.ManifestDir)
		}
		return err
	}
	return nil
}

func (creator *ClusterService) Output() (*ClusterOutput, error) {
	out, err := runTerraformOutput(creator.Context, creator.ManifestDir)
	if err != nil {
		return nil, err
	}
	clusterOutput := &ClusterOutput{
		ClusterID:                            h.DigString(out["cluster_id"], "value"),
		AdditionalComputeSecurityGroups:      h.DigArrayToString(out["additional_compute_security_groups"], "value"),
		AdditionalInfraSecurityGroups:        h.DigArrayToString(out["additional_infra_security_groups"], "value"),
		AdditionalControlPlaneSecurityGroups: h.DigArrayToString(out["additional_control_plane_security_groups"], "value"),
	}

	return clusterOutput, nil
}

func (creator *ClusterService) Destroy(createArgs *ClusterCreationArgs, extraArgs ...string) (string, error) {
	createArgs.URL = CON.GateWayURL
	args, _ := combineStructArgs(createArgs, extraArgs...)
	return runTerraformDestroyWithArgs(creator.Context, creator.ManifestDir, args)
}

func (creator *ClusterService) Plan(planargs *ClusterCreationArgs, extraArgs ...string) (string, error) {
	planargs.URL = CON.GateWayURL
	args, _ := combineStructArgs(planargs, extraArgs...)
	output, err := runTerraformPlanWithArgs(creator.Context, creator.ManifestDir, args)
	return output, err
}

func NewClusterService(manifestDir string) (*ClusterService, error) {
	sc := &ClusterService{}
	err := sc.Init(manifestDir)
	return sc, err
}
