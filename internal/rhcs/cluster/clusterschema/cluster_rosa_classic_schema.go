package clusterschema

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/build"
)

const (
	AwsCloudProvider      = "aws"
	RosaProduct           = "rosa"
	MinVersion            = "4.10.0"
	MaxClusterNameLength  = 15
	tagsPrefix            = "rosa_"
	tagsOpenShiftVersion  = tagsPrefix + "openshift_version"
	LowestHttpTokensVer   = "4.11.0"
	propertyRosaTfVersion = tagsPrefix + "tf_version"
	propertyRosaTfCommit  = tagsPrefix + "tf_commit"
	ErrHeadline           = "Can't build cluster"
)

var OCMProperties = map[string]string{
	propertyRosaTfVersion: build.Version,
	propertyRosaTfCommit:  build.Commit,
}

func ClusterRosaClassicFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "Name of the cluster. Cannot exceed 15 characters in length.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"cloud_region": {
			Description: "Cloud region identifier, for example 'us-east-1'.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"external_id": {
			Description: "Unique external identifier of the cluster.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"multi_az": {
			Description: "Indicates if the cluster should be deployed to " +
				"multiple availability zones. Default value is 'false'.",
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"disable_scp_checks": {
			Description: "Enables you to monitor your own projects in isolation from Red Hat " +
				"Site Reliability Engineer (SRE) platform metrics.",
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"disable_workload_monitoring": {
			Description: "Enables you to monitor your own projects in isolation from Red Hat " +
				"Site Reliability Engineer (SRE) platform metrics.",
			Type:     schema.TypeBool,
			Optional: true,
		},
		"sts": {
			Description: "STS configuration.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: StsFields(),
			},
			Optional: true,
		},
		"properties": {
			Description:      "User defined properties.",
			Type:             schema.TypeMap,
			Elem:             &schema.Schema{Type: schema.TypeString},
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: PropertiesValidators,
		},
		"ocm_properties": {
			Description: "Merged properties defined by OCM and the user defined 'properties'.",
			Type:        schema.TypeMap,
			Elem:        &schema.Schema{Type: schema.TypeString},
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
		"tags": {
			Description: "Apply user defined tags to all resources created in AWS.",
			Type:        schema.TypeMap,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			ForceNew:    true,
		},
		"replicas": {
			Description: "Number of worker nodes to provision. Single zone clusters need at least 2 nodes, " +
				"multizone clusters need at least 3 nodes.",
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"autoscaling_enabled", "min_replicas", "max_replicas"},
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
			Computed:    true,
		},
		"etcd_encryption": {
			Description: "Encrypt etcd data.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"domain": {
			Description: "DNS domain of cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"autoscaling_enabled": {
			Description:   "Enables autoscaling.",
			Type:          schema.TypeBool,
			Optional:      true,
			ConflictsWith: []string{"replicas"},
			RequiredWith:  []string{"min_replicas", "max_replicas"},
		},
		"min_replicas": {
			Description:   "Minimum replicas.",
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"replicas"},
			RequiredWith:  []string{"min_replicas", "autoscaling_enabled"},
		},
		"max_replicas": {
			Description:   "Maximum replicas.",
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"replicas"},
			RequiredWith:  []string{"max_replicas", "autoscaling_enabled"},
		},
		"aws_account_id": {
			Description: "Identifier of the AWS account.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
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
		"private": {
			Description: "Restrict master API endpoint and application routes to direct, private connectivity.",
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
			Computed:    true,
			ForceNew:    true,
		},
		"machine_cidr": {
			Description: "Block of IP addresses for nodes.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
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
			ForceNew:    true,
		},
		"pod_cidr": {
			Description: "Block of IP addresses for pods.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"host_prefix": {
			Description: "Length of the prefix of the subnet assigned to each node.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"version": {
			Description: "Desired version of OpenShift for the cluster, for example '4.1.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"state": {
			Description: "State of the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"default_mp_labels": {
			Description: "This value is the default machine pool labels. Format should be a comma-separated list of '{\"key1\"=\"value1\", \"key2\"=\"value2\"}'. " +
				"This list overwrites any modifications made to Node labels on an ongoing basis. ",
			Type:     schema.TypeMap,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
		"kms_key_arn": {
			Description: "The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
				"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID.",
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"fips": {
			Description: "Create cluster that uses FIPS Validated / Modules in Process cryptographic libraries.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"channel_group": {
			Description: "Name of the channel group where you select the OpenShift cluster version, for example 'stable'.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"current_version": {
			Description: "The currently running version of OpenShift on the cluster, for example '4.1.0'.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"disable_waiting_in_destroy": {
			Description: "Disable addressing cluster state in the destroy resource. Default value is false.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"destroy_timeout": {
			Description: "This value sets the maximum duration in minutes to allow for destroying resources. Default value is 60 minutes.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"ec2_metadata_http_tokens": {
			Description: "This value determines which EC2 metadata mode to use for metadata service interaction " +
				"options for EC2 instances can be optional or required. Required is available from " +
				"OpenShift version 4.11.0 and newer.",
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
				[]string{string(cmv1.Ec2MetadataHttpTokensOptional), string(cmv1.Ec2MetadataHttpTokensRequired)},
				false)),
		},
		"upgrade_acknowledgements_for": {
			Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
				" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
				"upgrade to OpenShift 4.12.z from 4.11 or before).",
			Type:     schema.TypeString,
			Optional: true,
		},
		"admin_credentials": {
			Description: "Admin user credentials",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: AdminFields(),
			},
			Optional: true,
			ForceNew: true,
		},
	}
}

