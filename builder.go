package oops

import (
	"fmt"
	"time"

	"github.com/samber/lo"
)

/**
 * Builder pattern.
 *
 * oops.Errorf("Could not fetch users: %w", err)
 *
 * oops.
 *	User("samuel@screeb.app", "firstname", "Samuel").
 *	Errorf("403 not permitted")
 *
 * oops.
 *	Time(requestDate).
 *	Duration(requestDuration).
 *	Tx(traceID).
 *	Errorf("Failed to execute http request")
 *
 * oops.
 *	With("tenant_id", tenant.ID, "created_at", tenant.CreatedAt).
 *	Errorf("Could not update settings")
 *
 */
type OopsErrorBuilder OopsError

func new() OopsErrorBuilder {
	return OopsErrorBuilder{
		err:      nil,
		msg:      "",
		code:     "",
		time:     time.Now(),
		duration: 0,

		// context
		domain:        "",
		tags:          []string{},
		transactionID: "",
		context:       map[string]any{},
		hint:          "",
		owner:         "",

		// user
		userID:   "",
		userData: map[string]any{},

		// stacktrace
		stacktrace: nil,
	}
}

func (o OopsErrorBuilder) copy() OopsErrorBuilder {
	return OopsErrorBuilder{
		// err:      err,
		// msg:      o.msg,
		code:     o.code,
		time:     o.time,
		duration: o.duration,

		domain:        o.domain,
		tags:          o.tags,
		transactionID: o.transactionID,
		context:       lo.Assign(map[string]any{}, o.context),
		hint:          o.hint,
		owner:         o.owner,

		userID:   o.userID,
		userData: lo.Assign(map[string]any{}, o.userData),

		// stacktrace: o.stacktrace,
	}
}

// Wrap wraps an error into an `oops.OopsError` object that satisfies `error`
func (o OopsErrorBuilder) Wrap(err error) error {
	if err == nil {
		return nil
	}

	o2 := o.copy()
	o2.err = err
	o2.stacktrace = newStacktrace()
	return OopsError(o2)
}

// Wrapf wraps an error into an `oops.OopsError` object that satisfies `error` and formats an error message.
func (o OopsErrorBuilder) Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	o2 := o.copy()
	o2.err = err
	o2.msg = fmt.Errorf(format, args...).Error()
	o2.stacktrace = newStacktrace()
	return OopsError(o2)
}

// Errorf formats an error and returns `oops.OopsError` object that satisfies `error`.
func (o OopsErrorBuilder) Errorf(format string, args ...any) error {
	o2 := o.copy()
	o2.msg = fmt.Errorf(format, args...).Error()
	o2.stacktrace = newStacktrace()
	return OopsError(o2)
}

// Code set a code or slug that describes the error.
// Error messages are intented to be read by humans, but such code is expected to
// be read by machines and even transported over different services.
func (o OopsErrorBuilder) Code(code string) OopsErrorBuilder {
	o2 := o.copy()
	o2.code = code
	return o2
}

// Time set the error time.
// Default: `time.Now()`
func (o OopsErrorBuilder) Time(time time.Time) OopsErrorBuilder {
	o2 := o.copy()
	o2.time = time
	return o2
}

// Since set the error duration.
func (o OopsErrorBuilder) Since(t time.Time) OopsErrorBuilder {
	o2 := o.copy()
	o2.duration = time.Since(t)
	return o2
}

// Duration set the error duration.
func (o OopsErrorBuilder) Duration(duration time.Duration) OopsErrorBuilder {
	o2 := o.copy()
	o2.duration = duration
	return o2
}

// In set the feature category or domain.
func (o OopsErrorBuilder) In(domain string) OopsErrorBuilder {
	o2 := o.copy()
	o2.domain = domain
	return o2
}

// Tags adds multiple tags, describing the feature returning an error.
func (o OopsErrorBuilder) Tags(tags ...string) OopsErrorBuilder {
	o2 := o.copy()
	o2.tags = append(o2.tags, tags...)
	return o2
}

// Tx set a transaction id, trace id or correlation id...
func (o OopsErrorBuilder) Tx(transactionID string) OopsErrorBuilder {
	o2 := o.copy()
	o2.transactionID = transactionID
	return o2
}

// With supplies a list of attributes declared by pair of key+value.
func (o OopsErrorBuilder) With(kv ...any) OopsErrorBuilder {
	o2 := o.copy()
	for i := 0; i < len(kv)-1; i += 2 {
		k := kv[i]
		v := kv[i+1]

		if key, ok := k.(string); ok {
			o2.context[key] = v
		}
	}

	return o2
}

// Hint set a hint for faster debugging.
func (o OopsErrorBuilder) Hint(hint string) OopsErrorBuilder {
	o2 := o.copy()
	o2.hint = hint
	return o2
}

// Owner set the name/email of the collegue/team responsible for handling this error.
// Useful for alerting purpose.
func (o OopsErrorBuilder) Owner(owner string) OopsErrorBuilder {
	o2 := o.copy()
	o2.owner = owner
	return o2
}

// User supplies user id and a chain of key/value.
func (o OopsErrorBuilder) User(userID string, userData ...any) OopsErrorBuilder {
	o2 := o.copy()
	o2.userID = userID

	for i := 0; i < len(userData)-1; i += 2 {
		k := userData[i]
		v := userData[i+1]

		if key, ok := k.(string); ok {
			o2.userData[key] = v
		}
	}

	return o2
}
