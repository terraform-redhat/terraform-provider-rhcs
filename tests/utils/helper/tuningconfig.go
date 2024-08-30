package helper

import "fmt"

type TuningConfigSpecRoot struct {
	Profile   []TuningConfigSpecProfile   `json:"profile,omitempty" yaml:"profile,omitempty"`
	Recommend []TuningConfigSpecRecommend `json:"recommend,omitempty" yaml:"recommend,omitempty"`
}

type TuningConfigSpecProfile struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Data string `yaml:"data,omitempty" json:"data,omitempty"`
}

type TuningConfigSpecRecommend struct {
	Priority int    `yaml:"priority,omitempty" json:"priority,omitempty"`
	Profile  string `yaml:"profile,omitempty" json:"profile,omitempty"`
}

func NewTuningConfigSpecRootStub(tcName string, vmDirtyRatio int, priority int) TuningConfigSpecRoot {
	return TuningConfigSpecRoot{
		Profile: []TuningConfigSpecProfile{
			{
				Data: NewTuningConfigSpecProfileData(vmDirtyRatio),
				Name: tcName + "-profile",
			},
		},
		Recommend: []TuningConfigSpecRecommend{
			{
				Priority: priority,
				Profile:  tcName + "-profile",
			},
		},
	}
}

func NewTuningConfigSpecProfileData(vmDirtyRatio int) string {
	return fmt.Sprintf("[main]\nsummary=Custom OpenShift profile\ninclude=openshift-node\n\n"+
		"[sysctl]\nvm.dirty_ratio=\"%d\"\n",
		vmDirtyRatio)
}
