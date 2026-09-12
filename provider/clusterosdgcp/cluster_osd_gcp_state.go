/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusterosdgcp

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Object types for nested attributes that can receive unknown values (e.g. from variables during terraform validate).
// Used by schema ObjectAttribute and when populating state from API.
//
//nolint:unused // Used by Terraform framework for (de)serialization
var (
	networkObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"machine_cidr": types.StringType,
			"service_cidr": types.StringType,
			"pod_cidr":     types.StringType,
			"host_prefix":  types.Int64Type,
		},
	}
	gcpNetworkObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"vpc_name":             types.StringType,
			"vpc_project_id":       types.StringType,
			"compute_subnet":       types.StringType,
			"control_plane_subnet": types.StringType,
		},
	}
	securityObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"secure_boot": types.BoolType,
		},
	}
	privateServiceConnectObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"service_attachment_subnet": types.StringType,
		},
	}
	autoscalingObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_replicas": types.Int64Type,
			"max_replicas": types.Int64Type,
		},
	}
	proxyObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"http_proxy":              types.StringType,
			"https_proxy":             types.StringType,
			"no_proxy":                types.StringType,
			"additional_trust_bundle": types.StringType,
		},
	}
)

// ClusterOsdGcpState holds the Terraform state for an OSD cluster on GCP.
type ClusterOsdGcpState struct {
	ID         types.String `tfsdk:"id"`
	ExternalID types.String `tfsdk:"external_id"`

	Name         types.String `tfsdk:"name"`
	CloudRegion  types.String `tfsdk:"cloud_region"`
	GCPProjectID types.String `tfsdk:"gcp_project_id"`
	Product      types.String `tfsdk:"product"`
	MultiAZ      types.Bool   `tfsdk:"multi_az"`

	Version      types.String `tfsdk:"version"`
	DomainPrefix types.String `tfsdk:"domain_prefix"`
	CCSEnabled   types.Bool   `tfsdk:"ccs_enabled"`

	BillingModel        types.String `tfsdk:"billing_model"`
	MarketplaceGCPTerms types.Bool   `tfsdk:"marketplace_gcp_terms"`

	ComputeMachineType types.String `tfsdk:"compute_machine_type"`
	ComputeNodes       types.Int64  `tfsdk:"compute_nodes"`
	AvailabilityZones  types.List   `tfsdk:"availability_zones"`
	Properties         types.Map    `tfsdk:"properties"`

	WIFConfigID             types.String `tfsdk:"wif_config_id"`
	WifVerifyTimeoutMinutes types.Int64  `tfsdk:"wif_verify_timeout_minutes"`

	WaitForCreateComplete types.Bool  `tfsdk:"wait_for_create_complete"`
	WaitTimeout           types.Int64 `tfsdk:"wait_timeout"`

	// GCP authentication (service account key) - used when not using WIF
	GCPAuthentication *GCPAuthenticationState `tfsdk:"gcp_authentication"`

	// Private Service Connect (types.Object supports null/unknown during validate)
	PrivateServiceConnect types.Object `tfsdk:"private_service_connect"`

	// GCP network (Shared VPC) - types.Object supports null/unknown during validate
	GCPNetwork types.Object `tfsdk:"gcp_network"`

	// CMEK encryption
	GCPEncryptionKey *GCPEncryptionKeyState `tfsdk:"gcp_encryption_key"`

	// GCP security (types.Object supports null/unknown during validate)
	Security types.Object `tfsdk:"security"`

	// Network CIDRs (types.Object supports null/unknown during validate)
	Network types.Object `tfsdk:"network"`

	// Autoscaling (types.Object supports null/unknown during validate)
	Autoscaling types.Object `tfsdk:"autoscaling"`

	// Proxy (types.Object supports null/unknown during validate)
	Proxy types.Object `tfsdk:"proxy"`

	// Private cluster: API endpoint and ingress use internal listening only.
	// Requires BYO VPC (gcp_network) and Private Service Connect (private_service_connect).
	// Cannot be changed after cluster creation.
	Private types.Bool `tfsdk:"private"`

	// Inline cluster-admin user (mirrors rhcs_cluster_rosa_classic). When set,
	// an htpasswd identity provider with a cluster-admin user is created as part
	// of the cluster, no separate resource needed.
	CreateAdminUser  types.Bool   `tfsdk:"create_admin_user"`
	AdminCredentials types.Object `tfsdk:"admin_credentials"`

	// Computed
	State          types.String `tfsdk:"state"`
	APIURL         types.String `tfsdk:"api_url"`
	ConsoleURL     types.String `tfsdk:"console_url"`
	Domain         types.String `tfsdk:"domain"`
	InfraID        types.String `tfsdk:"infra_id"`
	CurrentCompute types.Int64  `tfsdk:"current_compute"`
}

// GCPAuthenticationState holds service account key auth (used when not using WIF).
type GCPAuthenticationState struct {
	ClientEmail             types.String `tfsdk:"client_email"`
	ClientID                types.String `tfsdk:"client_id"`
	PrivateKey              types.String `tfsdk:"private_key"`
	PrivateKeyID            types.String `tfsdk:"private_key_id"`
	AuthURI                 types.String `tfsdk:"auth_uri"`
	TokenURI                types.String `tfsdk:"token_uri"`
	AuthProviderX509CertURL types.String `tfsdk:"auth_provider_x509_cert_url"`
	ClientX509CertURL       types.String `tfsdk:"client_x509_cert_url"`
	Type                    types.String `tfsdk:"type"`
}

// PrivateServiceConnectState holds PSC configuration.
type PrivateServiceConnectState struct {
	ServiceAttachmentSubnet types.String `tfsdk:"service_attachment_subnet"`
}

// GCPNetworkState holds Shared VPC configuration.
type GCPNetworkState struct {
	VPCName            types.String `tfsdk:"vpc_name"`
	VPCProjectID       types.String `tfsdk:"vpc_project_id"`
	ComputeSubnet      types.String `tfsdk:"compute_subnet"`
	ControlPlaneSubnet types.String `tfsdk:"control_plane_subnet"`
}

// GCPEncryptionKeyState holds CMEK configuration.
type GCPEncryptionKeyState struct {
	KmsKeyServiceAccount types.String `tfsdk:"kms_key_service_account"`
	KeyLocation          types.String `tfsdk:"key_location"`
	KeyName              types.String `tfsdk:"key_name"`
	KeyRing              types.String `tfsdk:"key_ring"`
}

// GcpSecurityState holds GCP security settings.
type GcpSecurityState struct {
	SecureBoot types.Bool `tfsdk:"secure_boot"`
}

// NetworkState holds network CIDRs.
type NetworkState struct {
	MachineCIDR types.String `tfsdk:"machine_cidr"`
	ServiceCIDR types.String `tfsdk:"service_cidr"`
	PodCIDR     types.String `tfsdk:"pod_cidr"`
	HostPrefix  types.Int64  `tfsdk:"host_prefix"`
}

// AutoscalingState holds autoscaling configuration.
type AutoscalingState struct {
	MinReplicas types.Int64 `tfsdk:"min_replicas"`
	MaxReplicas types.Int64 `tfsdk:"max_replicas"`
}

// ProxyState holds proxy configuration.
type ProxyState struct {
	HTTPProxy             types.String `tfsdk:"http_proxy"`
	HTTPSProxy            types.String `tfsdk:"https_proxy"`
	NoProxy               types.String `tfsdk:"no_proxy"`
	AdditionalTrustBundle types.String `tfsdk:"additional_trust_bundle"`
}
