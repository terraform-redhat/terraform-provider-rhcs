***REMOVED***

***REMOVED***
	"context"
	"os"
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
	RegisterFailHandler(Fail***REMOVED***
	token = os.Getenv(CON.TokenENVName***REMOVED***
	clusterID = CI.PrepareRHCSClusterByProfileENV(***REMOVED***
	ctx = context.Background(***REMOVED***
	RunSpecs(t, "RHCS Provider Test"***REMOVED***

}
