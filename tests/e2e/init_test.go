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
	RunSpecs(t, "RHCS Provider Test"***REMOVED***

}

var _ = BeforeSuite(func(***REMOVED*** {
	token = os.Getenv(CON.TokenENVName***REMOVED***
	var err error
	clusterID, err = CI.PrepareRHCSClusterByProfileENV(***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	ctx = context.Background(***REMOVED***
}***REMOVED***
