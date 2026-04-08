package oops

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Wrap wraps an error into an `oops.OopsError` object that satisfies `error`.
func Wrap(err error) error {
	if err == nil {
		return nil
	}

	return newBuilder().Wrap(err)
}

// Wrapf wraps an error into an `oops.OopsError` object that satisfies `error` and formats an error message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return newBuilder().Wrapf(err, format, args...)
}

// New returns `oops.OopsError` object that satisfies `error`.
func New(message string) error {
	return newBuilder().New(message)
}

// Errorf formats an error and returns `oops.OopsError` object that satisfies `error`.
func Errorf(format string, args ...any) error {
	return newBuilder().Errorf(format, args...)
}

func FromContext(ctx context.Context) OopsErrorBuilder {
	builder, ok := getBuilderFromContext(ctx)
	if !ok {
		return newBuilder()
	}

	return builder
}

func Join(e ...error) error {
	return newBuilder().Join(e...)
}

// Recover handle panic and returns `oops.OopsError` object that satisfies `error`.
func Recover(cb func()) (err error) {
	return newBuilder().Recover(cb)
}

// Recoverf handle panic and returns `oops.OopsError` object that satisfies `error` and formats an error message.
func Recoverf(cb func(), msg string, args ...any) (err error) {
	return newBuilder().Recoverf(cb, msg, args...)
}

// Assert panics if condition is false. Panic payload will be of type oops.OopsError.
// Assertions can be chained.
func Assert(condition bool) OopsErrorBuilder {
	return newBuilder().Assert(condition)
}

// Assertf panics if condition is false. Panic payload will be of type oops.OopsError.
// Assertions can be chained.
func Assertf(condition bool, msg string, args ...any) OopsErrorBuilder {
	return newBuilder().Assertf(condition, msg, args...)
}

// Code set a code or slug that describes the error.
// Error messages are intended to be read by humans, but such code is expected to
// be read by machines and even transported over different services.
func Code(code any) OopsErrorBuilder {
	return newBuilder().Code(code)
}

// Time set the error time.
// Default: `time.Now()`.
func Time(time time.Time) OopsErrorBuilder {
	return newBuilder().Time(time)
}

// Since set the error duration.
func Since(time time.Time) OopsErrorBuilder {
	return newBuilder().Since(time)
}

// Duration set the error duration.
func Duration(duration time.Duration) OopsErrorBuilder {
	return newBuilder().Duration(duration)
}

// In set the feature category or domain.
func In(domain string) OopsErrorBuilder {
	return newBuilder().In(domain)
}

// Tags adds multiple tags, describing the feature returning an error.
func Tags(tags ...string) OopsErrorBuilder {
	return newBuilder().Tags(tags...)
}

// Trace set a transaction id, trace id or correlation id...
func Trace(trace string) OopsErrorBuilder {
	return newBuilder().Trace(trace)
}

// Span represents a unit of work or operation.
func Span(span string) OopsErrorBuilder {
	return newBuilder().Span(span)
}

// With supplies a list of attributes declared by pair of key+value.
func With(kv ...any) OopsErrorBuilder {
	return newBuilder().With(kv...)
}

// WithContext supplies a list of values declared in context.
func WithContext(ctx context.Context, keys ...any) OopsErrorBuilder {
	return newBuilder().WithContext(ctx, keys...)
}

// Hint set a hint for faster debugging.
func Hint(hint string) OopsErrorBuilder {
	return newBuilder().Hint(hint)
}

// Public sets a message that is safe to show to an end user.
func Public(public string) OopsErrorBuilder {
	return newBuilder().Public(public)
}

// Owner set the name/email of the colleague/team responsible for handling this error.
// Useful for alerting purpose.
func Owner(owner string) OopsErrorBuilder {
	return newBuilder().Owner(owner)
}

// User supplies a user id with optional attributes.
// Attributes can be provided as alternating string key/value pairs,
// map[string]any values, and slog.Attr values.
func User(userID string, data ...any) OopsErrorBuilder {
	return newBuilder().User(userID, data...)
}

// Tenant supplies a tenant id with optional attributes.
// Attributes can be provided as alternating string key/value pairs,
// map[string]any values, and slog.Attr values.
func Tenant(tenantID string, data ...any) OopsErrorBuilder {
	return newBuilder().Tenant(tenantID, data...)
}

// Request supplies a http.Request.
func Request(req *http.Request, withBody bool) OopsErrorBuilder {
	return newBuilder().Request(req, withBody)
}

// Response supplies a http.Response.
func Response(res *http.Response, withBody bool) OopsErrorBuilder {
	return newBuilder().Response(res, withBody)
}

// CallerSkip sets the number of additional callers to skip when capturing
// the stack trace. This is useful when oops is wrapped in helper functions.
func CallerSkip(skip int) OopsErrorBuilder {
	return newBuilder().CallerSkip(skip)
}

// FrameSkip registers a frame filter that permanently excludes matching frames
// from all future stack traces. Both file and fun are matched using exact equality
// against the captured frame's file path and short function name respectively.
// An empty string matches anything (i.e., acts as a wildcard).
//
// The file path is matched against the cleaned/shortened path after Go module
// prefixes are stripped by removeGoPath, leaving paths like
// "github.com/myorg/myproject/pkg/errutil/helper.go". The function name is the
// short name only (e.g., "Wrap", not "github.com/samber/oops.(OopsErrorBuilder).Wrap").
// To discover the exact values, inspect a stack trace first using
// err.(oops.OopsError).Stacktrace().
//
// Calling FrameSkip with the same (file, fun) pair more than once is a no-op —
// duplicate entries are silently ignored.
//
// This function is intended to be called once at program startup, typically in
// an init() function or main(), not on a per-error basis.
//
// Example:
//
//	oops.FrameSkip("github.com/myproject/pkg/errutil/helper.go", "")  // match by exact file path
//	oops.FrameSkip("", "Wrap")                                         // match by short function name
func FrameSkip(file string, fun string) {
	entry := oopsStacktraceFrame{file: file, function: fun}
	for _, existing := range framesSkip {
		if existing == entry {
			return
		}
	}
	framesSkip = append(framesSkip, entry)
}

// GetPublic returns a message that is safe to show to an end user, or a default generic message.
func GetPublic(err error, defaultPublicMessage string) string {
	var oopsError OopsError

	if errors.As(err, &oopsError) {
		msg := oopsError.Public()
		if len(msg) > 0 {
			return msg
		}
	}

	return defaultPublicMessage
}
