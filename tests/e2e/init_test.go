***REMOVED***

***REMOVED***
	"context"
	"testing"

***REMOVED***
***REMOVED***
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***

var ctx context.Context
var token string
var clusterID string

func TestRHCSProvider(t *testing.T***REMOVED*** {
	token = CI.GetEnvWithDefault(CON.TokenENVName, ""***REMOVED***
	clusterID = CI.GetEnvWithDefault(CON.ClusterIDEnv, ""***REMOVED***
	ctx = context.Background(***REMOVED***
	RegisterFailHandler(Fail***REMOVED***
	RunSpecs(t, "RHCS Provider Test"***REMOVED***

}
