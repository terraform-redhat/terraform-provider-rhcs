package connection

***REMOVED***
***REMOVED***
	"os"
***REMOVED***

const (
	tokenURL       = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	clientID       = "cloud-services"
	clientSecret   = ""
	skipAuth       = true
	integration    = false
	healthcheckURL = "http://localhost:8083"
***REMOVED***

var cfg = os.Getenv("OCM_ENV"***REMOVED***

func gatewayURL(***REMOVED*** (url string***REMOVED*** {
	switch cfg {
	case "production", "prod":
		url = "api.openshift.com"
	case "staging", "stage":
		url = "api.stage.openshift.com"
	case "integration", "int":
		url = "api.integration.openshift.com"
	default:
		url = "api.stage.openshift.com"
	}

	url = fmt.Sprintf("https://%s", url***REMOVED***
	return url
}
func GateWayURL(***REMOVED*** string {
	return gatewayURL(***REMOVED***
}
