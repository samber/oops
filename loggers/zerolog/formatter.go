package oopszerolog

import (
	"github.com/rs/zerolog"
	"github.com/samber/oops"
)

// OopsStackMarshaller returns a marshaller function that extracts stack trace
// information from oops.OopsError instances for zerolog logging.
//
// This function can be used with zerolog's error marshalling to include
// stack traces in log output. When used with an oops.OopsError, it returns
// the formatted stack trace. For other error types, it returns the error
// as-is.
//
// Example usage:
//
//	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
//	logger.Error().
//	  Err(err).
//	  Str("stacktrace", oopszerolog.OopsStackMarshaller(err).(string)).
//	  Msg("operation failed")
func OopsStackMarshaller(err error) interface{} {
	if typedErr, ok := oops.AsOops(err); ok {
		return typedErr.Stacktrace()
	}
	return err
}

// OopsMarshalFunc returns a marshaller function that converts oops.OopsError
// instances into structured zerolog objects for logging.
//
// This function is designed to be used with zerolog's error marshalling
// capabilities. It provides a complete structured representation of oops
// errors, including all context, metadata, and stack traces.
//
// The returned marshaller implements zerolog's ObjectMarshaler interface
// and can be used directly with zerolog's logging methods.
//
// Example usage:
//
//	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
//	logger.Error().
//	  Interface("error", oopszerolog.OopsMarshalFunc(err)).
//	  Msg("operation failed")
func OopsMarshalFunc(err error) interface{} {
	if typedErr, ok := oops.AsOops(err); ok {
		return zerologErrorMarshaller{err: typedErr}
	}
	return err
}

// zerologErrorMarshaller implements zerolog's ObjectMarshaler interface
// to provide structured serialization of oops.OopsError instances.
//
// This type handles the conversion of oops error attributes into zerolog
// compatible formats, ensuring that all error context, metadata, and
// stack traces are properly serialized for logging.
type zerologErrorMarshaller struct {
	err oops.OopsError
}

// MarshalZerologObject implements zerolog's ObjectMarshaler interface.
// This method converts the oops.OopsError into a structured zerolog object
// with all relevant error information properly formatted.
//
// The method handles special cases for different types of error data:
// - Context maps are converted to nested dictionaries
// - Error values are converted to strings
// - Stack traces are excluded from the main object (handled separately)
// - All other values are serialized as-is.
func (m zerologErrorMarshaller) MarshalZerologObject(e *zerolog.Event) {
	// Convert the error to a map representation for processing
	payload := m.err.ToMap()

	// Iterate through all error attributes and format them appropriately
	for k, v := range payload {
		switch k {
		case "stacktrace":
			// Skip stacktrace in the main object - it should be handled separately
			// to avoid duplication and maintain clean log structure
		case "context":
			// Handle context as a nested dictionary for better structure
			if context, ok := v.(map[string]any); ok && len(context) > 0 {
				dict := zerolog.Dict()
				for k, v := range context {
					switch vTyped := v.(type) {
					case nil:
						// Skip nil values to keep logs clean
					case error:
						// Convert error values to strings for consistent logging
						dict = dict.Str(k, vTyped.Error())
					default:
						// Use interface marshalling for all other types
						dict = dict.Interface(k, vTyped)
					}
				}
				e.Dict(k, dict)
			}
		default:
			// For all other fields, use zerolog's interface marshalling
			// This handles most types automatically and provides good defaults
			e.Any(k, v)
		}
	}
}
