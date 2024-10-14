package profilehandler

import "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"

const (
	DefaultVPCCIDR = "10.0.0.0/16"
)

var (
	Tags             = map[string]string{"tag1": "test_tag1", "tag2": "test_tag2"}
	ClusterAdminUser = "rhcs-clusteradmin"
	DefaultMPLabels  = map[string]string{
		"test1": "testdata1",
	}
	CustomProperties = map[string]string{"custom_property": "test", "qe_usage": config.GetQEUsage()}
	LdapURL          = "ldap://ldap.forumsys.com/dc=example,dc=com?uid"
	GitLabURL        = "https://gitlab.cee.redhat.com"
	Organizations    = []string{"openshift"}
	HostedDomain     = "redhat.com"
)
