package oops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/trace"
)

/**
 * Builder pattern implementation for creating rich error objects.
 *
 * The builder pattern allows for fluent, chainable error creation with contextual
 * information. This design provides a clean API for building complex error objects
 * while maintaining readability and flexibility.
 *
 * Examples:
 *
 * Basic error creation:
 *   oops.Errorf("Could not fetch users: %w", err)
 *
 * Rich error with context:
 *   oops.
 *     User("steve@apple.com", "firstname", "Samuel").
 *     Tenant("apple", "country", "us").
 *     Errorf("403 not permitted")
 *
 * Error with timing and tracing:
 *   oops.
 *     Time(requestDate).
 *     Duration(requestDuration).
 *     Trace(traceID).
 *     Errorf("Failed to execute http request")
 *
 * Error with custom context:
 *   oops.
 *     With("project_id", project.ID, "created_at", project.CreatedAt).
 *     Errorf("Could not update settings")
 *
 * Thread Safety: OopsErrorBuilder instances are not thread-safe. Each builder
 * should be used by a single goroutine. The builder methods return new instances
 * to support chaining, which provides some isolation but doesn't guarantee
 * thread safety for concurrent access to the same builder instance.
 */

// OopsErrorBuilder implements the builder pattern for creating OopsError instances.
// It provides a fluent API for setting error attributes and creating error objects.
// The builder is designed to be chainable, allowing multiple method calls in sequence.
type OopsErrorBuilder OopsError

// newBuilder creates a newBuilder OopsErrorBuilder with default values.
// This function initializes all fields to their zero values except for time,
// which is set to the current time.
func newBuilder() OopsErrorBuilder {
	return OopsErrorBuilder{
		err:      nil,
		msg:      "",
		code:     "",
		time:     time.Now(),
		duration: 0,

		// context
		domain:  "",
		tags:    []string{},
		context: map[string]any{},

		trace: "",
		span:  "",

		hint:   "",
		public: "",
		owner:  "",

		// user
		userID:     "",
		userData:   map[string]any{},
		tenantID:   "",
		tenantData: map[string]any{},

		// http
		req: nil,
		res: nil,

		// stacktrace
		stacktrace: nil,
	}
}

// copy creates a deep copy of the current builder state.
// This method is used internally to create new builder instances for chaining.
// It performs deep copying of maps to ensure that modifications to the new
// builder don't affect the original.
func (o OopsErrorBuilder) copy() OopsErrorBuilder {
	return OopsErrorBuilder{
		// err:      err,  // Not copied as it's set by error creation methods
		// msg:      o.msg, // Not copied as it's set by error creation methods
		code:     o.code,
		time:     o.time,
		duration: o.duration,

		domain:  o.domain,
		tags:    o.tags,
		context: lo.Assign(map[string]any{}, o.context), // Deep copy of context map (pointer values are not copied)

		trace: o.trace,
		span:  o.span,

		hint:   o.hint,
		public: o.public,
		owner:  o.owner,

		userID:     o.userID,
		userData:   lo.Assign(map[string]any{}, o.userData), // Deep copy of user data (pointer values are not copied)
		tenantID:   o.tenantID,
		tenantData: lo.Assign(map[string]any{}, o.tenantData), // Deep copy of tenant data (pointer values are not copied)

		req: o.req,
		res: o.res,

		// stacktrace: o.stacktrace, // Not copied as it's generated per error
	}
}

// Wrap wraps an existing error into an OopsError with the current builder's context.
// If the input error is nil, returns nil. Otherwise, creates a new OopsError that
// wraps the original error while preserving all the contextual information set
// in the builder.
//
// Example:
//
//	err := oops.
//	  Code("database_error").
//	  In("database").
//	  Wrap(originalError)
func (o OopsErrorBuilder) Wrap(err error) error {
	if err == nil {
		return nil
	}

	o2 := o.copy()
	o2.err = err
	if o2.span == "" {
		o2.span = ulid.Make().String() // Generate unique span ID if not set
	}
	o2.stacktrace = newStacktrace(o2.span) // Capture stack trace at error creation
	return OopsError(o2)
}