type ClusterRosaClassicState struct {
	// required
	Name         string `tfsdk:"name"`
	CloudRegion  string `tfsdk:"cloud_region"`
	AWSAccountID string `tfsdk:"aws_account_id"`

	// optional
	ExternalID                *string           `tfsdk:"external_id"`
	MultiAZ                   *bool             `tfsdk:"multi_az"`
	DisableSCPChecks          *bool             `tfsdk:"disable_scp_checks"`
	DisableWorkloadMonitoring *bool             `tfsdk:"disable_workload_monitoring"`
	Sts                       *Sts              `tfsdk:"sts"`
	Properties                map[string]string `tfsdk:"properties"`
	Tags                      map[string]string `tfsdk:"tags"`
	Replicas                  *int              `tfsdk:"replicas"`
	ComputeMachineType        *string           `tfsdk:"compute_machine_type"`
	EtcdEncryption            *bool             `tfsdk:"etcd_encryption"`
	AutoScalingEnabled        *bool             `tfsdk:"autoscaling_enabled"`
	MinReplicas               *int              `tfsdk:"min_replicas"`
	MaxReplicas               *int              `tfsdk:"max_replicas"`
	AWSSubnetIDs              []string          `tfsdk:"aws_subnet_ids"`
	AWSPrivateLink            *bool             `tfsdk:"aws_private_link"`
	Private                   *bool             `tfsdk:"private"`
	AvailabilityZones         []string          `tfsdk:"availability_zones"`
	MachineCIDR               *string           `tfsdk:"machine_cidr"`
	Proxy                     *Proxy            `tfsdk:"proxy"`
	ServiceCIDR               *string           `tfsdk:"service_cidr"`
	PodCIDR                   *string           `tfsdk:"pod_cidr"`
	HostPrefix                *int              `tfsdk:"host_prefix"`
	Version                   *string           `tfsdk:"version"`
	DefaultMPLabels           map[string]string `tfsdk:"default_mp_labels"`
	KMSKeyArn                 *string           `tfsdk:"kms_key_arn"`
	FIPS                      *bool             `tfsdk:"fips"`
	ChannelGroup              *string           `tfsdk:"channel_group"`
	DisableWaitingInDestroy   *bool             `tfsdk:"disable_waiting_in_destroy"`
	DestroyTimeout            *int              `tfsdk:"destroy_timeout"`
	Ec2MetadataHttpTokens     *string           `tfsdk:"ec2_metadata_http_tokens"`
	UpgradeAcksFor            *string           `tfsdk:"upgrade_acknowledgements_for"`
	AdminCredentials          *AdminCredentials `tfsdk:"admin_credentials"`

	// computed
	ID             string            `tfsdk:"id"`
	State          string            `tfsdk:"state"`
	OCMProperties  map[string]string `tfsdk:"ocm_properties"`
	CCSEnabled     bool              `tfsdk:"ccs_enabled"`
	CurrentVersion string            `tfsdk:"current_version"`
	APIURL         string            `tfsdk:"api_url"`
	ConsoleURL     string            `tfsdk:"console_url"`
	Domain         string            `tfsdk:"domain"`
}

func PropertiesValidators(i interface{}, path cty.Path) diag.Diagnostics {
	propertiesMap, ok := i.(map[string]interface{})
	if !ok {
		return diag.Errorf("expected type to be map of string")
	}

	for k, _ := range propertiesMap {
		if _, isDefaultKey := OCMProperties[k]; isDefaultKey {
			return diag.Errorf(fmt.Sprintf("Can not override reserved properties keys. `%s` is a reserved property key"), k)
		}
	}

	return nil
}
