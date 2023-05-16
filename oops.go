package oops

import "time"

// Wrap wraps an error into an `oops.OopsError` object that satisfies `error`
func Wrap(err error) error {
	return new().Wrap(err)
}

// Wrapf wraps an error into an `oops.OopsError` object that satisfies `error` and formats an error message.
func Wrapf(err error, format string, args ...any) error {
	return new().Wrapf(err, format, args...)
}

// Errorf formats an error and returns `oops.OopsError` object that satisfies `error`.
func Errorf(format string, args ...any) error {
	return new().Errorf(format, args...)
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

// Tx set a transaction id, trace id or correlation id...
func Tx(transactionID string) OopsErrorBuilder {
	return new().Tx(transactionID)
}

// With supplies a list of attributes declared by pair of key+value.
func With(kv ...any) OopsErrorBuilder {
	return new().With(kv...)
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
