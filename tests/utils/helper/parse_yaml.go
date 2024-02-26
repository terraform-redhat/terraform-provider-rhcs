package helper

import (
	"io/ioutil"
	"log"
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
	Name    string                 `yaml:"as,omitempty"`
	Cluster map[string]interface{} `yaml:"cluster,omitempty"`
}

func ParseProfiles(profilesDir string) map[string]*profile {
	files, err := os.ReadDir(profilesDir)
	if err != nil {
		log.Fatal(err)
	}

	profileMap := make(map[string]*profile)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), profilesYamlSuffix) {
			yfile, err := ioutil.ReadFile(path.Join(profilesDir, file.Name()))
			if err != nil {
				log.Fatal(err)
			}

			p := new(profiles)
			err = yaml.Unmarshal(yfile, &p)
			if err != nil {
				log.Fatal(err)
			}

			for _, theProfile := range p.Profiles {
				profileMap[theProfile.Name] = theProfile
			}
		}
	}

	return profileMap
}

func GetProfile(profileName string, profilesDir string) *profile {
	profileMap := ParseProfiles(profilesDir)
	if _, exist := profileMap[profileName]; !exist {
		log.Fatalf("Can not find the profile %s in %s\n", profileName, profilesDir)
	}

	return profileMap[profileName]
}
