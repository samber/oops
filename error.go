//nolint:bodyclose
package oops

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"runtime"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
)

// Global configuration variables that control the behavior of error handling.
var (
	// SourceFragmentsHidden controls whether source code fragments are included in error output.
	// When true, source code context around error locations is hidden to reduce output size.
	SourceFragmentsHidden = true

	// DereferencePointers controls whether pointer values in error context are automatically
	// dereferenced when converting to map representations. This can be useful for logging
	// but may cause issues with nil pointers.
	DereferencePointers = true

	// Local specifies the timezone used for error timestamps. Defaults to UTC.
	Local *time.Location = time.UTC
)

// Type assertions to ensure OopsError implements required interfaces.
var (
	_ error          = (*OopsError)(nil)
	_ slog.LogValuer = (*OopsError)(nil)
)

// OopsError represents an enhanced error with additional contextual information.
// It implements the standard error interface while providing rich metadata for
// debugging, logging, and error handling.
type OopsError struct {
	// Core error information
	err      error         // The underlying error being wrapped
	msg      string        // Additional error message
	code     any           // Machine-readable error code/slug (any JSON/log-friendly type)
	time     time.Time     // When the error occurred
	duration time.Duration // Duration associated with the error

	// Contextual information
	domain  string         // Feature category or domain (e.g., "auth", "database")
	tags    []string       // Tags for categorizing the error
	context map[string]any // Key-value pairs for additional context

	// Tracing information
	trace string // Transaction/trace/correlation ID
	span  string // Current span identifier

	// Developer-facing information
	hint   string // Debugging hint for developers
	public string // User-safe error message
	owner  string // Team/person responsible for handling this error

	// User and tenant information
	userID     string         // User identifier
	userData   map[string]any // User-related data
	tenantID   string         // Tenant identifier
	tenantData map[string]any // Tenant-related data

	// HTTP request/response information
	req *lo.Tuple2[*http.Request, bool]  // HTTP request with body inclusion flag
	res *lo.Tuple2[*http.Response, bool] // HTTP response with body inclusion flag

	// Stack trace information
	stacktrace *oopsStacktrace // Captured stack trace
}

// Unwrap returns the underlying error that this OopsError wraps.
// This method implements the errors.Wrapper interface.
func (o OopsError) Unwrap() error {
	return o.err
}

// Is checks if this error matches the target error.
// This method implements the errors.Is interface for error comparison.
func (o OopsError) Is(err error) bool {
	if _, ok := err.(OopsError); ok {
		return true
	}

	return errors.Is(o.err, err)
}

// Error returns the error message without additional context.
// This method implements the error interface.
// If the error wraps another error, it returns "message: wrapped_error".
// Otherwise, it returns just the message.
func (o OopsError) Error() string {
	if o.err != nil {
		if o.msg == "" {
			return o.err.Error()
		}

		return fmt.Sprintf("%s: %s", o.msg, o.err.Error())
	}

	return o.msg
}

// Code returns the error code from the deepest error in the chain.
// Error codes are machine-readable identifiers that can be used for
// programmatic error handling and cross-service error correlation.
func (o OopsError) Code() any {
	return getDeepestErrorCode(o)
}

func getDeepestErrorCode(err OopsError) any {
	if err.err == nil {
		return err.code
	}

	if child, ok := AsOops(err.err); ok {
		deepest := getDeepestErrorCode(child)
		if deepest != nil {
			return deepest
		}
	}

	return err.code
}

// Time returns the timestamp when the error occurred.
// Returns the time from the deepest error in the chain.
func (o OopsError) Time() time.Time {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) time.Time {
			return e.time
		},
	)
}

// Duration returns the duration associated with the error.
// Returns the duration from the deepest error in the chain.
func (o OopsError) Duration() time.Duration {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) time.Duration {
			return e.duration
		},
	)
}

// Domain returns the domain/feature category of the error.
// Returns the domain from the deepest error in the chain.
func (o OopsError) Domain() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.domain
		},
	)
}

// Tags returns all unique tags from the error chain.
// Tags are merged from all errors in the chain and deduplicated.
func (o OopsError) Tags() []string {
	tags := []string{}

	recursive(o, func(e OopsError) bool {
		tags = append(tags, e.tags...)
		return true
	})

	return lo.Uniq(tags)
}

