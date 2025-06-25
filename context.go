package oops

import "context"

// contextKey is a custom type for context keys to avoid collisions
// with other packages that might use string keys in the same context.
type contextKey string

// contextKeyOops is the specific key used to store OopsErrorBuilder instances
// in Go contexts. This key is used internally by the package to retrieve
// error builders that have been stored in contexts.
const contextKeyOops = contextKey("oops")

// getBuilderFromContext retrieves an OopsErrorBuilder from a Go context.
// This function is used internally to extract error builders that have been
// previously stored in contexts using WithBuilder.
//
// The function returns the builder and a boolean indicating whether a builder
// was found in the context. If no builder is found, the second return value
// will be false.
//
// Thread Safety: This function is thread-safe and can be called concurrently
// on the same context from multiple goroutines.
func getBuilderFromContext(ctx context.Context) (OopsErrorBuilder, bool) {
	b, ok := ctx.Value(contextKeyOops).(OopsErrorBuilder)
	return b, ok
}

// WithBuilder stores an OopsErrorBuilder in a Go context for later retrieval.
// This function creates a new context with the builder stored under the
// package's internal key, allowing the builder to be accessed by subsequent
// middleware or handlers in the request chain.
//
// This is particularly useful in web applications where you want to propagate
// error context (like request IDs, user information, or tracing data) through
// the entire request processing pipeline without explicitly passing the builder
// to every function.
//
// The original context is not modified; a new context is returned with the
// builder added. This follows Go's context immutability pattern.
//
// Example usage:
//
//	func middleware(next http.Handler) http.Handler {
//	  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    // Create a builder with request-specific context
//	    builder := oops.
//	      Trace(r.Header.Get("X-Request-ID")).
//	      With("user_agent", r.UserAgent())
//
//	    // Store the builder in the request context
//	    ctx := oops.WithBuilder(r.Context(), builder)
//
//	    // Pass the enhanced context to the next handler
//	    next.ServeHTTP(w, r.WithContext(ctx))
//	  })
//	}
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	  // Retrieve the builder from context and create an error
//	  err := oops.FromContext(r.Context()).
//	    In("handler").
//	    Errorf("something went wrong")
//
//	  // The error will automatically include the trace ID and user agent
//	  // that were set in the middleware
//	}
func WithBuilder(ctx context.Context, builder OopsErrorBuilder) context.Context {
	return context.WithValue(ctx, contextKeyOops, builder)
}
