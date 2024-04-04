package constants

import (
	"fmt"
	"os"
)

const (
	DefaultVPCCIDR = "10.0.0.0/16"
)

var (
	NilMap           map[string]string
	Tags             = map[string]string{"tag1": "test_tag1", "tag2": "test_tag2"}
	ClusterAdminUser = "rhcs-clusteradmin"
	DefaultMPLabels  = map[string]string{
		"test1": "testdata1",
	}
	CustomProperties = map[string]string{"custom_property": "test", "qe_usage": GetEnvWithDefault(QEUsage, "")}
	LdapURL          = "ldap://ldap.forumsys.com/dc=example,dc=com?uid"
	GitLabURL        = "https://gitlab.cee.redhat.com"
	Organizations    = []string{"openshift"}
	HostedDomain     = "redhat.com"
)

func GetEnvWithDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		if key == TokenENVName {
			panic(fmt.Errorf("ENV Variable RHCS_TOKEN is empty, please make sure you set the env value"))
		}
	}
	return defaultValue
}
