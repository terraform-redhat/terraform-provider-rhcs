package logging

***REMOVED***
	"context"
***REMOVED***
	"os"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	tflogging "github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	ocmlogging "github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

// TfLogger is a logger that uses the Go `log` package.
type TfLogger struct {
	debugEnabled bool
	infoEnabled  bool
	warnEnabled  bool
	errorEnabled bool
}

// New creates the provider.
func New(***REMOVED*** ocmlogging.Logger {
	tfLogger := &TfLogger{}
	logLevel := tflogging.LogLevel(***REMOVED***
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
func (l *TfLogger***REMOVED*** DebugEnabled(***REMOVED*** bool {
	return l.debugEnabled
}

// InfoEnabled returns true iff the information level is enabled.
func (l *TfLogger***REMOVED*** InfoEnabled(***REMOVED*** bool {
	return l.infoEnabled
}

// WarnEnabled returns true iff the warning level is enabled.
func (l *TfLogger***REMOVED*** WarnEnabled(***REMOVED*** bool {
	return l.warnEnabled
}

// ErrorEnabled returns true iff the error level is enabled.
func (l *TfLogger***REMOVED*** ErrorEnabled(***REMOVED*** bool {
	return l.errorEnabled
}

// Debug sends to the log a debug message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger***REMOVED*** Debug(ctx context.Context, format string, args ...interface{}***REMOVED*** {
	if l.debugEnabled {
		msg := fmt.Sprintf(format, args...***REMOVED***
		tflog.Debug(ctx, msg***REMOVED***
	}
}

// Info sends to the log an information message formatted using the fmt.Sprintf function and the
// given format and arguments.
func (l *TfLogger***REMOVED*** Info(ctx context.Context, format string, args ...interface{}***REMOVED*** {
	if l.infoEnabled {
		msg := fmt.Sprintf(format, args...***REMOVED***
		tflog.Info(ctx, msg***REMOVED***
	}
}

// Warn sends to the log a warning message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger***REMOVED*** Warn(ctx context.Context, format string, args ...interface{}***REMOVED*** {
	if l.warnEnabled {
		msg := fmt.Sprintf(format, args...***REMOVED***
		tflog.Warn(ctx, msg***REMOVED***
	}
}

// Error sends to the log an error message formatted using the fmt.Sprintf function and the given
// format and arguments.
func (l *TfLogger***REMOVED*** Error(ctx context.Context, format string, args ...interface{}***REMOVED*** {
	if l.errorEnabled {
		msg := fmt.Sprintf(format, args...***REMOVED***
		tflog.Error(ctx, msg***REMOVED***
	}
}

// Fatal sends to the log an error message formatted using the fmt.Sprintf function and the given
// format and arguments. After that it will os.Exit(1***REMOVED***
// This level is always enabled
func (l *TfLogger***REMOVED*** Fatal(ctx context.Context, format string, args ...interface{}***REMOVED*** {
	msg := fmt.Sprintf(format, args...***REMOVED***
	tflog.Error(ctx, msg***REMOVED***
	os.Exit(1***REMOVED***
}
