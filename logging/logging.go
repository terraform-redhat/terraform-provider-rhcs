package logging

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	tflogging "github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	ocmlogging "github.com/openshift-online/ocm-sdk-go/logging"
)

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
	logLevel := tflogging.LogLevel()
	switch logLevel {
	case "TRACE", "DEBUG":
		tfLogger.debugEnabled = true
		tfLogger.infoEnabled = true
		tfLogger.warnEnabled = true
		tfLogger.errorEnabled = true
	case "INFO":
		tfLogger.infoEnabled = true
		tfLogger.warnEnabled = true
		tfLogger.errorEnabled = true
	case "WARN":
		tfLogger.warnEnabled = true
		tfLogger.errorEnabled = true
	case "ERROR", "":
		tfLogger.errorEnabled = true
	}
	return tfLogger
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
func (l *TfLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.debugEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Debug(ctx, msg)
	}
}

// Info sends to the log an information message formatted using the fmt.Sprintf function and the
// given format and arguments.
func (l *TfLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.infoEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Info(ctx, msg)
	}
}

// Warn sends to the log a warning message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.warnEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Warn(ctx, msg)
	}
}

// Error sends to the log an error message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.errorEnabled {
		msg := fmt.Sprintf(format, args...)
		tflog.Error(ctx, msg)
	}
}

// Fatal sends to the log an error message formatted using the fmt.Sprintf function and the given
// format and arguments. After that it will os.Exit(1)
// This level is always enabled
func (l *TfLogger) Fatal(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	tflog.Error(ctx, msg)
	os.Exit(1)
}
