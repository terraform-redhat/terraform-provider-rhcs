package exec

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

var defaultManifestsConfigurationDir string

func init() {
	if constants.ManifestsDirENV != "" {
		defaultManifestsConfigurationDir = constants.ManifestsDirENV
	}
	currentDir, _ := os.Getwd()
	defaultManifestsConfigurationDir = path.Join(strings.SplitAfter(currentDir, "tests")[0], "tf-manifests")
	if _, err := os.Stat(defaultManifestsConfigurationDir); err != nil {
		panic(fmt.Sprintf("Manifests dir %s doesn't exist. Make sure you have the manifests dir in testing repo or set the correct env MANIFESTS_DIR value", defaultManifestsConfigurationDir))
	}

	err := AlignRHCSSourceVersion(defaultManifestsConfigurationDir)
	if err != nil {
		panic(err)
	}
}

// Provider dirs' name definition
const (
	awsProviderDir  = "aws"
	rhcsProviderDir = "rhcs"
)

// Dirs of aws provider
var (
	accountRolesDir                      = path.Join(awsProviderDir, "account-roles")
	oidcProviderOperatorRolesManifestDir = path.Join(awsProviderDir, "oidc-provider-operator-roles")
	awsVPCDir                            = path.Join(awsProviderDir, "vpc")
	awsVPCTagDir                         = path.Join(awsProviderDir, "vpc-tags")
	awsSecurityGroupDir                  = path.Join(awsProviderDir, "security-groups")
	proxyDir                             = path.Join(awsProviderDir, "proxy")
	kmsDir                               = path.Join(awsProviderDir, "kms")
	sharedVpcPolicyAndHostedZoneDir      = path.Join(awsProviderDir, "shared-vpc-policy-and-hosted-zone")
)

// Dirs of rhcs provider
var (
	clusterDir           = path.Join(rhcsProviderDir, "clusters")
	importResourceDir    = path.Join(rhcsProviderDir, "resource-import")
	idpDir               = path.Join(rhcsProviderDir, "idps")
	machinePoolDir       = path.Join(rhcsProviderDir, "machine-pools")
	dnsDir               = path.Join(rhcsProviderDir, "dns")
	rhcsInfoDir          = path.Join(rhcsProviderDir, "rhcs-info")
	kubeletConfigDir     = path.Join(rhcsProviderDir, "kubelet-config")
	clusterAutoscalerDir = path.Join(rhcsProviderDir, "cluster-autoscaler")
)

func GrantTFvarsFile(manifestDir string) string {
	return path.Join(manifestDir, "terraform.tfvars")
}

func GrantTFstateFile(manifestDir string) string {
	return path.Join(manifestDir, "terraform.tfstate")
}

func GetAccountRoleManifestDir(clusterType CON.ClusterType) string {
	return path.Join(accountRolesDir, clusterType.String())
}

func GetClusterManifestsDir(clusterType CON.ClusterType) string {
	return path.Join(clusterDir, clusterType.String())
}

func GetDnsDomainManifestDir(clusterType CON.ClusterType) string {
	return dnsDir
}

func GetIDPsManifestDir(idpType CON.IDPType, clusterType CON.ClusterType) string {
	return idpDir
}

func GetImportManifestDir(clusterType CON.ClusterType) string {
	return importResourceDir
}

func GetKmsManifestDir(clusterType CON.ClusterType) string {
	return kmsDir
}

func GetKubeletConfigManifestDir(clusterType CON.ClusterType) string {
	return kubeletConfigDir
}

func GetMachinePoolsManifestDir(clusterType CON.ClusterType) string {
	return machinePoolDir
}

func GetOIDCProviderOperatorRolesManifestDir(clusterType CON.ClusterType) string {
	return path.Join(oidcProviderOperatorRolesManifestDir, clusterType.String())
}

func GetProxyManifestDir(clusterType CON.ClusterType) string {
	return proxyDir
}

func GetRhcsInfoManifestDir(clusterType CON.ClusterType) string {
	return rhcsInfoDir
}

func GetSecurityGroupManifestDir(clusterType CON.ClusterType) string {
	return awsSecurityGroupDir
}

func GetVpcTagManifestDir(clusterType CON.ClusterType) string {
	return awsVPCTagDir
}

func GetVpcManifestDir(clusterType CON.ClusterType) string {
	return awsVPCDir
}

func GetSharedVpcPolicyAndHostedZoneDir(clusterType CON.ClusterType) string {
	return sharedVpcPolicyAndHostedZoneDir
}

func GetClusterAutoscalerDir(clusterType CON.ClusterType) string {
	return clusterAutoscalerDir
}
