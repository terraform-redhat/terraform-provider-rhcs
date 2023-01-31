package helper

***REMOVED***
	"context"
***REMOVED***
***REMOVED***
	"os"
	"time"

***REMOVED***
***REMOVED***
	"github.com/onsi/gomega/format"
	client "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

// ********************************************************************************
// Helper functions that were copied from uhc-clusters-service gitlab repository
const (
	ResponseHeader = "X-Operation-ID"
***REMOVED***

var logger logging.Logger

func CheckError(err error***REMOVED*** {
	if err != nil {
		Fail(fmt.Sprintf("Got an error: %v", err***REMOVED******REMOVED***
	}
}

type Response interface {
	Status(***REMOVED*** int
	Header(***REMOVED*** http.Header
}

func CheckResponse(response Response, err error, expectedStatus int***REMOVED*** {
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	if response.Status(***REMOVED*** != expectedStatus {
		opid := response.Header(***REMOVED***.Get(ResponseHeader***REMOVED***
		Fail(fmt.Sprintf("Expected http response status '%d' but got '%d' with opID '%s': %v", expectedStatus,
			response.Status(***REMOVED***, opid, response***REMOVED******REMOVED***
	}
}

func CheckEmpty(value string, name string***REMOVED*** {
	if value == "" {
		fmt.Printf("ERROR: Empty `%s` param value\n", name***REMOVED***
		os.Exit(2***REMOVED***
	}
}

func GetLogger(***REMOVED*** logging.Logger {
	if logger == nil {
		// Create the logger:
		var err error
		logger, err = logging.NewStdLoggerBuilder(***REMOVED***.
			Build(***REMOVED***
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't create logger: %v", err***REMOVED***
***REMOVED***
		format.TruncateThreshold = uint(4000***REMOVED***
		format.CharactersAroundMismatchToInclude = uint(200***REMOVED***
		return logger
	}
	return logger
}

func WaitForBackendToBeReady(ctx context.Context, connection *client.Connection***REMOVED*** {
	logger.Info(ctx, "Waiting for backend to be ready..."***REMOVED***
	_, err := RunAttempt(func(***REMOVED*** (interface{}, bool***REMOVED*** {
		response, err := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***.List(***REMOVED***.Search("default='t'"***REMOVED***.Send(***REMOVED***
		if err != nil {
			if response.Status(***REMOVED*** == http.StatusInternalServerError {
				Fail("Got internal server error, stopping attempts"***REMOVED***
	***REMOVED*** else {
				logger.Info(ctx, "Got an error querying versions. err = %v", err***REMOVED***
				return nil, true
	***REMOVED***
***REMOVED***
		if response.Error(***REMOVED*** != nil {
			logger.Info(ctx, "Got an error querying versions. err = %v", response.Error(***REMOVED******REMOVED***
			return nil, true
***REMOVED***
		if response.Total(***REMOVED*** == 0 {
			logger.Info(ctx, "Default version not found, waiting another attempt..."***REMOVED***
			return nil, true
***REMOVED***
		return nil, false
	}, 100, time.Second***REMOVED***

	if err != nil {
		Fail("Failed to wait for backend to be ready. Quiting attempts."***REMOVED***
	}
}

// RunAttempt will run a function (attempt***REMOVED***, until the function returns false - meaning no further attempts should
// follow, or until the number of attempts reached maxAttempts. Between each 2 attempts, it will wait for a given
// delay time.
// In case maxAttempts have been reached, an error will be returned, with the latest attempt result.
// The attempt function should return true as long as another attempt should be made, false when no further attempts
// are required - i.e. the attempt succeeded, and the result is available as returned value.
func RunAttempt(attempt func(***REMOVED*** (interface{}, bool***REMOVED***, maxAttempts int, delay time.Duration***REMOVED*** (interface{}, error***REMOVED*** {
	var result interface{}
	for i := 0; i < maxAttempts; i++ {
		fmt.Println("Running attempt", i***REMOVED***
		result, toContinue := attempt(***REMOVED***
		if toContinue {
			fmt.Println("Need to continue for another attempt"***REMOVED***
***REMOVED*** else {
			fmt.Println("Attempt successful, returning result"***REMOVED***
			return result, nil
***REMOVED***
		fmt.Println("Sleeping for", delay***REMOVED***
		time.Sleep(delay***REMOVED***
	}
	return result, fmt.Errorf("got to max attempts %d", maxAttempts***REMOVED***
}

func CreateConnectionWithToken(token string,
	tokenURL string,
	gatewayURL string,
	clientID string,
	clientSecret string***REMOVED*** *client.Connection {
	// Create the connection:
	newConnection, err := client.NewConnectionBuilder(***REMOVED***.
		Logger(logger***REMOVED***.
		Insecure(true***REMOVED***.
		TokenURL(tokenURL***REMOVED***.
		URL(gatewayURL***REMOVED***.
		Client(clientID, clientSecret***REMOVED***.
		Tokens(token***REMOVED***.
		Build(***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	return newConnection
}
