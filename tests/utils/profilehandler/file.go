package profilehandler

import (
	"fmt"
	"slices"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// The cfg will be used to define the testing environment
func loadProfileYamlFile(profileName string) (*Profile, error) {
	p, err := helper.GetProfile(profileName, config.GetClusterProfilesDir())
	if err != nil {
		return nil, err
	}
	Logger.Infof("Loaded cluster profile configuration from profile %s : %v", profileName, p.Cluster)
	profile := &Profile{
		Name: profileName,
	}
	err = helper.MapStructure(p.Cluster, profile)
	if err != nil {
		return nil, err
	}

	setDefaultOrEnvProfileValues(profile)
	return profile, err
}

func setDefaultOrEnvProfileValues(profile *Profile) {
	// Supporting global env setting to overrite profile settings
	setProfileFromEnv := func(name string, envRetriever func() string, profileSetter func(profile *Profile, value string)) {
		value := envRetriever()
		if value != "" {
			profileSetter(profile, value)
			Logger.Infof("Got global env settings for %s, overwritten the profile setting with value %s", name, value)
		}
	}

	setProfileFromEnv("Channel Group", config.GetChannelGroup, func(profile *Profile, value string) {
		profile.ChannelGroup = value
	})
	setProfileFromEnv("Version", config.GetVersion, func(profile *Profile, value string) {
		profile.Version = value
	})
	setProfileFromEnv("Major Version", config.GetMajorVersion, func(profile *Profile, value string) {
		profile.MajorVersion = value
	})
	setProfileFromEnv("Region", config.GetRegion, func(profile *Profile, value string) {
		profile.Region = value
	})
	setProfileFromEnv("Compute Machine Type", config.GetComputeMachineType, func(profile *Profile, value string) {
		profile.ComputeMachineType = value
	})
	setProfileFromEnv("Cluster Name", config.GetRHCSClusterName, func(profile *Profile, value string) {
		profile.ClusterName = value
	})

	if len(profile.AllowedRegistries) > 0 || len(profile.BlockedRegistries) > 0 {
		profile.UseRegistryConfig = true
	}

	if profile.Version == "" {
		profile.Version = constants.VersionLatest
	}
	if profile.ChannelGroup == "" {
		profile.ChannelGroup = constants.VersionStableChannel
	}
}

func LoadProfileYamlFileByENV() (*Profile, error) {
	profileEnv := config.GetClusterProfile()
	if profileEnv == "" {
		panic(fmt.Errorf("ENV Variable CLUSTER_PROFILE is empty, please make sure you set the env value"))
	}
	return loadProfileYamlFile(profileEnv)

}

func getRandomProfile(clusterTypes ...constants.ClusterType) (profile *Profile, err error) {
	if len(clusterTypes) > 0 {
		Logger.Infof("Get random profile for cluster types: %v", clusterTypes)
	} else {
		Logger.Info("Get random profile from all profiles")
	}

	profilesMap, err := helper.ParseProfiles(config.GetClusterProfilesDir())
	if err != nil {
		return
	}
	profilesNames := make([]string, 0, len(profilesMap))
	for k, v := range profilesMap {
		clusterType := constants.FindClusterType(fmt.Sprintf("%v", v.Cluster["cluster_type"]))
		if !v.NeedSpecificConfig {
			if len(clusterTypes) <= 0 || slices.Contains(clusterTypes, clusterType) {
				profilesNames = append(profilesNames, k)
			}
		}
	}
	Logger.Debugf("Got profile names %v", profilesNames)
	profileName := profilesMap[profilesNames[helper.RandomInt(len(profilesNames))]].Name
	profile, err = loadProfileYamlFile(profileName)
	Logger.Debugf("Choose profile: %s", profile.Name)
	return profile, err
}
