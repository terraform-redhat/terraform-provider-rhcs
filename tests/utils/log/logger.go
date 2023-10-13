package log

***REMOVED***
	logging "github.com/sirupsen/logrus"
***REMOVED***

func GetLogger(***REMOVED*** *logging.Logger {
	// Create the logger
	logger := logging.New(***REMOVED***
	// Set logger level for your debug command
	logger.SetLevel(logging.InfoLevel***REMOVED***
	return logger
}

var Logger *logging.Logger = GetLogger(***REMOVED***
