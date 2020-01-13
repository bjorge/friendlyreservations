package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/bjorge/friendlyreservations/platform"
)

// only a single implementation for now
type commonLogImpl struct{}

// New is the factory method to create a logger
func New() platform.Logger {
	impl := &commonLogImpl{}
	if strings.HasSuffix(os.Args[0], ".test") {
		impl.LogDebugf("Logger running in test environment")
	} else {
		impl.LogDebugf("Logger not running in test environment")
	}
	return impl
}

// LogDebugf is called to log a debug level message to the log
func (r *commonLogImpl) LogDebugf(format string, args ...interface{}) {
	logLineFormat := fmt.Sprintf("DEBUG: %s\n", format)
	fmt.Printf(logLineFormat, args...)
}

// LogErrorf is called to log an error level message to the log
func (r *commonLogImpl) LogErrorf(format string, args ...interface{}) {
	logLineFormat := fmt.Sprintf("ERROR: %s\n", format)
	fmt.Printf(logLineFormat, args...)
}

// LogInfof is called to log an info level message to the log
func (r *commonLogImpl) LogInfof(format string, args ...interface{}) {
	logLineFormat := fmt.Sprintf("INFO: %s\n", format)
	fmt.Printf(logLineFormat, args...)
}

// LogWarningf is called to log a warning level message to the log
func (r *commonLogImpl) LogWarningf(format string, args ...interface{}) {
	logLineFormat := fmt.Sprintf("WARN: %s\n", format)
	fmt.Printf(logLineFormat, args...)
}
