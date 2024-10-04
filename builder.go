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
 * Builder pattern.
 *
 * oops.Errorf("Could not fetch users: %w", err)
 *
 * oops.
 *	User("steve@apple.com", "firstname", "Samuel").
 *	Tenant("apple", "country", "us").
 *	Errorf("403 not permitted")
 *
 * oops.
 *	Time(requestDate).
 *	Duration(requestDuration).
 *	Tx(traceID).
 *	Errorf("Failed to execute http request")
 *
 * oops.
 *	With("project_id", project.ID, "created_at", project.CreatedAt).
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

func (o OopsErrorBuilder) copy() OopsErrorBuilder {
	return OopsErrorBuilder{
		// err:      err,
		// msg:      o.msg,
		code:     o.code,
		time:     o.time,
		duration: o.duration,

		domain:  o.domain,
		tags:    o.tags,
		context: lo.Assign(map[string]any{}, o.context),

		trace: o.trace,
		span:  o.span,

		hint:   o.hint,
		public: o.public,
		owner:  o.owner,

		userID:     o.userID,
		userData:   lo.Assign(map[string]any{}, o.userData),
		tenantID:   o.tenantID,
		tenantData: lo.Assign(map[string]any{}, o.tenantData),

		req: o.req,
		res: o.res,

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
	if o2.span == "" {
		o2.span = ulid.Make().String()
	}
	o2.stacktrace = newStacktrace(o2.span)
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
	if o2.span == "" {
		o2.span = ulid.Make().String()
	}
	o2.stacktrace = newStacktrace(o2.span)
	return OopsError(o2)
}

// Errorf formats an error and returns `oops.OopsError` object that satisfies `error`.
func (o OopsErrorBuilder) Errorf(format string, args ...any) error {
	o2 := o.copy()
	o2.err = fmt.Errorf(format, args...)
	o2.msg = fmt.Errorf(format, args...).Error()
	if o2.span == "" {
		o2.span = ulid.Make().String()
	}
	o2.stacktrace = newStacktrace(o2.span)
	return OopsError(o2)
}

func (o OopsErrorBuilder) Join(e ...error) error {
	return o.Wrap(errors.Join(e...))
}

// Recover handle panic and returns `oops.OopsError` object that satisfies `error`.
func (o OopsErrorBuilder) Recover(cb func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = o.Wrap(e)
			} else {
				err = o.Wrap(fmt.Errorf("%v", r))
			}
		}
	}()

	cb()
	return
}

// Recoverf handle panic and returns `oops.OopsError` object that satisfies `error` and formats an error message.
func (o OopsErrorBuilder) Recoverf(cb func(), msg string, args ...any) (err error) {
	return o.Wrapf(o.Recover(cb), msg, args...)
}

// Assert panics if condition is false. Panic payload will be of type oops.OopsError.
// Assertions can be chained.
func (o OopsErrorBuilder) Assert(condition bool) OopsErrorBuilder {
	if !condition {
		panic(o.Errorf("assertion failed"))
	}

	return o // no need to copy
}

// Assertf panics if condition is false. Panic payload will be of type oops.OopsError.
// Assertions can be chained.
func (o OopsErrorBuilder) Assertf(condition bool, msg string, args ...any) OopsErrorBuilder {
	if !condition {
		panic(o.Errorf(msg, args...))
	}

	return o // no need to copy
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

// WithContext supplies a list of values declared in context.
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

// Trace set a transaction id, trace id or correlation id...
func (o OopsErrorBuilder) Trace(trace string) OopsErrorBuilder {
	o2 := o.copy()
	o2.trace = trace
	return o2
}

// Span represents a unit of work or operation.
func (o OopsErrorBuilder) Span(span string) OopsErrorBuilder {
	o2 := o.copy()
	o2.span = span
	return o2
}

// Hint set a hint for faster debugging.
func (o OopsErrorBuilder) Hint(hint string) OopsErrorBuilder {
	o2 := o.copy()
	o2.hint = hint
	return o2
}

// Public represents a message that is safe to be shown to an end-user.
func (o OopsErrorBuilder) Public(public string) OopsErrorBuilder {
	o2 := o.copy()
	o2.public = public
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

// Tenant supplies tenant id and a chain of key/value.
func (o OopsErrorBuilder) Tenant(tenantID string, tenantData ...any) OopsErrorBuilder {
	o2 := o.copy()
	o2.tenantID = tenantID

	for i := 0; i < len(tenantData)-1; i += 2 {
		k := tenantData[i]
		v := tenantData[i+1]

		if key, ok := k.(string); ok {
			o2.tenantData[key] = v
		}
	}

	return o2
}

// Request supplies a http.Request.
func (o OopsErrorBuilder) Request(req *http.Request, withBody bool) OopsErrorBuilder {
	o2 := o.copy()
	o2.req = lo.ToPtr(lo.T2(req, withBody))
	return o2
}

// Response supplies a http.Response.
func (o OopsErrorBuilder) Response(res *http.Response, withBody bool) OopsErrorBuilder {
	o2 := o.copy()
	o2.res = lo.ToPtr(lo.T2(res, withBody))
	return o2
}