// Wrapf wraps an existing error with additional formatted message.
// Similar to Wrap, but adds a formatted message that describes the context
// in which the error occurred.
//
// Example:
//
//	err := oops.
//	  Code("database_error").
//	  In("database").
//	  Wrapf(originalError, "failed to execute query: %s", queryName)
func (o OopsErrorBuilder) Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	o2 := o.copy()
	o2.err = err
	o2.msg = fmt.Errorf(format, args...).Error() // Format the additional message
	if o2.span == "" {
		o2.span = ulid.Make().String()
	}
	o2.stacktrace = newStacktrace(o2.span)
	return OopsError(o2)
}

// New creates a new error with the specified message.
// This method creates a simple error without wrapping an existing one.
// The message is treated as the primary error message.
//
// Example:
//
//	err := oops.
//	  Code("validation_error").
//	  New("invalid input parameters")
func (o OopsErrorBuilder) New(message string) error {
	o2 := o.copy()
	o2.err = errors.New(message)
	if o2.span == "" {
		o2.span = ulid.Make().String()
	}
	o2.stacktrace = newStacktrace(o2.span)
	return OopsError(o2)
}

// Errorf creates a new error with a formatted message.
// Similar to New, but allows for formatted messages using printf-style formatting.
//
// Example:
//
//	err := oops.
//	  Code("validation_error").
//	  Errorf("invalid input: expected %s, got %s", expectedType, actualType)
func (o OopsErrorBuilder) Errorf(format string, args ...any) error {
	o2 := o.copy()
	o2.err = fmt.Errorf(format, args...)
	if o2.span == "" {
		o2.span = ulid.Make().String()
	}
	o2.stacktrace = newStacktrace(o2.span)
	return OopsError(o2)
}

// Join combines multiple errors into a single error.
// This method uses the standard errors.Join function to combine multiple
// errors while preserving the builder's context.
//
// Example:
//
//	err := oops.
//	  Code("multi_error").
//	  Join(err1, err2, err3)
func (o OopsErrorBuilder) Join(e ...error) error {
	return o.Wrap(errors.Join(e...))
}

// Recover handles panics and converts them to OopsError instances.
// This method executes the provided callback function and catches any panics,
// converting them to properly formatted OopsError instances with stack traces.
// If the panic payload is already an error, it wraps that error. Otherwise,
// it creates a new error from the panic value.
//
// Example:
//
//	err := oops.
//	  Code("panic_recovered").
//	  Recover(func() {
//	    // Potentially panicking code
//	    riskyOperation()
//	  })
func (o OopsErrorBuilder) Recover(cb func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = o.Wrap(e) // Wrap existing error
			} else {
				err = o.Wrap(fmt.Errorf("%v", r)) // Convert panic value to error
			}
		}
	}()

	cb()
	return err
}

// Recoverf handles panics with additional context message.
// Similar to Recover, but adds a formatted message to describe the context
// in which the panic occurred.
//
// Example:
//
//	err := oops.
//	  Code("panic_recovered").
//	  Recoverf(func() {
//	    riskyOperation()
//	  }, "panic in operation: %s", operationName)
func (o OopsErrorBuilder) Recoverf(cb func(), msg string, args ...any) (err error) {
	return o.Wrapf(o.Recover(cb), msg, args...)
}

// Assert panics if the condition is false.
// This method provides a way to add assertions to code that will panic
// with an OopsError if the condition fails. The assertion can be chained
// with other builder methods.
//
// Example:
//
//	oops.
//	  Code("assertion_failed").
//	  Assert(userID != "", "user ID cannot be empty").
//	  Assert(email != "", "user email cannot be empty").
//	  Assert(orgID != "", "user organization ID cannot be empty")
func (o OopsErrorBuilder) Assert(condition bool) OopsErrorBuilder {
	if !condition {
		panic(o.Errorf("assertion failed"))
	}

	return o // Return self for chaining
}

