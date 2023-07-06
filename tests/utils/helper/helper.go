package helper

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	client "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift-online/ocm-sdk-go/logging"
)

// ********************************************************************************
// Helper functions that were copied from uhc-clusters-service gitlab repository
const (
	ResponseHeader = "X-Operation-ID"
)

var logger logging.Logger

func CheckError(err error) {
	if err != nil {
		Fail(fmt.Sprintf("Got an error: %v", err))
	}
}

type Response interface {
	Status() int
	Header() http.Header
}

func CheckResponse(response Response, err error, expectedStatus int) {
	Expect(err).ToNot(HaveOccurred())
	if response.Status() != expectedStatus {
		opid := response.Header().Get(ResponseHeader)
		Fail(fmt.Sprintf("Expected http response status '%d' but got '%d' with opID '%s': %v", expectedStatus,
			response.Status(), opid, response))
	}
}

func CheckEmpty(value string, name string) {
	if value == "" {
		fmt.Printf("ERROR: Empty `%s` param value\n", name)
		os.Exit(2)
	}
}

func GetLogger() logging.Logger {
	if logger == nil {
		// Create the logger:
		var err error
		logger, err = logging.NewStdLoggerBuilder().
			Build()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't create logger: %v", err)
		}
		format.TruncateThreshold = uint(4000)
		format.CharactersAroundMismatchToInclude = uint(200)
		return logger
	}
	return logger
}

func WaitForBackendToBeReady(ctx context.Context, connection *client.Connection) {
	logger.Info(ctx, "Waiting for backend to be ready...")
	_, err := RunAttempt(func() (interface{}, bool) {
		response, err := connection.ClustersMgmt().V1().Versions().List().Search("default='t'").Send()
		if err != nil {
			if response.Status() == http.StatusInternalServerError {
				Fail("Got internal server error, stopping attempts")
			} else {
				logger.Info(ctx, "Got an error querying versions. err = %v", err)
				return nil, true
			}
		}
		if response.Error() != nil {
			logger.Info(ctx, "Got an error querying versions. err = %v", response.Error())
			return nil, true
		}
		if response.Total() == 0 {
			logger.Info(ctx, "Default version not found, waiting another attempt...")
			return nil, true
		}
		return nil, false
	}, 100, time.Second)

	if err != nil {
		Fail("Failed to wait for backend to be ready. Quiting attempts.")
	}
}

// RunAttempt will run a function (attempt), until the function returns false - meaning no further attempts should
// follow, or until the number of attempts reached maxAttempts. Between each 2 attempts, it will wait for a given
// delay time.
// In case maxAttempts have been reached, an error will be returned, with the latest attempt result.
// The attempt function should return true as long as another attempt should be made, false when no further attempts
// are required - i.e. the attempt succeeded, and the result is available as returned value.
func RunAttempt(attempt func() (interface{}, bool), maxAttempts int, delay time.Duration) (interface{}, error) {
	var result interface{}
	for i := 0; i < maxAttempts; i++ {
		fmt.Println("Running attempt", i)
		result, toContinue := attempt()
		if toContinue {
			fmt.Println("Need to continue for another attempt")
		} else {
			fmt.Println("Attempt successful, returning result")
			return result, nil
		}
		fmt.Println("Sleeping for", delay)
		time.Sleep(delay)
	}
	return result, fmt.Errorf("got to max attempts %d", maxAttempts)
}

func CreateConnectionWithToken(token string,
	tokenURL string,
	gatewayURL string,
	clientID string,
	clientSecret string) *client.Connection {
	// Create the connection:
	newConnection, err := client.NewConnectionBuilder().
		Logger(logger).
		Insecure(true).
		TokenURL(tokenURL).
		URL(gatewayURL).
		Client(clientID, clientSecret).
		Tokens(token).
		Build()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	return newConnection
}
