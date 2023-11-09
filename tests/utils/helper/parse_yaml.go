package helper

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type profiles struct {
	Profiles []*profile `yaml:"profiles,omitempty"`
}
type profile struct {
	Name    string                 `yaml:"as,omitempty"`
	Cluster map[string]interface{} `yaml:"cluster,omitempty"`
}

func ParseProfiles(fileName string) map[string]*profile {
	yfile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	p := new(profiles)
	err = yaml.Unmarshal(yfile, &p)
	if err != nil {
		log.Fatal(err)
	}

	profileMap := make(map[string]*profile)
	for _, theProfile := range p.Profiles {
		profileMap[theProfile.Name] = theProfile
	}

	return profileMap
}

func GetProfile(profileName string, fileName string) *profile {
	profileMap := ParseProfiles(fileName)
	if _, exist := profileMap[profileName]; !exist {
		log.Fatalf("Can not find the profile %s in %s\n", profileName, fileName)
	}

	return profileMap[profileName]
}
