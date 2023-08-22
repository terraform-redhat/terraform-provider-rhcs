package connection

import (
	"os"

	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

var (
	RHCSOCMToken = os.Getenv(con.TokenENVName)
)
