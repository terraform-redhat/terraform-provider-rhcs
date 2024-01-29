package hcp

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/sts"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"
)

type ClusterRosaHcpState struct {
	APIURL              types.String `tfsdk:"api_url"`
	AvailabilityZones   types.List   `tfsdk:"availability_zones"`
	AWSAccountID        types.String `tfsdk:"aws_account_id"`
	AWSBillingAccountID types.String `tfsdk:"aws_billing_account_id"`
	AWSSubnetIDs        types.List   `tfsdk:"aws_subnet_ids"`
	AWSPrivateLink      types.Bool   `tfsdk:"aws_private_link"`
	Sts                 *sts.HcpSts  `tfsdk:"sts"`
	EtcdEncryption      types.Bool   `tfsdk:"etcd_encryption"`
	AutoScalingEnabled  types.Bool   `tfsdk:"autoscaling_enabled"`
	ChannelGroup        types.String `tfsdk:"channel_group"`
	CloudRegion         types.String `tfsdk:"cloud_region"`
	ComputeMachineType  types.String `tfsdk:"compute_machine_type"`
	Replicas            types.Int64  `tfsdk:"replicas"`
	ConsoleURL          types.String `tfsdk:"console_url"`
	Domain              types.String `tfsdk:"domain"`
	ID                  types.String `tfsdk:"id"`
	KMSKeyArn           types.String `tfsdk:"kms_key_arn"`
	ExternalID          types.String `tfsdk:"external_id"`
	Name                types.String `tfsdk:"name"`
	PodCIDR             types.String `tfsdk:"pod_cidr"`
	MachineCIDR         types.String `tfsdk:"machine_cidr"`
	ServiceCIDR         types.String `tfsdk:"service_cidr"`
	HostPrefix          types.Int64  `tfsdk:"host_prefix"`
	Properties          types.Map    `tfsdk:"properties"`
	OCMProperties       types.Map    `tfsdk:"ocm_properties"`
	Tags                types.Map    `tfsdk:"tags"`
	Proxy               *proxy.Proxy `tfsdk:"proxy"`
	State               types.String `tfsdk:"state"`
	Version             types.String `tfsdk:"version"`
	CurrentVersion      types.String `tfsdk:"current_version"`

	UpgradeAcksFor types.String `tfsdk:"upgrade_acknowledgements_for"`

	DisableWaitingInDestroy types.Bool  `tfsdk:"disable_waiting_in_destroy"`
	DestroyTimeout          types.Int64 `tfsdk:"destroy_timeout"`
	WaitForCreateComplete   types.Bool  `tfsdk:"wait_for_create_complete"`
}
