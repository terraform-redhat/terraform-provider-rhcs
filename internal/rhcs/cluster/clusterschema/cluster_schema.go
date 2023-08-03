package clusterschema

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ClusterFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"product": {
			Description: "Product ID OSD or Rosa",
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: "Name of the cluster.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"cloud_provider": {
			Description: "Cloud clusterservice identifier, for example 'aws'.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"cloud_region": {
			Description: "Cloud region identifier, for example 'us-east-1'.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"multi_az": {
			Description: "Indicates if the cluster should be deployed to " +
				"multiple availability zones. Default value is 'false'.",
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"properties": {
			Description: "User defined properties.",
			Type:        schema.TypeMap,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Computed:    true,
		},
		"api_url": {
			Description: "URL of the API server.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"console_url": {
			Description: "URL of the console.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"compute_nodes": {
			Description: "Number of compute nodes of the cluster.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"compute_machine_type": {
			Description: "Identifier of the machine type used by the compute nodes, " +
				"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
				"source to find the possible values.",
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"ccs_enabled": {
			Description: "Enables customer cloud subscription.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"aws_account_id": {
			Description: "Identifier of the AWS account.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"aws_access_key_id": {
			Description: "Identifier of the AWS access key.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"aws_secret_access_key": {
			Description: "AWS access key.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"aws_subnet_ids": {
			Description: "aws subnet ids",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"aws_private_link": {
			Description: "aws subnet ids",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"availability_zones": {
			Description: "availability zones",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"machine_cidr": {
			Description: "Block of IP addresses for nodes.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"proxy": {
			Description: "proxy",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: ProxyFields(),
			},
			Optional: true,
		},
		"service_cidr": {
			Description: "Block of IP addresses for services.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"pod_cidr": {
			Description: "Block of IP addresses for pods.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"host_prefix": {
			Description: "Length of the prefix of the subnet assigned to each node.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"version": {
			Description: "Identifier of the version of OpenShift, for example 'openshift-v4.1.0'.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"state": {
			Description: "State of the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"wait": {
			Description: "Wait till the cluster is ready.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
	}
}

type ClusterState struct {
	// required
	Product       string `tfsdk:"product"`
	Name          string `tfsdk:"name"`
	CloudProvider string `tfsdk:"cloud_provider"`
	CloudRegion   string `tfsdk:"cloud_region"`

	// optional
	MultiAZ            *bool             `tfsdk:"multi_az"`
	Properties         map[string]string `tfsdk:"properties"`
	APIURL             *string           `tfsdk:"api_url"`
	ConsoleURL         *string           `tfsdk:"console_url"`
	ComputeNodes       *int              `tfsdk:"compute_nodes"`
	ComputeMachineType *string           `tfsdk:"compute_machine_type"`
	CCSEnabled         *bool             `tfsdk:"ccs_enabled"`
	AWSAccountID       *string           `tfsdk:"aws_account_id"`
	AWSAccessKeyID     *string           `tfsdk:"aws_access_key_id"`
	AWSSecretAccessKey *string           `tfsdk:"aws_secret_access_key"`
	AWSSubnetIDs       []string          `tfsdk:"aws_subnet_ids"`
	AWSPrivateLink     *bool             `tfsdk:"aws_private_link"`
	AvailabilityZones  []string          `tfsdk:"availability_zones"`
	MachineCIDR        *string           `tfsdk:"machine_cidr"`
	Proxy              *Proxy            `tfsdk:"proxy"`
	ServiceCIDR        *string           `tfsdk:"service_cidr"`
	PodCIDR            *string           `tfsdk:"pod_cidr"`
	HostPrefix         *int              `tfsdk:"host_prefix"`
	Version            *string           `tfsdk:"version"`
	Wait               *bool             `tfsdk:"wait"`

	// computed
	ID    string `tfsdk:"id"`
	State string `tfsdk:"state"`
}
