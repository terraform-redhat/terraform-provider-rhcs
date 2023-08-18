package helper

***REMOVED***
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
***REMOVED***

type profiles struct {
	Profiles []*profile `yaml:"profiles,omitempty"`
}
type profile struct {
	Name    string                 `yaml:"as,omitempty"`
	Cluster map[string]interface{} `yaml:"cluster,omitempty"`
}

func ParseProfiles(fileName string***REMOVED*** map[string]*profile {
	yfile, err := ioutil.ReadFile(fileName***REMOVED***
	if err != nil {
		log.Fatal(err***REMOVED***
	}

	p := new(profiles***REMOVED***
	err = yaml.Unmarshal(yfile, &p***REMOVED***
	if err != nil {
		log.Fatal(err***REMOVED***
	}

	profileMap := make(map[string]*profile***REMOVED***
	for _, theProfile := range p.Profiles {
		profileMap[theProfile.Name] = theProfile
	}

	return profileMap
}

func GetProfile(profileName string, fileName string***REMOVED*** *profile {
	profileMap := ParseProfiles(fileName***REMOVED***
	if _, exist := profileMap[profileName]; !exist {
		log.Fatalf("Can not find the profile %s in %s\n", profileName, fileName***REMOVED***
	}

	return profileMap[profileName]
}
