package connection

import (
	"fmt"
	"os"
)

const (
	tokenURL       = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	clientID       = "cloud-services"
	clientSecret   = ""
	skipAuth       = true
	integration    = false
	healthcheckURL = "http://localhost:8083"
)

var cfg = os.Getenv("OCM_ENV")

func gatewayURL() (url string) {
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

	url = fmt.Sprintf("https://%s", url)
	return url
}
func GateWayURL() string {
	return gatewayURL()
}