// HasTag checks if the error or any of its wrapped errors contain the specified tag.
// This is useful for conditional error handling based on error categories.
func (o OopsError) HasTag(tag string) bool {
	found := false
	recursive(o, func(e OopsError) bool {
		if lo.Contains(e.tags, tag) {
			found = true
		}
		return !found
	})

	return found
}

// Context returns a flattened key-value context map from the error chain.
// Context from all errors in the chain is merged, with later errors taking precedence.
// Pointer values are dereferenced if DereferencePointers is enabled.
// Lazy evaluation functions are executed to get their values.
func (o OopsError) Context() map[string]any {
	return dereferencePointers(
		lazyMapEvaluation(
			mergeNestedErrorMap(
				o,
				func(e OopsError) map[string]any {
					return e.context
				},
			),
		),
	)
}

// Trace returns the transaction/trace/correlation ID.
// If no trace ID is set, generates a new ULID-based trace ID.
// Returns the trace ID from the deepest error in the chain.
func (o OopsError) Trace() string {
	trace := getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.trace
		},
	)

	if trace != "" {
		return trace
	}

	return ulid.Make().String()
}

// Span returns the current span identifier.
// Unlike other attributes, span returns the current error's span, not the deepest one.
func (o OopsError) Span() string {
	return o.span
}

// Hint returns a debugging hint for resolving the error.
// Returns the hint from the deepest error in the chain.
func (o OopsError) Hint() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.hint
		},
	)
}

// Public returns a user-safe error message.
// Returns the public message from the deepest error in the chain.
func (o OopsError) Public() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.public
		},
	)
}

// Owner returns the name/email of the person/team responsible for handling this error.
// Returns the owner from the deepest error in the chain.
func (o OopsError) Owner() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.owner
		},
	)
}

// User returns the user ID and associated user data.
// Returns the user information from the deepest error in the chain.
func (o OopsError) User() (string, map[string]any) {
	userID := getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.userID
		},
	)

	userData := dereferencePointers(
		lazyMapEvaluation(
			mergeNestedErrorMap(
				o,
				func(e OopsError) map[string]any {
					return e.userData
				},
			),
		),
	)

	return userID, userData
}

// Tenant returns the tenant ID and associated tenant data.
// Returns the tenant information from the deepest error in the chain.
func (o OopsError) Tenant() (string, map[string]any) {
	tenantID := getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.tenantID
		},
	)

	tenantData := dereferencePointers(
		lazyMapEvaluation(
			mergeNestedErrorMap(
				o,
				func(e OopsError) map[string]any {
					return e.tenantData
				},
			),
		),
	)

	return tenantID, tenantData
}

// Request returns the associated HTTP request.
// Returns the request from the deepest error in the chain.
func (o OopsError) Request() *http.Request {
	t := o.request()
	if t != nil {
		return t.A
	}

	return nil
}

// request returns the internal request tuple with body inclusion flag.
func (o OopsError) request() *lo.Tuple2[*http.Request, bool] {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) *lo.Tuple2[*http.Request, bool] {
			return e.req
		},
	)
}

// Response returns the associated HTTP response.
// Returns the response from the deepest error in the chain.
func (o OopsError) Response() *http.Response {
	t := o.response()
	if t != nil {
		return t.A
	}

	return nil
}

// response returns the internal response tuple with body inclusion flag.
func (o OopsError) response() *lo.Tuple2[*http.Response, bool] {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) *lo.Tuple2[*http.Response, bool] {
			return e.res
		},
	)
}

// Stacktrace returns a formatted string representation of the error's stack trace.
// The stack trace shows the call hierarchy leading to the error, excluding
// frames from the Go standard library and this package.
// The stacktrace is basically written from the bottom to the top, in order to dedup frames.
// It support recursive code.
func (o OopsError) Stacktrace() string {
	blocks := []lo.Tuple3[error, string, []oopsStacktraceFrame]{}
	recursive(o, func(e OopsError) bool {
		if e.stacktrace != nil && len(e.stacktrace.frames) > 0 {
			blocks = append(blocks, lo.T3(
				e.err,
				e.msg,
				e.stacktrace.frames,
			))
		}
		return true
	})

	if len(blocks) == 0 {
		return ""
	}

	return "Oops: " + strings.Join(framesToStacktraceBlocks(blocks), "\nThrown: ")
}

