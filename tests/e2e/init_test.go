***REMOVED***

***REMOVED***
	"context"
	"os"
	"testing"

***REMOVED***
***REMOVED***
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***

var ctx context.Context
var token string

func TestRHCSProvider(t *testing.T***REMOVED*** {
	token = os.Getenv(CON.TokenENVName***REMOVED***
	ctx = context.Background(***REMOVED***
	RegisterFailHandler(Fail***REMOVED***
	RunSpecs(t, "RHCS Provider Test"***REMOVED***

}
