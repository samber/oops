package oopslogrus

import (
	"errors"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewOopsFormatter creates a new logrus formatter that automatically
// detects and formats oops.OopsError instances in log entries.
//
// This formatter wraps an existing logrus formatter and enhances it with
// oops error handling capabilities. When an oops.OopsError is detected
// in a log entry's "error" field, the formatter automatically extracts
// all error context, metadata, and stack traces and adds them to the
// log entry data.
//
// If no secondary formatter is provided, the standard logrus formatter
// is used as the underlying formatter.
//
// Example usage:
//
//	logger := logrus.New()
//	logger.SetFormatter(oopslogrus.NewOopsFormatter(nil))
//	logger.WithError(err).Error("operation failed")
func NewOopsFormatter(secondaryFormatter logrus.Formatter) *oopsFormatter {
	if secondaryFormatter == nil {
		secondaryFormatter = logrus.StandardLogger().Formatter
	}

	return &oopsFormatter{
		formatter: secondaryFormatter,
	}
}

// oopsFormatter implements logrus.Formatter to provide enhanced
// formatting for oops.OopsError instances while delegating to an
// underlying formatter for the actual log output formatting.
//
// This formatter acts as a middleware that processes log entries
// before they are formatted by the underlying formatter, ensuring
// that oops error information is properly extracted and included
// in the final log output.
type oopsFormatter struct {
	formatter logrus.Formatter
}

// Format implements logrus.Formatter interface to process log entries
// and enhance them with oops error information when applicable.
//
// This method examines the log entry for oops.OopsError instances in
// the "error" field and automatically extracts all error context,
// metadata, and stack traces. The enhanced entry is then passed to
// the underlying formatter for final formatting.
//
// Performance: This method has minimal overhead when no oops errors
// are present. When oops errors are detected, the overhead is proportional
// to the amount of error context and metadata being extracted.
//
// Thread Safety: This method is thread-safe and can be called concurrently.
func (f *oopsFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Check if the log entry contains an error field
	errField, ok := entry.Data["error"]
	if ok {
		switch err := errField.(type) {
		case error:
			// Try to extract oops error information
			var oopsError oops.OopsError
			if errors.As(err, &oopsError) {
				// Enhance the log entry with oops error data
				oopsErrorToEntryData(&oopsError, entry)
			}
		case any:
			// Handle non-error types in the error field
			// This case is included for completeness but typically
			// the error field should contain actual error instances
		}
	}

	// Delegate to the underlying formatter for final formatting
	return f.formatter.Format(entry)
}

// oopsErrorToEntryData extracts error information from an oops.OopsError
// and adds it to a logrus.Entry for enhanced logging.
//
// This function processes all error attributes including context, metadata,
// stack traces, and timing information, and adds them to the log entry's
// data map. The function also handles conditional inclusion of stack traces
// and source code fragments based on the log level.
//
// Performance: This function performs map operations and data extraction,
// which may have some overhead for errors with large amounts of context
// or metadata.
func oopsErrorToEntryData(err *oops.OopsError, entry *logrus.Entry) {
	// Update the log entry timestamp with the error timestamp
	// This ensures that the log entry reflects when the error actually occurred
	entry.Time = err.Time()

	// Convert the error to a map representation containing all error data
	payload := err.ToMap()

	// Conditionally remove stack traces and source code fragments for non-error levels
	// This helps reduce log noise for informational and warning messages while
	// preserving detailed debugging information for actual errors
	if entry.Level < logrus.ErrorLevel {
		delete(payload, "stacktrace")
		delete(payload, "sources")
	}

	// Create a new map for the entry's data
	newData := make(logrus.Fields, len(entry.Data)+len(payload))

	// Copy the original data
	for k, v := range entry.Data {
		newData[k] = v
	}

	// Add all error data to the log entry
	// This includes context, metadata, user information, timing, and other
	// error attributes that were captured when the error was created
	for k, v := range payload {
		newData[k] = v
	}

	entry.Data = newData
}
