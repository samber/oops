package oops

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

var SourceFragmentsHidden = true

type OopsError struct {
	err      error
	msg      string
	code     string
	time     time.Time
	duration time.Duration

	// context
	domain        string
	tags          []string
	transactionID string
	context       map[string]any
	hint          string
	owner         string

	// user
	userID   string
	userData map[string]any

	// stacktrace
	stacktrace *oopsStacktrace
}

func (o OopsError) Unwrap() error {
	return o.err
}

func (o OopsError) Error() string {
	if o.err != nil {
		if o.msg == "" {
			return o.err.Error()
		}

		return fmt.Sprintf("%s: %s", o.msg, o.err.Error())
	}

	return o.msg
}

func (o OopsError) Code() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.code
		},
	)
}

func (o OopsError) Time() time.Time {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) time.Time {
			return e.time
		},
	)
}

func (o OopsError) Duration() time.Duration {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) time.Duration {
			return e.duration
		},
	)
}

func (o OopsError) Domain() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.domain
		},
	)
}

func (o OopsError) Tags() []string {
	tags := []string{}

	recursive(o, func(e OopsError) {
		tags = append(tags, e.tags...)
	})

	return lo.Uniq(tags)
}

func (o OopsError) Transaction() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.transactionID
		},
	)
}

func (o OopsError) Context() map[string]any {
	return mergeNestedErrorMap(
		o,
		func(e OopsError) map[string]any {
			return e.context
		},
	)
}

func (o OopsError) Hint() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.hint
		},
	)
}

func (o OopsError) Owner() string {
	return getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.owner
		},
	)
}

func (o OopsError) User() (string, map[string]any) {
	userID := getDeepestErrorAttribute(
		o,
		func(e OopsError) string {
			return e.userID
		},
	)
	userData := mergeNestedErrorMap(
		o,
		func(e OopsError) map[string]any {
			return e.userData
		},
	)

	return userID, userData
}

func (o OopsError) Stacktrace() string {
	blocks := []string{}
	topFrame := ""

	recursive(o, func(e OopsError) {
		if e.stacktrace != nil && len(*e.stacktrace) > 0 {
			msg := coalesceOrEmpty(e.msg, "Error")
			block := fmt.Sprintf("%s\n%s", msg, e.stacktrace.String(topFrame))

			blocks = append([]string{block}, blocks...)

			topFrame = (*e.stacktrace)[0].String()
		}
	})

	if len(blocks) == 0 {
		return ""
	}

	return "Oops: " + strings.Join(blocks, "\nThrown at: ")
}

func (o OopsError) Sources() string {
	blocks := [][]string{}

	recursive(o, func(e OopsError) {
		if e.stacktrace != nil && len(*e.stacktrace) > 0 {
			header, body := e.stacktrace.Source()

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

	return strings.Join(
		lo.Map(blocks, func(items []string, _ int) string {
			return strings.Join(items, "\n")
		}),
		"\n\n",
	)
}

func (o OopsError) LogValuer() slog.Value {
	attrs := []slog.Attr{slog.String("message", o.msg)}

	if err := o.Error(); err != "" {
		attrs = append(attrs, slog.String("err", o.Error()))
	}

	if code := o.Code(); code != "" {
		attrs = append(attrs, slog.String("code", o.Code()))
	}

	if t := o.Time(); t != (time.Time{}) {
		attrs = append(attrs, slog.Time("time", o.Time().UTC()))
	}

	if duration := o.Duration(); duration != 0 {
		attrs = append(attrs, slog.Duration("duration", o.Duration()))
	}

	if domain := o.Domain(); domain != "" {
		attrs = append(attrs, slog.String("domain", o.Domain()))
	}

	if tags := o.Tags(); len(tags) > 0 {
		attrs = append(attrs, slog.Any("tags", o.Tags()))
	}

	if transactionID := o.Transaction(); transactionID != "" {
		attrs = append(attrs, slog.String("transaction", o.Transaction()))
	}

	if hint := o.Hint(); hint != "" {
		attrs = append(attrs, slog.String("hint", o.Hint()))
	}

	if owner := o.Owner(); owner != "" {
		attrs = append(attrs, slog.String("owner", o.Owner()))
	}

	if context := o.Context(); len(context) > 0 {
		attrs = append(attrs,
			slog.Group(
				"context",
				lo.ToAnySlice(
					lo.MapToSlice(o.Context(), func(k string, v any) slog.Attr {
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

	if stacktrace := o.Stacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String("stacktrace", o.Stacktrace()))
	}

	if sources := o.Sources(); sources != "" && !SourceFragmentsHidden {
		attrs = append(attrs, slog.String("sources", o.Sources()))
	}

	return slog.GroupValue(attrs...)
}

func (o OopsError) ToMap() map[string]any {
	payload := map[string]any{}

	if err := o.Error(); err != "" {
		payload["error"] = err
	}

	if code := o.Code(); code != "" {
		payload["code"] = code
	}

	if t := o.Time(); t != (time.Time{}) {
		payload["time"] = t.UTC()
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

	if transactionID := o.Transaction(); transactionID != "" {
		payload["transaction"] = transactionID
	}

	if context := o.Context(); len(context) > 0 {
		payload["context"] = context
	}

	if hint := o.Hint(); hint != "" {
		payload["hint"] = hint
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

	if stacktrace := o.Stacktrace(); stacktrace != "" {
		payload["stacktrace"] = stacktrace
	}

	if sources := o.Sources(); sources != "" && !SourceFragmentsHidden {
		payload["sources"] = sources
	}

	return payload
}

func (o OopsError) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.ToMap())
}

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
		output += fmt.Sprintf("At: %s\n", t.UTC())
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

	if transactionID := o.Transaction(); transactionID != "" {
		output += fmt.Sprintf("Transaction: %s\n", transactionID)
	}

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

	if stacktrace := o.Stacktrace(); stacktrace != "" {
		output += fmt.Sprintf("Stackstrace:\n%s\n", stacktrace)
	}

	if sources := o.Sources(); sources != "" && !SourceFragmentsHidden {
		output += fmt.Sprintf("Sources:\n%s\n", sources)
	}

	return output
}

func (o *OopsError) formatSummary() string {
	return o.Error()
}

func getDeepestErrorAttribute[T comparable](err OopsError, getter func(OopsError) T) T {
	if err.err == nil {
		return getter(err)
	}

	if child, ok := AsOops(err.err); ok {
		return coalesceOrEmpty(getDeepestErrorAttribute(child, getter), getter(err))
	}

	return getter(err)
}

func mergeNestedErrorMap(err OopsError, getter func(OopsError) map[string]any) map[string]any {
	if err.err == nil {
		return getter(err)
	}

	if child, ok := AsOops(err.err); ok {
		return lo.Assign(map[string]any{}, getter(err), mergeNestedErrorMap(child, getter))
	}

	return getter(err)
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
