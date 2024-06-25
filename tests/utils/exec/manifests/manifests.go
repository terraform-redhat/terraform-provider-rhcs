package manifests

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

func initManifestsDir() string {
	if constants.ManifestsDirENV != "" {
		return constants.ManifestsDirENV
	}
	currentDir, _ := os.Getwd()
	manifestsDir := path.Join(strings.SplitAfter(currentDir, "tests")[0], "tf-manifests")
	if _, err := os.Stat(manifestsDir); err != nil {
		panic(fmt.Sprintf("Manifests dir %s doesn't exist. Make sure you have the manifests dir in testing repo or set the correct env MANIFESTS_DIR value", manifestsDir))
	}
	return manifestsDir
}

var ManifestsConfigurationDir = initManifestsDir()

// Provider dirs' name definition
const (
	awsProviderDir  = "aws"
	rhcsProviderDir = "rhcs"
	idpsDir         = "idps"
)

// AWS dirs
func GetAWSAccountRolesManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "account-roles", clusterType.String())
}

func GetAWSKMSManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "kms")
}

func GetAWSOIDCProviderOperatorRolesManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "oidc-provider-operator-roles", clusterType.String())
}

func GetAWSProxyManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "proxy")
}

func GetAWSSecurityGroupManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "security-groups")
}

func GetAWSSharedVPCPolicyAndHostedZoneManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "shared-vpc-policy-and-hosted-zone")
}

func GetAWSVPCManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "vpc", clusterType.String())
}

func GetAWSVPCTagManifestDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, awsProviderDir, "vpc-tags")
}

// RHCS provider dirs
func GetClusterManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "clusters", clusterType.String())
}

func GetClusterWaiterManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "cluster-waiter")
}

func GetClusterAutoscalerManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "cluster-autoscaler", clusterType.String())
}

func GetDnsDomainManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "dns")
}

func GetIDPManifestsDir(clusterType constants.ClusterType, idpType constants.IDPType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, idpsDir, string(idpType))
}

func GetIngressManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "ingresses", clusterType.String())
}

func GetImportManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "resource-import")
}

func GetKubeletConfigManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "kubelet-config")
}

func GetMachinePoolsManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "machine-pools", clusterType.String())
}

func GetRHCSInfoManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "rhcs-info")
}

func GetTuningConfigManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "tuning-config")
}

func GetTrustedIPsManifestsDir(clusterType constants.ClusterType) string {
	return path.Join(ManifestsConfigurationDir, rhcsProviderDir, "trusted-ips")
}