// StackFrames returns the raw stack frames as runtime.Frame objects.
// This is useful for custom stack trace formatting or analysis.
func (o OopsError) StackFrames() []runtime.Frame {
	if o.stacktrace == nil || len(o.stacktrace.frames) == 0 {
		return nil
	}

	frames := make([]runtime.Frame, len(o.stacktrace.frames))
	for i, frame := range o.stacktrace.frames {
		frames[i] = runtime.Frame{
			PC:       frame.pc,
			File:     frame.file,
			Line:     frame.line,
			Function: frame.function,
		}
	}

	return frames
}

// Sources returns formatted source code fragments around the error location.
// This provides context about the code that caused the error, which is
// particularly useful for debugging. The output includes line numbers and
// highlights the exact line where the error occurred.
func (o OopsError) Sources() string {
	blocks := []lo.Tuple2[string, *oopsStacktrace]{}

	recursive(o, func(e OopsError) bool {
		if e.stacktrace != nil && len(e.stacktrace.frames) > 0 {
			blocks = append(blocks, lo.T2(
				e.msg,
				e.stacktrace,
			))
		}
		return true
	})

	if len(blocks) == 0 {
		return ""
	}

	return "Oops: " + strings.Join(
		framesToSourceBlocks(blocks),
		"\n\nThrown: ",
	)
}

// LogValuer returns a slog.Value representation of the error.
// This method implements the slog.LogValuer interface for structured logging.
//
// Deprecated: Use LogValue instead.
func (o OopsError) LogValuer() slog.Value {
	return o.LogValue()
}

// LogValue returns a slog.Value representation of the error for structured logging.
// This method implements the slog.LogValuer interface and provides a flattened
// representation of the error's context and metadata suitable for logging systems.
func (o OopsError) LogValue() slog.Value { //nolint:gocyclo
	attrs := []slog.Attr{slog.String("message", o.msg)}

	if err := o.Error(); err != "" {
		attrs = append(attrs, slog.String("err", err))
	}

	if code := o.Code(); code != nil {
		attrs = append(attrs, slog.Any("code", code))
	}

	if t := o.Time(); t != (time.Time{}) {
		attrs = append(attrs, slog.Time("time", t.In(Local)))
	}

	if duration := o.Duration(); duration != 0 {
		attrs = append(attrs, slog.Duration("duration", duration))
	}

	if domain := o.Domain(); domain != "" {
		attrs = append(attrs, slog.String("domain", domain))
	}

	if tags := o.Tags(); len(tags) > 0 {
		attrs = append(attrs, slog.Any("tags", tags))
	}

	if trace := o.Trace(); trace != "" {
		attrs = append(attrs, slog.String("trace", trace))
	}

	// if span := o.Span(); span != "" {
	// 	attrs = append(attrs, slog.String("span", span))
	// }

	if hint := o.Hint(); hint != "" {
		attrs = append(attrs, slog.String("hint", hint))
	}

	if public := o.Public(); public != "" {
		attrs = append(attrs, slog.String("public", public))
	}

	if owner := o.Owner(); owner != "" {
		attrs = append(attrs, slog.String("owner", owner))
	}

	if context := o.Context(); len(context) > 0 {
		attrs = append(attrs,
			slog.Group(
				"context",
				lo.ToAnySlice(
					lo.MapToSlice(context, slog.Any),
				)...,
			),
		)
	}

	if userID, userData := o.User(); userID != "" || len(userData) > 0 {
		userPayload := []slog.Attr{}
		if userID != "" {
			userPayload = append(userPayload, slog.String("id", userID))
			userPayload = append(
				userPayload,
				lo.MapToSlice(userData, slog.Any)...,
			)
		}

		attrs = append(attrs, slog.Group("user", lo.ToAnySlice(userPayload)...))
	}

	if tenantID, tenantData := o.Tenant(); tenantID != "" || len(tenantData) > 0 {
		tenantPayload := []slog.Attr{}
		if tenantID != "" {
			tenantPayload = append(tenantPayload, slog.String("id", tenantID))
			tenantPayload = append(
				tenantPayload,
				lo.MapToSlice(tenantData, slog.Any)...,
			)
		}

		attrs = append(attrs, slog.Group("tenant", lo.ToAnySlice(tenantPayload)...))
	}

	if req := o.request(); req != nil {
		dump, e := httputil.DumpRequestOut(req.A, req.B)
		if e == nil {
			attrs = append(attrs, slog.String("request", string(dump)))
		}
	}

	if res := o.response(); res != nil {
		dump, e := httputil.DumpResponse(res.A, res.B)
		if e == nil {
			attrs = append(attrs, slog.String("response", string(dump)))
		}
	}

	if stacktrace := o.Stacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String("stacktrace", stacktrace))
	}

	if sources := o.Sources(); sources != "" && !SourceFragmentsHidden {
		attrs = append(attrs, slog.String("sources", sources))
	}

	return slog.GroupValue(attrs...)
}

