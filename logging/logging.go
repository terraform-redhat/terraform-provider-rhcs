// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	ocmlogging "github.com/openshift-online/ocm-sdk-go/logging"
)

// Terraform log levels recognized via the TF_LOG environment variable.
const (
	levelTrace = "TRACE"
	levelDebug = "DEBUG"
	levelInfo  = "INFO"
	levelWarn  = "WARN"
	levelError = "ERROR"
)

// validLogLevels lists the levels recognized by Terraform's TF_LOG
// environment variable.
var validLogLevels = []string{levelTrace, levelDebug, levelInfo, levelWarn, levelError}

// TfLogger is a logger that uses the Go `log` package.
type TfLogger struct {
	debugEnabled bool
	infoEnabled  bool
	warnEnabled  bool
	errorEnabled bool
}

// New creates the provider.
func New() ocmlogging.Logger {
	tfLogger := &TfLogger{}
	logLevel := tfLogLevel()
	switch logLevel {
	case levelTrace, levelDebug:
		tfLogger.debugEnabled = true
		tfLogger.infoEnabled = true
		tfLogger.warnEnabled = true
		tfLogger.errorEnabled = true
	case levelInfo:
		tfLogger.infoEnabled = true
		tfLogger.warnEnabled = true
		tfLogger.errorEnabled = true
	case levelWarn:
		tfLogger.warnEnabled = true
		tfLogger.errorEnabled = true
	case levelError, "":
		tfLogger.errorEnabled = true
	}
	return tfLogger
}

// tfLogLevel returns the normalized Terraform log level from the TF_LOG
// environment variable. It mirrors the semantics previously provided by
// terraform-plugin-sdk/v2/helper/logging.LogLevel(): an unset TF_LOG yields an
// empty string, a recognized level (case-insensitive) is upper-cased, and any
// other non-empty value falls back to TRACE.
func tfLogLevel() string {
	envLevel := os.Getenv("TF_LOG")
	if envLevel == "" {
		return ""
	}
	upperLevel := strings.ToUpper(envLevel)
	for _, level := range validLogLevels {
		if upperLevel == level {
			return upperLevel
		}
	}
	log.Printf("[WARN] Invalid log level: %q. Defaulting to level: TRACE. Valid levels are: %+v",
		envLevel, validLogLevels)
	return levelTrace
}

// DebugEnabled returns true iff the debug level is enabled.
func (l *TfLogger) DebugEnabled() bool {
	return l.debugEnabled
}

// InfoEnabled returns true iff the information level is enabled.
func (l *TfLogger) InfoEnabled() bool {
	return l.infoEnabled
}

// WarnEnabled returns true iff the warning level is enabled.
func (l *TfLogger) WarnEnabled() bool {
	return l.warnEnabled
}

// ErrorEnabled returns true iff the error level is enabled.
func (l *TfLogger) ErrorEnabled() bool {
	return l.errorEnabled
}

// Debug sends to the log a debug message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger) Debug(ctx context.Context, format string, args ...any) {
	if l.debugEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Debug(ctx, msg)
	}
}

// Info sends to the log an information message formatted using the fmt.Sprintf function and the
// given format and arguments.
func (l *TfLogger) Info(ctx context.Context, format string, args ...any) {
	if l.infoEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Info(ctx, msg)
	}
}

// Warn sends to the log a warning message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger) Warn(ctx context.Context, format string, args ...any) {
	if l.warnEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Warn(ctx, msg)
	}
}

// Error sends to the log an error message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger) Error(ctx context.Context, format string, args ...any) {
	if l.errorEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Error(ctx, msg)
	}
}

// Fatal sends to the log an error message formatted using the fmt.Sprintf function and the given
// format and arguments. After that it will os.Exit(1)
// This level is always enabled
func (l *TfLogger) Fatal(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	tflog.Error(ctx, msg)
	os.Exit(1)
}
