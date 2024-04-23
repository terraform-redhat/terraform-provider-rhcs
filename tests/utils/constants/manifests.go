package constants

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func initManifestsDir() string {
	if ManifestsDirENV != "" {
		return ManifestsDirENV
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
	AWSProviderDir  = "aws"
	RHCSProviderDir = "rhcs"
	MachinePoolsDir = "machine-pools"
	IngressDir      = "ingresses"
)

// Dirs of aws provider
var (
	AccountRolesClassicDir                      = path.Join(ManifestsConfigurationDir, AWSProviderDir, "account-roles", "rosa-classic")
	AccountRolesHCPDir                          = path.Join(ManifestsConfigurationDir, AWSProviderDir, "account-roles", "rosa-hcp")
	AddAccountRolesClassicDir                   = path.Join(ManifestsConfigurationDir, AWSProviderDir, "add-account-roles", "rosa-classic")
	AddAccountRolesHCPDir                       = path.Join(ManifestsConfigurationDir, AWSProviderDir, "add-account-roles", "rosa-hcp")
	OIDCProviderOperatorRolesClassicManifestDir = path.Join(ManifestsConfigurationDir, AWSProviderDir, "oidc-provider-operator-roles", "rosa-classic")
	OIDCProviderOperatorRolesHCPManifestDir     = path.Join(ManifestsConfigurationDir, AWSProviderDir, "oidc-provider-operator-roles", "rosa-hcp")
	AWSVPCDir                                   = path.Join(ManifestsConfigurationDir, AWSProviderDir, "vpc")
	AWSVPCTagDir                                = path.Join(ManifestsConfigurationDir, AWSProviderDir, "vpc-tags")
	AWSSecurityGroupDir                         = path.Join(ManifestsConfigurationDir, AWSProviderDir, "security-groups")
	ProxyDir                                    = path.Join(ManifestsConfigurationDir, AWSProviderDir, "proxy")
	KMSDir                                      = path.Join(ManifestsConfigurationDir, AWSProviderDir, "kms")
	SharedVpcPolicyAndHostedZoneDir             = path.Join(ManifestsConfigurationDir, AWSProviderDir, "shared-vpc-policy-and-hosted-zone")
)

func GetAccountRoleDefaultManifestDir(clusterType ClusterType) string {
	if clusterType.HCP {
		return AccountRolesHCPDir
	} else {
		return AccountRolesClassicDir
	}
}

func GetAddAccountRoleDefaultManifestDir(clusterType ClusterType) string {
	if clusterType.HCP {
		return AddAccountRolesHCPDir
	} else {
		return AddAccountRolesClassicDir
	}
}

func GetOIDCProviderOperatorRolesDefaultManifestDir(clusterType ClusterType) string {
	if clusterType.HCP {
		return OIDCProviderOperatorRolesHCPManifestDir
	} else {
		return OIDCProviderOperatorRolesClassicManifestDir
	}
}

// Dirs of rhcs provider
var (
	ClusterDir                  = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "clusters")
	ImportResourceDir           = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "resource-import")
	IDPsDir                     = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "idps")
	ClassicMachinePoolDir       = path.Join(ManifestsConfigurationDir, RHCSProviderDir, MachinePoolsDir, "classic")
	HCPMachinePoolDir           = path.Join(ManifestsConfigurationDir, RHCSProviderDir, MachinePoolsDir, "hcp")
	DNSDir                      = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "dns")
	ClassicIngressDir           = path.Join(ManifestsConfigurationDir, RHCSProviderDir, IngressDir, "classic")
	HCPIngressDir               = path.Join(ManifestsConfigurationDir, RHCSProviderDir, IngressDir, "hcp")
	RhcsInfoDir                 = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "rhcs-info")
	DefaultMachinePoolDir       = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "default-machine-pool")
	KubeletConfigDir            = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "kubelet-config")
	ClassicClusterAutoscalerDir = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "cluster-autoscaler", "classic")
	HCPClusterAutoscalerDir     = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "cluster-autoscaler", "hcp")
	TuningConfigDir             = path.Join(ManifestsConfigurationDir, RHCSProviderDir, "tuning-config")
)

func GetClusterManifestsDir(clusterType ClusterType) string {
	return path.Join(ClusterDir, clusterType.String())
}

// Supports abs and relatives
func GrantClusterManifestDir(manifestDir string) string {
	var targetDir string
	if strings.Contains(manifestDir, ClusterDir) {
		targetDir = manifestDir
	} else {
		targetDir = path.Join(ClusterDir, manifestDir)
	}
	return targetDir
}
