package profilehandler

import "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"

const (
	DefaultVPCCIDR = "10.0.0.0/16"
)

// FakeClusterProxy* are injected when a fake cluster profile has proxy:true.
// The test [id:67607] only reads OCM-stored values — no real proxy server is needed.
const (
	FakeClusterHTTPProxy  = "http://proxy.example.com:3128"
	FakeClusterHTTPSProxy = "https://proxy.example.com:3128"
	FakeClusterNoProxy    = "quay.io"
	// Real X.509 CA certificate (kube-apiserver-localhost-signer, sourced from
	// uhc-clusters-service/test/helper/proxy_helpers.go). OCM uses strict X.509
	// validation via x509.NewCertPool().AppendCertsFromPEM(); this cert passes.
	// OCM stores it and returns "REDACTED" on read, which test [id:67607] asserts.
	FakeClusterTrustBundle = `-----BEGIN CERTIFICATE-----
MIIDQDCCAiigAwIBAgIIBGGRHcOoPJswDQYJKoZIhvcNAQELBQAwPjESMBAGA1UE
CxMJb3BlbnNoaWZ0MSgwJgYDVQQDEx9rdWJlLWFwaXNlcnZlci1sb2NhbGhvc3Qt
c2lnbmVyMB4XDTIwMTAwMjA0NTUyMVoXDTMwMDkzMDA0NTUyMVowPjESMBAGA1UE
CxMJb3BlbnNoaWZ0MSgwJgYDVQQDEx9rdWJlLWFwaXNlcnZlci1sb2NhbGhvc3Qt
c2lnbmVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7pu15j70WVck
l8htrXqucI/hhtRYZAq/J/OUxs6vew6bGLuphheyx7enQCki9HxpoaHIt4AScoeB
jQLtg3/puqFgwtdDSHM3BfpljwSAPdzApKHBPSiL1GmN+JLrPPJNn1oBWN83/36q
oCATV50iqhNshpTqnFZwxfOV/T3/xasoHeAUyt9wBomR9ypsMIRx7YzBz5IgDxXD
WNYcLM3B4jXlTBP78AKGdEltwPphYCr+LVV2Bl65GVX5TuHP9A+fer8NMsWw+xys
IUwjLlf9CndVa7EbG93dv7/8tE8va9NCHSII4GC3H0FArqDQHuHuH2nrwz6CiqS7
2WjkS5/ojwIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUwAwEB
/zAdBgNVHQ4EFgQUsmb3MqDL3e0Hq3VZxJoNBoo8YT8wDQYJKoZIhvcNAQELBQAD
ggEBAL0N8x5fd2m1YgnJdIR9Qjr6KJKIvaQDFuCAXFWwLLbTl9h7YKq8I6agyBnG
pT9SeLN2qiSCeDp9XkMFRKLAR4NrgUDi2Kr1UmrxMFcaisJ36xDkX04xTzSwppQx
X01ezgrRINIu0hk2DMMmSpMNxbVV5YPQ4W0YERw0jww4cQm5R6nmiTFpXUTtsHIA
wGw2dzA56I5C1gT0FA0hO3fKt4dGGvWYy3Bn7OYn8W3G8G/qU9WEgjoDhEiiyJeb
Xeqy61VPAZMhxBVBayzZnfk69sT2iHiiEAab/nMYhsgL2VsFQP/PdpRoBA+fc6M/
VFsNV4Vr5fAxG6/h0WAB94JGf+A=
-----END CERTIFICATE-----
`
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
