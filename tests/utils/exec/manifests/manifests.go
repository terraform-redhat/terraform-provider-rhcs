package manifests

import (
	"path"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

// Provider dirs' name definition
const (
	awsProviderDir  = "aws"
	rhcsProviderDir = "rhcs"
	idpsDir         = "idps"
)

// AWS dirs
func GetAWSAccountRolesManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "account-roles", clusterType.String())
}

func GetAWSKMSManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "kms")
}

func GetAWSOIDCProviderOperatorRolesManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "oidc-provider-operator-roles", clusterType.String())
}

func GetAWSProxyManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "proxy")
}

func GetAWSSecurityGroupManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "security-groups")
}

func GetAWSSharedVPCPolicyAndHostedZoneManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "shared-vpc-policy-and-hosted-zone")
}

func GetAWSVPCManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "vpc", clusterType.String())
}

func GetAWSVPCTagManifestDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), awsProviderDir, "vpc-tags")
}

// RHCS provider dirs
func GetClusterManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "clusters", clusterType.String())
}

func GetClusterWaiterManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "cluster-waiter")
}

func GetClusterAutoscalerManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "cluster-autoscaler", clusterType.String())
}

func GetDnsDomainManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "dns")
}

func GetIDPManifestsDir(clusterType constants.ClusterType, idpType constants.IDPType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, idpsDir, string(idpType))
}

func GetIngressManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "ingresses", clusterType.String())
}

func GetImportManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "resource-import")
}

func GetKubeletConfigManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "kubelet-config")
}

func GetMachinePoolsManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "machine-pools", clusterType.String())
}

func GetRHCSInfoManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "rhcs-info")
}

func GetTuningConfigManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "tuning-config")
}

func GetTrustedIPsManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "trusted-ips")
}

func GetImageMirrorManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "image-mirrors", clusterType.String())
}

func GetBreakGlassCredentialManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(config.GetManifestsDir(), rhcsProviderDir, "break-glass-credentials")
}
