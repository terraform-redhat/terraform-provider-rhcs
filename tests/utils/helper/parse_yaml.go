package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	profilesYamlSuffix = "_profiles.yml"
)

type profiles struct {
	Profiles []*profile `yaml:"profiles,omitempty"`
}
type profile struct {
	Name               string                 `yaml:"as,omitempty"`
	NeedSpecificConfig bool                   `yaml:"need_specific_config,omitempty"` // Some profiles need external configuration files
	Cluster            map[string]interface{} `yaml:"cluster,omitempty"`
}

func ParseProfiles(profilesDir string) (map[string]*profile, error) {
	files, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, err
	}

	profileMap := make(map[string]*profile)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), profilesYamlSuffix) {
			yfile, err := ioutil.ReadFile(path.Join(profilesDir, file.Name()))
			if err != nil {
				return nil, err
			}

			p := new(profiles)
			err = yaml.Unmarshal(yfile, &p)
			if err != nil {
				return nil, err
			}

			for _, theProfile := range p.Profiles {
				profileMap[theProfile.Name] = theProfile
			}
		}
	}

	return profileMap, nil
}

func GetProfile(profileName string, profilesDir string) (*profile, error) {
	profileMap, err := ParseProfiles(profilesDir)
	if err != nil {
		return nil, err
	}
	if _, exist := profileMap[profileName]; !exist {
		return nil, fmt.Errorf("Can not find the profile %s in %s", profileName, profilesDir)
	}

	return profileMap[profileName], nil
}
