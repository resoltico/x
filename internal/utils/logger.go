// Author: Ervins Strauhmanis
// License: MIT

package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// InitLogger initializes and configures the logger
func InitLogger(debugMode bool) *logrus.Logger {
	logger := logrus.New()
	
	// Set output to stdout
	logger.SetOutput(os.Stdout)
	
	// Use text formatter instead of JSON for better readability in debug mode
	if debugMode {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Debug mode enabled")
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		logger.SetLevel(logrus.InfoLevel)
	}
	
	return logger
}

// GetLogger returns the default logger instance
func GetLogger() *logrus.Logger {
	return logrus.StandardLogger()
}