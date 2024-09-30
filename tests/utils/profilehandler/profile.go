package profilehandler

type Profile struct {
	Name                    string   `ini:"name,omitempty" json:"name,omitempty"`
	ClusterName             string   `ini:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	DomainPrefix            string   `ini:"domain_prefix,omitempty" json:"domain_prefix,omitempty"`
	ClusterType             string   `ini:"cluster_type,omitempty" json:"cluster_type,omitempty"`
	ProductID               string   `ini:"product_id,omitempty" json:"product_id,omitempty"`
	MajorVersion            string   `ini:"major_version,omitempty" json:"major_version,omitempty"`
	Version                 string   `ini:"version,omitempty" json:"version,omitempty"`                 //Specific OCP version to be specified
	VersionPattern          string   `ini:"version_pattern,omitempty" json:"version_pattern,omitempty"` //Version supports indicated version started with openshift-v or major-1 (y-1) or minor-1 (z-1)
	ChannelGroup            string   `ini:"channel_group,omitempty" json:"channel_group,omitempty"`
	CloudProvider           string   `ini:"cloud_provider,omitempty" json:"cloud_provider,omitempty"`
	Region                  string   `ini:"region,omitempty" json:"region,omitempty"`
	InstanceType            string   `ini:"instance_type,omitempty" json:"instance_type,omitempty"`
	Zones                   string   `ini:"zones,omitempty" json:"zones,omitempty"`           // zones should be like a,b,c,d
	StorageLB               bool     `ini:"storage_lb,omitempty" json:"storage_lb,omitempty"` // the unit is GIB, don't support unit set
	Tagging                 bool     `ini:"tagging,omitempty" json:"tagging,omitempty"`
	Labeling                bool     `ini:"labeling,omitempty" json:"labeling,omitempty"`
	Etcd                    bool     `ini:"etcd_encryption,omitempty" json:"etcd_encryption,omitempty"`
	FIPS                    bool     `ini:"fips,omitempty" json:"fips,omitempty"`
	CCS                     bool     `ini:"ccs,omitempty" json:"ccs,omitempty"`
	STS                     bool     `ini:"sts,omitempty" json:"sts,omitempty"`
	Autoscale               bool     `ini:"autoscaling_enabled,omitempty" json:"autoscaling_enabled,omitempty"`
	MultiAZ                 bool     `ini:"multi_az,omitempty" json:"multi_az,omitempty"`
	BYOVPC                  bool     `ini:"byovpc,omitempty" json:"byovpc,omitempty"`
	PrivateLink             bool     `ini:"private_link,omitempty" json:"private_link,omitempty"`
	Private                 bool     `ini:"private,omitempty" json:"private,omitempty"`
	BYOK                    bool     `ini:"byok,omitempty" json:"byok,omitempty"`
	KMSKey                  bool     `ini:"kms_key_arn,omitempty" json:"kms_key_arn,omitempty"`
	DifferentEncryptionKeys bool     `ini:"different_encryption_keys,omitempty" json:"different_encryption_keys,omitempty"`
	NetWorkingSet           bool     `ini:"networking_set,omitempty" json:"networking_set,omitempty"`
	Proxy                   bool     `ini:"proxy,omitempty" json:"proxy,omitempty"`
	OIDCConfig              string   `ini:"oidc_config,omitempty" json:"oidc_config,omitempty"`
	ProvisionShard          string   `ini:"provisionShard,omitempty" json:"provisionShard,omitempty"`
	Ec2MetadataHttpTokens   string   `ini:"ec2_metadata_http_tokens,omitempty" json:"ec2_metadata_http_tokens,omitempty"`
	ComputeReplicas         int      `ini:"compute_replicas,omitempty" json:"compute_replicas,omitempty"`
	ComputeMachineType      string   `ini:"compute_machine_type,omitempty" json:"compute_machine_type,omitempty"`
	AuditLogForward         bool     `ini:"auditlog_forward,omitempty" json:"auditlog_forward,omitempty"`
	AdminEnabled            bool     `ini:"admin_enabled,omitempty" json:"admin_enabled,omitempty"`
	ManagedPolicies         bool     `ini:"managed_policies,omitempty" json:"managed_policies,omitempty"`
	WorkerDiskSize          int      `ini:"worker_disk_size,omitempty" json:"worker_disk_size,omitempty"`
	AdditionalSGNumber      int      `ini:"additional_sg_number,omitempty" json:"additional_sg_number,omitempty"`
	UnifiedAccRolesPath     string   `ini:"unified_acc_role_path,omitempty" json:"unified_acc_role_path,omitempty"`
	SharedVpc               bool     `ini:"shared_vpc,omitempty" json:"shared_vpc,omitempty"`
	MachineCIDR             string   `ini:"machine_cidr,omitempty" json:"machine_cidr,omitempty"`
	ServiceCIDR             string   `ini:"service_cidr,omitempty" json:"service_cidr,omitempty"`
	PodCIDR                 string   `ini:"pod_cidr,omitempty" json:"pod_cidr,omitempty"`
	HostPrefix              int      `ini:"host_prefix,omitempty" json:"host_prefix,omitempty"`
	FullResources           bool     `ini:"full_resources,omitempty" json:"full_resources,omitempty"`
	DontWaitForCluster      bool     `ini:"no_wait_cluster,omitempty" json:"no_wait_cluster,omitempty"`
	UseRegistryConfig       bool     `ini:"use_registry_config,omitempty" json:"use_registry_config,omitempty"`
	AllowedRegistries       []string `ini:"allowed_registries,omitempty" json:"allowed_registries,omitempty"`
	BlockedRegistries       []string `ini:"blocked_registries,omitempty" json:"blocked_registries,omitempty"`
}
