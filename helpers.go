package oops

import "errors"

// AsOops checks if an error is an oops.OopsError instance and returns it if so.
// This function is an alias to errors.As and provides a convenient way to
// type-assert errors to oops.OopsError without importing the errors package.
//
// This function is useful when you need to access the rich metadata and
// context information stored in oops.OopsError instances, such as error
// codes, stack traces, user information, or custom context data.
//
// Example usage:
//
//	err := someFunction()
//	if oopsErr, ok := oops.AsOops(err); ok {
//	  // Access oops-specific information
//	  fmt.Printf("Error code: %s\n", oopsErr.Code())
//	  fmt.Printf("Domain: %s\n", oopsErr.Domain())
//	  fmt.Printf("Stack trace: %s\n", oopsErr.Stacktrace())
//
//	  // Check for specific tags
//	  if oopsErr.HasTag("critical") {
//	    // Handle critical errors differently
//	    sendAlert(oopsErr)
//	  }
//	}
//
//	// Chain with other error handling
//	if oopsErr, ok := oops.AsOops(err); ok && oopsErr.Code() == "database_error" {
//	  // Handle database errors specifically
//	  retryOperation()
//	}
func AsOops(err error) (OopsError, bool) {
	var e OopsError
	ok := errors.As(err, &e)
	return e, ok
}