// ToMap converts the error to a map representation suitable for JSON serialization.
// This method provides a flattened view of all error attributes and is useful
// for logging, debugging, and cross-service error transmission.
func (o OopsError) ToMap() map[string]any { //nolint:gocyclo
	payload := map[string]any{}

	if err := o.Error(); err != "" {
		payload["error"] = err
	}

	if code := o.Code(); code != nil {
		payload["code"] = code
	}

	if t := o.Time(); t != (time.Time{}) {
		payload["time"] = t.In(Local)
	}

	if duration := o.Duration(); duration != 0 {
		payload["duration"] = duration.String()
	}

	if domain := o.Domain(); domain != "" {
		payload["domain"] = domain
	}

	if tags := o.Tags(); len(tags) > 0 {
		payload["tags"] = tags
	}

	if context := o.Context(); len(context) > 0 {
		payload["context"] = context
	}

	if trace := o.Trace(); trace != "" {
		payload["trace"] = trace
	}

	// if span := o.Span(); span != "" {
	// 	payload["span"] = span
	// }

	if hint := o.Hint(); hint != "" {
		payload["hint"] = hint
	}

	if public := o.Public(); public != "" {
		payload["public"] = public
	}

	if owner := o.Owner(); owner != "" {
		payload["owner"] = owner
	}

	if userID, userData := o.User(); userID != "" || len(userData) > 0 {
		user := lo.Assign(map[string]any{}, userData)
		if userID != "" {
			user["id"] = userID
		}

		payload["user"] = user
	}

	if tenantID, tenantData := o.Tenant(); tenantID != "" || len(tenantData) > 0 {
		tenant := lo.Assign(map[string]any{}, tenantData)
		if tenantID != "" {
			tenant["id"] = tenantID
		}

		payload["tenant"] = tenant
	}

	if req := o.request(); req != nil {
		dump, e := httputil.DumpRequestOut(req.A, req.B)
		if e == nil {
			payload["request"] = string(dump)
		}
	}

	if res := o.response(); res != nil {
		dump, e := httputil.DumpResponse(res.A, res.B)
		if e == nil {
			payload["response"] = string(dump)
		}
	}

	if stacktrace := o.Stacktrace(); stacktrace != "" {
		payload["stacktrace"] = stacktrace
	}

	if sources := o.Sources(); sources != "" && !SourceFragmentsHidden {
		payload["sources"] = sources
	}

	return payload
}

// MarshalJSON implements the json.Marshaler interface.
// This allows OopsError to be directly serialized to JSON.
func (o OopsError) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.ToMap())
}

// Format implements the fmt.Formatter interface for custom formatting.
// Supports the following format verbs:
// - %v: standard error message
// - %+v: verbose format with stack trace and context
// - %#v: Go syntax representation.
func (o OopsError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			// Verbose format with stack trace and context
			_, _ = fmt.Fprint(s, o.formatVerbose())
		} else {
			// Standard format
			_, _ = fmt.Fprint(s, o.formatSummary())
		}
	case 's':
		_, _ = fmt.Fprint(s, o.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", o.Error())
	default:
		_, _ = fmt.Fprint(s, o.formatSummary())
	}
}