// Assertf panics if the condition is false with a custom message.
// Similar to Assert, but allows for a custom formatted message when
// the assertion fails.
//
// Example:
//
//	oops.
//	  Code("assertion_failed").
//	  Assertf(userID != "", "user ID cannot be empty, got: %s", userID).
//	  Assertf(email != "", "user email cannot be empty, got: %s", email).
//	  Assertf(orgID != "", "user organization ID cannot be empty, got: %s", orgID)
func (o OopsErrorBuilder) Assertf(condition bool, msg string, args ...any) OopsErrorBuilder {
	if !condition {
		panic(o.Errorf(msg, args...))
	}

	return o // Return self for chaining
}

// Code sets a machine-readable error code or slug.
// Error codes are useful for programmatic error handling and cross-service
// error correlation. They should be consistent and well-documented.
//
// Example:
//
//	oops.Code("database_connection_failed").Errorf("connection timeout")
func (o OopsErrorBuilder) Code(code string) OopsErrorBuilder {
	o2 := o.copy()
	o2.code = code
	return o2
}

// Time sets the timestamp when the error occurred.
// If not set, the error will use the current time when created.
//
// Example:
//
//	oops.Time(time.Now()).Errorf("operation failed")
func (o OopsErrorBuilder) Time(time time.Time) OopsErrorBuilder {
	o2 := o.copy()
	o2.time = time
	return o2
}

// Since calculates the duration since the specified time.
// This is useful for measuring how long an operation took before failing.
//
// Example:
//
//	start := time.Now()
//	// ... perform operation ...
//	oops.Since(start).Errorf("operation timed out")
func (o OopsErrorBuilder) Since(t time.Time) OopsErrorBuilder {
	o2 := o.copy()
	o2.duration = time.Since(t)
	return o2
}

// Duration sets the duration associated with the error.
// This is useful for errors that are related to timeouts or performance issues.
//
// Example:
//
//	oops.Duration(5 * time.Second).Errorf("request timeout")
func (o OopsErrorBuilder) Duration(duration time.Duration) OopsErrorBuilder {
	o2 := o.copy()
	o2.duration = duration
	return o2
}

// In sets the domain or feature category for the error.
// Domains help categorize errors by the part of the system they relate to.
//
// Example:
//
//	oops.In("database").Errorf("connection failed")
func (o OopsErrorBuilder) In(domain string) OopsErrorBuilder {
	o2 := o.copy()
	o2.domain = domain
	return o2
}

// Tags adds multiple tags for categorizing the error.
// Tags are useful for filtering and grouping errors in monitoring systems.
//
// Example:
//
//	oops.Tags("auth", "permission", "critical").Errorf("access denied")
func (o OopsErrorBuilder) Tags(tags ...string) OopsErrorBuilder {
	o2 := o.copy()
	o2.tags = append(o2.tags, tags...)
	return o2
}

// With adds key-value pairs to the error context.
// Context values are useful for debugging and provide additional information
// about the error. Values can be of any type and will be serialized appropriately.
//
// Performance: Context values are stored in a map and processed during error
// creation. Large numbers of context values may impact performance later, but not
// during error creation.
//
// Example:
//
//	oops.With("user_id", 123, "operation", "create").Errorf("validation failed")
func (o OopsErrorBuilder) With(kv ...any) OopsErrorBuilder {
	o2 := o.copy()

	// Process key-value pairs in chunks of 2
	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			key, ok := kv[i].(string)
			if ok {
				o2.context[key] = kv[i+1]
			}
		}
	}

	return o2
}

// WithContext extracts values from a Go context and adds them to the error context.
// This is useful for propagating context values through error chains.
//
// Example:
//
//	oops.WithContext(ctx, "request_id", "user_id").Errorf("operation failed")
func (o OopsErrorBuilder) WithContext(ctx context.Context, keys ...any) OopsErrorBuilder {
	o2 := o.copy()

	for i := 0; i < len(keys); i++ {
		switch k := keys[i].(type) {
		case fmt.Stringer:
			o2.context[k.String()] = contextValueOrNil(ctx, k.String())
		case string:
			o2.context[k] = contextValueOrNil(ctx, k)
		case *string:
			o2.context[*k] = contextValueOrNil(ctx, *k)
		default:
			o2.context[fmt.Sprint(k)] = contextValueOrNil(ctx, k)
		}
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		o2.trace = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		o2.span = spanCtx.SpanID().String()
	}

	return o2
}

