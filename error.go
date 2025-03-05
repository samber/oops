package oops

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
)

var (
	SourceFragmentsHidden                = true
	DereferencePointers                  = true
	Local                 *time.Location = time.UTC
)

var _ error = (*OopsError)(nil)

type OopsError struct {
	err      error
	msg      string
	code     string
	time     time.Time
	duration time.Duration

	// context
	domain  string
	tags    []string
	context map[string]any

	trace string
	span  string

	hint   string
	public string
	owner  string

	// user
	userID     string
	userData   map[string]any
	tenantID   string
	tenantData map[string]any

	// http
	req *lo.Tuple2[*http.Request, bool]
	res *lo.Tuple2[*http.Response, bool]

	// stacktrace
	stacktrace *oopsStacktrace
}

// Unwrap returns the underlying error.
func (o OopsError) Unwrap() error {
	return o.err
}

func (c OopsError) Is(err error) bool {
	return c.err == err
}

// Error returns the error message, without context.
func (o OopsError) Error() string {
	if o.err != nil {
		if o.msg == "" {
			return o.err.Error()
		}

		return fmt.Sprintf("%s: %s", o.msg, o.err.Error())
	}

	return o.msg
}

// Code returns the error cause. Error code is intented to be used by machines.
func (o OopsError) Code() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.code
		},
	)
}

// Time returns the time when the error occured.
func (o OopsError) Time() time.Time {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) time.Time {
			return e.time
		},
	)
}

// Duration returns the duration of the error.
func (o OopsError) Duration() time.Duration {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) time.Duration {
			return e.duration
		},
	)
}

// Domain returns the domain of the error.
func (o OopsError) Domain() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.domain
		},
	)
}

// Tags returns the tags of the error.
func (o OopsError) Tags() []string {
	tags := []string{}

	recursive(o, func(e OopsError) {
		tags = append(tags, e.tags...)
	})

	return lo.Uniq(tags)
}

// HasTag returns true if the tags of the error contain provided value.
func (o OopsError) HasTag(tag string) bool {
	if lo.Contains(o.tags, tag) {
		return true
	}

	if o.err == nil {
		return false
	}

	if child, ok := AsOops(o.err); ok {
		return child.HasTag(tag)
	}

	return false
}

// Context returns a k/v context of the error.
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

// Trace returns the transaction id, trace id, request id, correlation id, etc.
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

// Span returns the current span instead of the deepest one.
func (o OopsError) Span() string {
	return o.span
}

// Hint returns a hint to the user on how to resolve the error.
func (o OopsError) Hint() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.hint
		},
	)
}

// Public returns a message that is safe to show to an end user.
func (o OopsError) Public() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.public
		},
	)
}

// Owner identify the owner responsible for resolving the error.
func (o OopsError) Owner() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.owner
		},
	)
}

// User returns the user id and user data.
func (o OopsError) User() (string, map[string]any) {
	userID := getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.userID
		},
	)
	userData := lazyMapEvaluation(
		mergeNestedErrorMap(
			o,
			func(e OopsError) map[string]any {
				return e.userData
			},
		),
	)

	return userID, userData
}

// Tenant returns the tenant id and tenant data.
func (o OopsError) Tenant() (string, map[string]any) {
	tenantID := getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.tenantID
		},
	)
	tenantData := lazyMapEvaluation(
		mergeNestedErrorMap(
			o,
			func(e OopsError) map[string]any {
				return e.tenantData
			},
		),
	)

	return tenantID, tenantData
}

// Request returns the http request.
func (o OopsError) Request() *http.Request {
	t := o.request()
	if t != nil {
		return t.A
	}

	return nil
}

func (o OopsError) request() *lo.Tuple2[*http.Request, bool] {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) *lo.Tuple2[*http.Request, bool] {
			return e.req
		},
	)
}

// Response returns the http response.
func (o OopsError) Response() *http.Response {
	t := o.response()
	if t != nil {
		return t.A
	}

	return nil
}

func (o OopsError) response() *lo.Tuple2[*http.Response, bool] {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) *lo.Tuple2[*http.Response, bool] {
			return e.res
		},
	)
}

// Stacktrace returns a pretty printed stacktrace of the error.
func (o OopsError) Stacktrace() string {
	blocks := []string{}
	topFrame := ""

	recursive(o, func(e OopsError) {
		if e.stacktrace != nil && len(e.stacktrace.frames) > 0 {
			err := lo.TernaryF(e.err != nil, func() string { return e.err.Error() }, func() string { return "" })
			msg := coalesceOrEmpty(e.msg, err, "Error")
			block := fmt.Sprintf("%s\n%s", msg, e.stacktrace.String(topFrame))

			blocks = append([]string{block}, blocks...)

			topFrame = e.stacktrace.frames[0].String()
		}
	})

	if len(blocks) == 0 {
		return ""
	}

	return "Oops: " + strings.Join(blocks, "\nThrown: ")
}

// Sources returns the source fragments of the error.
func (o OopsError) Sources() string {
	blocks := [][]string{}

	recursive(o, func(e OopsError) {
		if e.stacktrace != nil && len(e.stacktrace.frames) > 0 {
			header, body := e.stacktrace.Source()

			if e.msg != "" {
				header = fmt.Sprintf("%s\n%s", e.msg, header)
			}

			if header != "" && len(body) > 0 {
				blocks = append(
					[][]string{append([]string{header}, body...)},
					blocks...,
				)
			}
		}
	})

	if len(blocks) == 0 {
		return ""
	}

	return "Oops: " + strings.Join(
		lo.Map(blocks, func(items []string, _ int) string {
			return strings.Join(items, "\n")
		}),
		"\n\nThrown: ",
	)
}