// formatVerbose returns a detailed string representation of the error
// including all context, stack trace, and source code fragments.
func (o *OopsError) formatVerbose() string { //nolint:gocyclo
	var output strings.Builder
	_, _ = fmt.Fprintf(&output, "Oops: %s\n", o.Error())

	if code := o.Code(); code != nil {
		_, _ = fmt.Fprintf(&output, "Code: %v\n", code)
	}

	if t := o.Time(); t != (time.Time{}) {
		_, _ = fmt.Fprintf(&output, "Time: %s\n", t.In(Local))
	}

	if duration := o.Duration(); duration != 0 {
		_, _ = fmt.Fprintf(&output, "Duration: %s\n", duration.String())
	}

	if domain := o.Domain(); domain != "" {
		_, _ = fmt.Fprintf(&output, "Domain: %s\n", domain)
	}

	if tags := o.Tags(); len(tags) > 0 {
		_, _ = fmt.Fprintf(&output, "Tags: %s\n", strings.Join(tags, ", "))
	}

	if trace := o.Trace(); trace != "" {
		_, _ = fmt.Fprintf(&output, "Trace: %s\n", trace)
	}

	// if span := o.Span(); span != "" {
	// 	_, _ = fmt.Fprintf(&output,"Span: %s\n", span)
	// }

	if hint := o.Hint(); hint != "" {
		_, _ = fmt.Fprintf(&output, "Hint: %s\n", hint)
	}

	if owner := o.Owner(); owner != "" {
		_, _ = fmt.Fprintf(&output, "Owner: %s\n", owner)
	}

	if context := o.Context(); len(context) > 0 {
		output.WriteString("Context:\n")
		for k, v := range context {
			_, _ = fmt.Fprintf(&output, "  * %s: %v\n", k, v)
		}
	}

	if userID, userData := o.User(); userID != "" || len(userData) > 0 {
		output.WriteString("User:\n")

		if userID != "" {
			_, _ = fmt.Fprintf(&output, "  * id: %s\n", userID)
		}

		for k, v := range userData {
			_, _ = fmt.Fprintf(&output, "  * %s: %v\n", k, v)
		}
	}

	if tenantID, tenantData := o.Tenant(); tenantID != "" || len(tenantData) > 0 {
		output.WriteString("Tenant:\n")

		if tenantID != "" {
			_, _ = fmt.Fprintf(&output, "  * id: %s\n", tenantID)
		}

		for k, v := range tenantData {
			_, _ = fmt.Fprintf(&output, "  * %s: %v\n", k, v)
		}
	}

	if req := o.request(); req != nil {
		dump, e := httputil.DumpRequestOut(req.A, req.B)
		if e == nil {
			lines := strings.Split(string(dump), "\n")
			lines = lo.Map(lines, func(line string, _ int) string {
				return "  * " + line
			})
			_, _ = fmt.Fprintf(&output, "Request:\n%s\n", strings.Join(lines, "\n"))
		}
	}

	if res := o.response(); res != nil {
		dump, e := httputil.DumpResponse(res.A, res.B)
		if e == nil {
			lines := strings.Split(string(dump), "\n")
			lines = lo.Map(lines, func(line string, _ int) string {
				return "  * " + line
			})
			_, _ = fmt.Fprintf(&output, "Response:\n%s\n", strings.Join(lines, "\n"))
		}
	}

	if stacktrace := o.Stacktrace(); stacktrace != "" {
		lines := strings.Split(stacktrace, "\n")
		stacktrace = "  " + strings.Join(lines, "\n  ")
		_, _ = fmt.Fprintf(&output, "Stacktrace:\n%s\n", stacktrace)
	}

	if sources := o.Sources(); sources != "" && !SourceFragmentsHidden {
		_, _ = fmt.Fprintf(&output, "Sources:\n%s\n", sources)
	}

	return output.String()
}

// formatSummary returns a brief summary of the error for logging.
func (o *OopsError) formatSummary() string {
	return o.Error()
}

// recursive is a helper function that traverses the error chain
// and applies a function to each OopsError in the chain.
func recursive(err OopsError, tap func(OopsError) bool) {
	if !tap(err) {
		return
	}

	if err.err == nil {
		return
	}

	if child, ok := AsOops(err.err); ok {
		recursive(child, tap)
	}
}

// // recursive is a helper function that traverses the error chain
// // and applies a function to each OopsError in the chain.
// func recursiveBackward(err OopsError, tap func(OopsError)) {
// 	if err.err == nil {
// 		tap(err)
// 		return
// 	}

// 	if child, ok := AsOops(err.err); ok {
// 		recursiveBackward(child, tap)
// 	}

// 	tap(err)
// }
