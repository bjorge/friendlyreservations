package utilities

import (
	"context"
	"fmt"

	"google.golang.org/appengine/log"
)

var logToStdOut = false

// BUG(bjorge): change name to Debugf

// DebugLog is called to log a debug level message to the log
func DebugLog(ctx context.Context, format string, args ...interface{}) {
	if logToStdOut {
		logLineFormat := fmt.Sprintf("DEBUG: %s\n", format)
		fmt.Printf(logLineFormat, args...)
	} else {
		log.Debugf(ctx, format, args...)
	}
}

// LogErrorf is called to log an error level message to the log
func LogErrorf(ctx context.Context, format string, args ...interface{}) {
	if logToStdOut {
		logLineFormat := fmt.Sprintf("ERROR: %s\n", format)
		fmt.Printf(logLineFormat, args...)
	} else {
		log.Errorf(ctx, format, args...)
	}
}

// LogInfof is called to log an info level message to the log
func LogInfof(ctx context.Context, format string, args ...interface{}) {
	if logToStdOut {
		logLineFormat := fmt.Sprintf("INFO: %s\n", format)
		fmt.Printf(logLineFormat, args...)
	} else {
		log.Infof(ctx, format, args...)
	}
}

// LogWarningf is called to log a warning level message to the log
func LogWarningf(ctx context.Context, format string, args ...interface{}) {
	if logToStdOut {
		logLineFormat := fmt.Sprintf("WARN: %s\n", format)
		fmt.Printf(logLineFormat, args...)
	} else {
		log.Warningf(ctx, format, args...)
	}
}
