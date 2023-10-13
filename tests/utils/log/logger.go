package log

import (
	logging "github.com/sirupsen/logrus"
)

func GetLogger() *logging.Logger {
	// Create the logger
	logger := logging.New()
	// Set logger level for your debug command
	logger.SetLevel(logging.InfoLevel)
	return logger
}

var Logger *logging.Logger = GetLogger()
