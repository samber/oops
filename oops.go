package oops

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Wrap wraps an error into an `oops.OopsError` object that satisfies `error`
func Wrap(err error) error {
	if err == nil {
		return nil
	}

	return new().Wrap(err)
}

// Wrapf wraps an error into an `oops.OopsError` object that satisfies `error` and formats an error message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return new().Wrapf(err, format, args...)
}

// Errorf formats an error and returns `oops.OopsError` object that satisfies `error`.
func Errorf(format string, args ...any) error {
	return new().Errorf(format, args...)
}

func Join(e ...error) error {
	return new().Join(e...)
}

// Recover handle panic and returns `oops.OopsError` object that satisfies `error`.
func Recover(cb func()) (err error) {
	return new().Recover(cb)
}

// Recoverf handle panic and returns `oops.OopsError` object that satisfies `error` and formats an error message.
func Recoverf(cb func(), msg string, args ...any) (err error) {
	return new().Recoverf(cb, msg, args...)
}

// Assert panics if condition is false. Panic payload will be of type oops.OopsError.
// Assertions can be chained.
func Assert(condition bool) OopsErrorBuilder {
	o := new()
	return o.Assert(condition)
}

// Assertf panics if condition is false. Panic payload will be of type oops.OopsError.
// Assertions can be chained.
func Assertf(condition bool, msg string, args ...any) OopsErrorBuilder {
	o := new()
	return o.Assertf(condition, msg, args...)
}

// Code set a code or slug that describes the error.
// Error messages are intented to be read by humans, but such code is expected to
// be read by machines and even transported over different services.
func Code(code string) OopsErrorBuilder {
	return new().Code(code)
}

// Time set the error time.
// Default: `time.Now()`
func Time(time time.Time) OopsErrorBuilder {
	return new().Time(time)
}

// Since set the error duration.
func Since(time time.Time) OopsErrorBuilder {
	return new().Since(time)
}

// Duration set the error duration.
func Duration(duration time.Duration) OopsErrorBuilder {
	return new().Duration(duration)
}

// In set the feature category or domain.
func In(domain string) OopsErrorBuilder {
	return new().In(domain)
}

// Tags adds multiple tags, describing the feature returning an error.
func Tags(tags ...string) OopsErrorBuilder {
	return new().Tags(tags...)
}

// Trace set a transaction id, trace id or correlation id...
func Trace(trace string) OopsErrorBuilder {
	return new().Trace(trace)
}

// Span represents a unit of work or operation.
func Span(span string) OopsErrorBuilder {
	return new().Span(span)
}

// With supplies a list of attributes declared by pair of key+value.
func With(kv ...any) OopsErrorBuilder {
	return new().With(kv...)
}

// With supplies a list of attributes declared by pair of key+value.
func WithContext(ctx context.Context, keys ...any) OopsErrorBuilder {
	return new().WithContext(ctx, keys...)
}

// Hint set a hint for faster debugging.
func Hint(hint string) OopsErrorBuilder {
	return new().Hint(hint)
}

// Owner set the name/email of the collegue/team responsible for handling this error.
// Useful for alerting purpose.
func Owner(owner string) OopsErrorBuilder {
	return new().Owner(owner)
}

// User supplies user id and a chain of key/value.
func User(userID string, data map[string]any) OopsErrorBuilder {
	return new().User(userID, data)
}

// Tenant supplies tenant id and a chain of key/value.
func Tenant(tenantID string, data map[string]any) OopsErrorBuilder {
	return new().Tenant(tenantID, data)
}

// Request supplies a http.Request.
func Request(req *http.Request, withBody bool) OopsErrorBuilder {
	return new().Request(req, withBody)
}

// Response supplies a http.Response.
func Response(res *http.Response, withBody bool) OopsErrorBuilder {
	return new().Response(res, withBody)
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
