package profilehandler

import (
	"fmt"
	"os"
	"slices"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// The cfg will be used to define the testing environment
var cfg = constants.RHCS

func loadProfileYamlFile(profileName string) (*Profile, error) {
	p, err := helper.GetProfile(profileName, GetYAMLProfilesDir())
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
	if os.Getenv("CHANNEL_GROUP") != "" {
		Logger.Infof("Got global env settings for CHANNEL_GROUP, overwritten the profile setting with value %s", os.Getenv("CHANNEL_GROUP"))
		profile.ChannelGroup = os.Getenv("CHANNEL_GROUP")
	}
	if os.Getenv("VERSION") != "" {
		Logger.Infof("Got global env settings for VERSION, overwritten the profile setting with value %s", os.Getenv("VERSION"))
		profile.Version = os.Getenv("VERSION")
	}
	if os.Getenv("REGION") != "" {
		Logger.Infof("Got global env settings for REGION, overwritten the profile setting with value %s", os.Getenv("REGION"))
		profile.Region = os.Getenv("REGION")
	}

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

func LoadProfileYamlFileByENV() (profile *Profile, err error) {
	profileEnv := os.Getenv(constants.RhcsClusterProfileENV)
	if profileEnv == "" {
		panic(fmt.Errorf("ENV Variable CLUSTER_PROFILE is empty, please make sure you set the env value"))
	}
	return loadProfileYamlFile(profileEnv)
}

func GetYAMLProfilesDir() string {
	return cfg.YAMLProfilesDir
}

func getRandomProfile(clusterTypes ...constants.ClusterType) (profile *Profile, err error) {
	if len(clusterTypes) > 0 {
		Logger.Infof("Get random profile for cluster types: %v", clusterTypes)
	} else {
		Logger.Info("Get random profile from all profiles")
	}

	profilesMap, err := helper.ParseProfiles(GetYAMLProfilesDir())
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
