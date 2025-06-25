package oops

import (
	"context"

	"github.com/samber/lo"
)

// coalesceOrEmpty returns the first non-zero value from the provided arguments,
// or the zero value of the type if all arguments are zero values.
//
// This function is a wrapper around lo.Coalesce that simplifies the return
// value handling by always returning a value (either the first non-zero value
// or the zero value of the type).
//
// Example usage:
//
//	result := coalesceOrEmpty("", "default", "fallback") // returns "default"
//	result := coalesceOrEmpty(0, 42, 100)               // returns 42
//	result := coalesceOrEmpty("", "", "")               // returns ""
func coalesceOrEmpty[T comparable](v ...T) T {
	result, _ := lo.Coalesce(v...)
	return result
}

// contextValueOrNil safely extracts a value from a Go context, handling
// nil contexts and nil values appropriately.
//
// This function provides a safe way to extract values from contexts without
// causing panics. It handles the case where the context is nil and also
// properly handles nil values stored in the context.
//
// The function is particularly useful when working with context values that
// might be nil or when the context itself might be nil, which can happen
// in certain edge cases in web applications or when contexts are not
// properly initialized.
//
// Example usage:
//
//	value := contextValueOrNil(ctx, "user_id")
//	if value != nil {
//	  userID := value.(string)
//	  // Use userID safely
//	}
//
//	// Safe even with nil context
//	value := contextValueOrNil(nil, "key") // returns nil
func contextValueOrNil(ctx context.Context, k any) any {
	if ctx == nil {
		return nil
	}

	v := ctx.Value(k)
	if v == nil {
		return nil
	}

	return v
}
