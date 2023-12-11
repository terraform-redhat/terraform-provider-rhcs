package constants

const (
	DefaultVPCCIDR = "11.0.0.0/16"
)

var (
	NilMap           map[string]string
	Tags             = map[string]string{"tag1": "test_tag1", "tag2": "test_tag2"}
	ClusterAdminUser = "cluster_admin_name"
	DefaultMPLabels  = map[string]string{
		"test1": "testdata1",
	}
	LdapURL       = "ldap://ldap.forumsys.com/dc=example,dc=com?uid"
	GitLabURL     = "https://gitlab.cee.redhat.com"
	Organizations = []string{"openshift"}
	HostedDomain  = "redhat.com"
)
