package cms

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	"github.com/onsi/gomega/format"
	sdk "github.com/openshift-online/ocm-sdk-go"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var connection *sdk.Connection
var ctx context.Context

func TestResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CMS Helper Suite")
}

var _ = BeforeEach(func() {
	format.MaxLength = 0 // set gomega format MaxLength to 0 to see all the diff when fails
	// Create the server:
	var ca string
	TestServer, ca = MakeTCPTLSServer()
	// Create an access token:
	token := MakeTokenString("Bearer", 10*time.Minute)

	ctx = context.Background()

	var err error
	connection, err = sdk.NewConnectionBuilder().URL(TestServer.URL()).TrustedCAFile(ca).Tokens(token).BuildContext(ctx)
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	// Close the server:
	TestServer.Close()

	// Close the connection:
	connection.Close()
})