// LogValuer returns a slog.Value for logging.
func (o OopsError) LogValuer() slog.Value {
	attrs := []slog.Attr{slog.String("message", o.msg)}

	if err := o.Error(); err != "" {
		attrs = append(attrs, slog.String("err", err))
	}

	if code := o.Code(); code != "" {
		attrs = append(attrs, slog.String("code", code))
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
					lo.MapToSlice(context, func(k string, v any) slog.Attr {
						return slog.Any(k, v)
					}),
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
				lo.MapToSlice(userData, func(k string, v any) slog.Attr {
					return slog.Any(k, v)
				})...,
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
				lo.MapToSlice(tenantData, func(k string, v any) slog.Attr {
					return slog.Any(k, v)
				})...,
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

// ToMap returns a map representation of the error.
func (o OopsError) ToMap() map[string]any {
	payload := map[string]any{}

	if err := o.Error(); err != "" {
		payload["error"] = err
	}

	if code := o.Code(); code != "" {
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

// MarshalJSON implements json.Marshaler.
func (o OopsError) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.ToMap())
}

// Format implements fmt.Formatter.
// If the format is "%+v", then the details of the error are included.
// Otherwise, using "%v", just the summary is included.
func (o OopsError) Format(s fmt.State, verb rune) {
	if verb == 'v' && s.Flag('+') {
		fmt.Fprint(s, o.formatVerbose())
	} else {
		fmt.Fprint(s, o.formatSummary())
	}
}

func (o *OopsError) formatVerbose() string {
	output := fmt.Sprintf("Oops: %s\n", o.Error())

	if code := o.Code(); code != "" {
		output += fmt.Sprintf("Code: %s\n", code)
	}

	if t := o.Time(); t != (time.Time{}) {
		output += fmt.Sprintf("Time: %s\n", t.In(Local))
	}

	if duration := o.Duration(); duration != 0 {
		output += fmt.Sprintf("Duration: %s\n", duration.String())
	}

	if domain := o.Domain(); domain != "" {
		output += fmt.Sprintf("Domain: %s\n", domain)
	}

	if tags := o.Tags(); len(tags) > 0 {
		output += fmt.Sprintf("Tags: %s\n", strings.Join(tags, ", "))
	}

	if trace := o.Trace(); trace != "" {
		output += fmt.Sprintf("Trace: %s\n", trace)
	}

	// if span := o.Span(); span != "" {
	// 	output += fmt.Sprintf("Span: %s\n", span)
	// }

	if hint := o.Hint(); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}

	if owner := o.Owner(); owner != "" {
		output += fmt.Sprintf("Owner: %s\n", owner)
	}

	if context := o.Context(); len(context) > 0 {
		output += "Context:\n"
		for k, v := range context {
			output += fmt.Sprintf("  * %s: %v\n", k, v)
		}
	}

	if userID, userData := o.User(); userID != "" || len(userData) > 0 {
		output += "User:\n"

		if userID != "" {
			output += fmt.Sprintf("  * id: %s\n", userID)
		}

		for k, v := range userData {
			output += fmt.Sprintf("  * %s: %v\n", k, v)
		}
	}

	if tenantID, tenantData := o.Tenant(); tenantID != "" || len(tenantData) > 0 {
		output += "Tenant:\n"

		if tenantID != "" {
			output += fmt.Sprintf("  * id: %s\n", tenantID)
		}

		for k, v := range tenantData {
			output += fmt.Sprintf("  * %s: %v\n", k, v)
		}
	}

	if req := o.request(); req != nil {
		dump, e := httputil.DumpRequestOut(req.A, req.B)
		if e == nil {
			lines := strings.Split(string(dump), "\n")
			lines = lo.Map(lines, func(line string, _ int) string {
				return "  * " + line
			})
			output += fmt.Sprintf("Request:\n%s\n", strings.Join(lines, "\n"))
		}
	}

	if res := o.response(); res != nil {
		dump, e := httputil.DumpResponse(res.A, res.B)
		if e == nil {
			lines := strings.Split(string(dump), "\n")
			lines = lo.Map(lines, func(line string, _ int) string {
				return "  * " + line
			})
			output += fmt.Sprintf("Response:\n%s\n", strings.Join(lines, "\n"))
		}
	}

	if stacktrace := o.Stacktrace(); stacktrace != "" {
		lines := strings.Split(stacktrace, "\n")
		stacktrace = "  " + strings.Join(lines, "\n  ")
		output += fmt.Sprintf("Stacktrace:\n%s\n", stacktrace)
	}

	if sources := o.Sources(); sources != "" && !SourceFragmentsHidden {
		output += fmt.Sprintf("Sources:\n%s\n", sources)
	}

	return output
}

func (o *OopsError) formatSummary() string {
	return o.Error()
}

func recursive(err OopsError, tap func(OopsError)) {
	tap(err)

	if err.err == nil {
		return
	}

	if child, ok := AsOops(err.err); ok {
		recursive(child, tap)
	}
}
