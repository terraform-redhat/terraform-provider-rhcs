// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

const (
	EnvWorkspace  = "WORKSPACE"
	EnvQEUsage    = "QE_USAGE"
	EnvTestOutput = "RHCS_OUTPUT"

	EnvRHCSToken = "RHCS_TOKEN"
	EnvRHCSURL   = "RHCS_URL"

	EnvClusterID = "CLUSTER_ID"

	EnvClusterProfile     = "CLUSTER_PROFILE"
	EnvClusterProfilesDir = "PROFILES_DIR"

	EnvRegion                = "REGION"
	EnvMajorVersion          = "MAJOR_VERSION"
	EnvVersion               = "VERSION"
	EnvChannelGroup          = "CHANNEL_GROUP"
	EnvRHCSVersion           = "RHCS_VERSION"
	EnvRHCSSource            = "RHCS_SOURCE"
	EnvRHCSModuleRef         = "RHCS_MODULE_REF" // Set this to update the git ref for the RHCS modules
	EnvWaitOperators         = "WAIT_OPERATORS"
	EnvRHCSClusterName       = "RHCS_CLUSTER_NAME"
	EnvRHCSClusterNamePrefix = "RHCS_CLUSTER_NAME_PREFIX"
	EnvRHCSClusterNameSuffix = "RHCS_CLUSTER_NAME_SUFFIX"
	EnvComputeMachineType    = "COMPUTE_MACHINE_TYPE"

	EnvManifestsFolder                   = "MANIFESTS_FOLDER"
	EnvSharedVpcAWSSharedCredentialsFile = "SHARED_VPC_AWS_SHARED_CREDENTIALS_FILE"

	EnvNoClusterDestroy = "NO_CLUSTER_DESTROY"

	EnvSubnetIDs         = "SUBNET_IDS"
	EnvAvailabilityZones = "AVAILABILITY_ZONES"
)

func GetRootDir() string {
	currentDir, _ := os.Getwd()
	project := "terraform-provider-rhcs"
	return GetEnvWithDefault(EnvWorkspace, strings.SplitAfter(currentDir, project)[0])
}

func GetClusterProfilesDir() string {
	return GetEnvWithDefault(EnvClusterProfilesDir, path.Join(GetRootDir(), "tests", "ci", "profiles"))
}

func GetRHCSOutputDir() string {
	var rhcsNewOutPath string

	if rhcsNewOutPath := GetEnvWithDefault(EnvTestOutput, ""); rhcsNewOutPath != "" {
		return rhcsNewOutPath
	}
	rhcsNewOutPath = path.Join(GetRootDir(), "tests", "rhcs_output")
	os.MkdirAll(rhcsNewOutPath, 0777)
	return rhcsNewOutPath
}

func GetKubeConfigDir() string {
	outputDIR := GetRHCSOutputDir()
	configDir := path.Join(outputDIR, "kubeconfig")
	if _, err := os.Stat(configDir); err != nil {
		os.MkdirAll(configDir, 0777)
	}
	return configDir
}

func GetQEUsage() string {
	return GetEnvWithDefault(EnvQEUsage, "")
}

func GetClusterProfile() string {
	return GetEnvWithDefault(EnvClusterProfile, "")
}

func GetRHCSURL() string {
	return GetEnvWithDefault(EnvRHCSURL, constants.DefaultRHCSURL)
}

func GetRHCSOCMToken() string {
	return GetEnvWithDefault(EnvRHCSToken, "")
}

func GetRHCSClusterName() string {
	return GetEnvWithDefault(EnvRHCSClusterName, "")
}

func GetRHCSClusterNamePrefix() string {
	return GetEnvWithDefault(EnvRHCSClusterNamePrefix, "")
}

func GetRHCSClusterNameSuffix() string {
	return GetEnvWithDefault(EnvRHCSClusterNameSuffix, "")
}

func GetRegion() string {
	return GetEnvWithDefault(EnvRegion, "")
}

func GetChannelGroup() string {
	return GetEnvWithDefault(EnvChannelGroup, "")
}

func GetVersion() string {
	return GetEnvWithDefault(EnvVersion, "")
}

func GetMajorVersion() string {
	return GetEnvWithDefault(EnvMajorVersion, "")
}

func GetComputeMachineType() string {
	return GetEnvWithDefault(EnvComputeMachineType, "")
}

func GetSubnetIDs() string {
	return GetEnvWithDefault(EnvSubnetIDs, "")
}

func GetAvailabilityZones() string {
	return GetEnvWithDefault(EnvAvailabilityZones, "")
}

func GetRHCSSource() string {
	return GetEnvWithDefault(EnvRHCSSource, "")
}

func GetRHCSVersion() string {
	return GetEnvWithDefault(EnvRHCSVersion, "")
}

func GetRHCSModuleRef() string {
	return GetEnvWithDefault(EnvRHCSModuleRef, "")
}

func IsWaitForOperators() bool {
	return GetEnvWithDefault(EnvWaitOperators, "false") == "true"
}

func IsNoClusterDestroy() bool {
	return GetEnvWithDefault(EnvNoClusterDestroy, "false") == "true"
}

func GetManifestsDir() string {
	manifestsDir := GetEnvWithDefault(EnvManifestsFolder, "")
	if manifestsDir != "" {
		return manifestsDir
	}
	currentDir, _ := os.Getwd()
	manifestsDir = path.Join(strings.SplitAfter(currentDir, "tests")[0], "tf-manifests")
	if _, err := os.Stat(manifestsDir); err != nil {
		panic(fmt.Sprintf("Manifests dir %s doesn't exist. Make sure you have the manifests dir in testing repo or set the correct env MANIFESTS_DIR value", manifestsDir))
	}
	return manifestsDir
}

func GetSharedVpcAWSSharedCredentialsFile() string {
	return GetEnvWithDefault(EnvSharedVpcAWSSharedCredentialsFile, "")
}

func GetEnvWithDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetEnvOrFail(key string) string {
	if value := GetEnvWithDefault(key, ""); value != "" {
		return value
	} else {
		panic(fmt.Errorf("ENV Variable %s is empty, please make sure you set the env value", key))
	}
}

func GetClusterNameFilename() string {
	return path.Join(GetRHCSOutputDir(), "cluster-name")
}

func GetClusterTrustBundleFilename() string {
	return path.Join(GetRHCSOutputDir(), "ca.cert")
}

func GetClusterAdminUserFilename() string {
	return path.Join(GetRHCSOutputDir(), "cluster-admin-user")
}