// Trace sets a transaction, trace, or correlation ID.
// This is useful for distributed tracing and correlating errors across services.
//
// Example:
//
//	oops.Trace("req-123-456").Errorf("service call failed")
func (o OopsErrorBuilder) Trace(trace string) OopsErrorBuilder {
	o2 := o.copy()
	o2.trace = trace
	return o2
}

// Span sets the current span identifier.
// Spans represent units of work and are useful for distributed tracing.
//
// Example:
//
//	oops.Span("database-query").Errorf("query failed")
func (o OopsErrorBuilder) Span(span string) OopsErrorBuilder {
	o2 := o.copy()
	o2.span = span
	return o2
}

// Hint provides a debugging hint for resolving the error.
// Hints should provide actionable guidance for developers.
//
// Example:
//
//	oops.Hint("Check database connection and credentials").Errorf("connection failed")
func (o OopsErrorBuilder) Hint(hint string) OopsErrorBuilder {
	o2 := o.copy()
	o2.hint = hint
	return o2
}

// Public sets a user-safe error message.
// This message should be safe to display to end users without exposing
// internal system details.
//
// Example:
//
//	oops.Public("Unable to process your request").Errorf("internal server error")
func (o OopsErrorBuilder) Public(public string) OopsErrorBuilder {
	o2 := o.copy()
	o2.public = public
	return o2
}

// Owner sets the person or team responsible for handling this error.
// This is useful for alerting and error routing.
//
// Example:
//
//	oops.Owner("database-team@company.com").Errorf("connection failed")
func (o OopsErrorBuilder) Owner(owner string) OopsErrorBuilder {
	o2 := o.copy()
	o2.owner = owner
	return o2
}

// User adds user information to the error context.
// This method accepts a user ID followed by key-value pairs for user data.
//
// Example:
//
//	oops.User("user-123", "firstname", "John", "lastname", "Doe").Errorf("permission denied")
func (o OopsErrorBuilder) User(userID string, userData ...any) OopsErrorBuilder {
	o2 := o.copy()
	o2.userID = userID

	// Process user data key-value pairs
	for i := 0; i < len(userData); i += 2 {
		if i+1 < len(userData) {
			key, ok := userData[i].(string)
			if ok {
				o2.userData[key] = userData[i+1]
			}
		}
	}

	return o2
}

// Tenant adds tenant information to the error context.
// This method accepts a tenant ID followed by key-value pairs for tenant data.
//
// Example:
//
//	oops.Tenant("tenant-456", "name", "Acme Corp", "plan", "premium").Errorf("quota exceeded")
func (o OopsErrorBuilder) Tenant(tenantID string, tenantData ...any) OopsErrorBuilder {
	o2 := o.copy()
	o2.tenantID = tenantID

	// Process tenant data key-value pairs
	for i := 0; i < len(tenantData); i += 2 {
		if i+1 < len(tenantData) {
			key, ok := tenantData[i].(string)
			if ok {
				o2.tenantData[key] = tenantData[i+1]
			}
		}
	}

	return o2
}

// Request adds HTTP request information to the error context.
// The withBody parameter controls whether the request body is included.
// Including request bodies may impact performance and memory usage.
//
// Example:
//
//	oops.Request(req, true).Errorf("request processing failed")
func (o OopsErrorBuilder) Request(req *http.Request, withBody bool) OopsErrorBuilder {
	o2 := o.copy()
	o2.req = lo.ToPtr(lo.T2(req, withBody))
	return o2
}

// Response adds HTTP response information to the error context.
// The withBody parameter controls whether the response body is included.
// Including response bodies may impact performance and memory usage.
//
// Example:
//
//	oops.Response(res, false).Errorf("response processing failed")
//
//nolint:bodyclose
func (o OopsErrorBuilder) Response(res *http.Response, withBody bool) OopsErrorBuilder {
	o2 := o.copy()
	o2.res = lo.ToPtr(lo.T2(res, withBody))
	return o2
}
