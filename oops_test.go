//nolint:errcheck,forcetypeassert,bodyclose
package oops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestOopsWrap(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Empty(err.(OopsError).msg)
	is.Equal("assert.AnError general error for testing", err.Error())
	err = newBuilder().Wrap(nil)
	is.NoError(err)
}

func TestOopsWrap_wrapped(t *testing.T) { //nolint:paralleltest
	is := assert.New(t)

	// simulate an OopsError wrapped in a StringError, wrapped in a OopsError
	innerErr := fmt.Errorf("an error: %w", fmt.Errorf("another error: %w", With("user", "foobar").Wrap(context.DeadlineExceeded)))
	err := newBuilder().Wrap(innerErr)
	is.Error(err)
	is.Equal(innerErr, err.(OopsError).err)
	is.Empty(err.(OopsError).msg)
	is.Equal("an error: another error: context deadline exceeded", err.Error())
	is.Equal(map[string]any{"user": "foobar"}, err.(OopsError).Context())
	// simulate long http request
	ctx, cancel := context.WithTimeoutCause(context.TODO(), 1*time.Millisecond, With("hello", "world").Errorf("hello timeout"))
	defer cancel()
	time.Sleep(100 * time.Millisecond)
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://google.com", nil)
	_, err = Wrap2(http.DefaultClient.Do(req))
	is.Error(err)
	if runtime.Version() >= "go1.23" {
		is.Equal("Get \"https://google.com\": hello timeout", err.(OopsError).err.Error())
		is.Equal(map[string]any{"hello": "world"}, err.(OopsError).Context())
	} else {
		is.Equal("Get \"https://google.com\": context deadline exceeded", err.(OopsError).err.Error())
		is.Equal(map[string]any{}, err.(OopsError).Context())
	}
}

func TestOopsWrapf(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Wrapf(assert.AnError, "a message %d", 42)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("a message 42", err.(OopsError).msg)
	is.Equal("a message 42: assert.AnError general error for testing", err.Error())
	err = newBuilder().Wrapf(nil, "a message %d", 42)
	is.NoError(err)
}

func TestOopsFromContext(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	domain := "domain"
	key, val := "foo", "bar"
	builder := newBuilder().In(domain).With(key, val).WithContext(context.Background())
	ctx := WithBuilder(context.Background(), builder)
	err := FromContext(ctx).Errorf("a message %d", 42)
	is.Error(err)
	is.Equal(domain, err.(OopsError).domain)
	is.Equal(val, err.(OopsError).context[key])
	is.NotZero(FromContext(context.Background()).time)
}

func TestOopsNew(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().New("a message")
	is.Error(err)
	is.Equal(errors.New("a message"), err.(OopsError).err)
	is.Empty(err.(OopsError).msg)
	is.Equal("a message", err.Error())
}

func TestOopsErrorf(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Errorf("a message %d", 42)
	is.Error(err)
	is.Equal(fmt.Errorf("a message %d", 42), err.(OopsError).err)
	is.Empty(err.(OopsError).msg)
	is.Equal("a message 42", err.Error())
}

func TestOopsCode(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Code("iam_missing_permission").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("iam_missing_permission", err.(OopsError).code)
}

func TestOopsTime(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	now := time.Now()
	err := newBuilder().Time(now).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(now, err.(OopsError).time)
}

func TestOopsSince(t *testing.T) { //nolint:paralleltest
	is := assert.New(t)

	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	err := newBuilder().Since(start).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.GreaterOrEqual(err.(OopsError).duration.Milliseconds(), int64(10))
}

func TestOopsDuration(t *testing.T) { //nolint:paralleltest
	is := assert.New(t)

	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	err := newBuilder().Duration(time.Since(start)).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.GreaterOrEqual(err.(OopsError).duration.Milliseconds(), int64(10))
}

func TestOopsIn(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().In("authz").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("authz", err.(OopsError).domain)
}

func TestOopsTags(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Tags("iam", "authz", "iam").Join(
		newBuilder().Tags("iam", "internal").Wrap(assert.AnError),
		newBuilder().Tags("not-found").Wrap(assert.AnError))
	join, ok := err.(OopsError).err.(interface{ Unwrap() []error })
	is.True(ok)
	is.Len(join.Unwrap(), 2)
	is.Equal(assert.AnError, join.Unwrap()[0].(OopsError).err)
	is.Equal(assert.AnError, join.Unwrap()[1].(OopsError).err)
	is.Equal([]string{"iam", "authz", "iam"}, err.(OopsError).tags) // not deduplicated
	is.Equal([]string{"iam", "internal"}, join.Unwrap()[0].(OopsError).tags)
	is.Equal([]string{"not-found"}, join.Unwrap()[1].(OopsError).tags)
	is.Equal([]string{"iam", "authz", "internal"}, err.(OopsError).Tags()) // deduplicated and recursive
}

func TestOopsHasTag(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Tags("iam", "authz").Join(
		newBuilder().Tags("internal").Wrap(assert.AnError),
		newBuilder().Tags("not-found").Wrap(assert.AnError))
	is.Error(err)
	is.True(err.(OopsError).HasTag("internal"))
	is.True(err.(OopsError).HasTag("authz"))
	is.False(err.(OopsError).HasTag("not-found")) // Does not go over all joined errors so far
	is.False(err.(OopsError).HasTag("1234"))
	is.False(OopsError{}.HasTag("not-found"))
}

func TestOopsTx(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Trace("1234").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("1234", err.(OopsError).trace)
}

func TestOopsTxSpanFromOtel(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	traceID, terr := trace.TraceIDFromHex("12345678901234567890123456789012")
	is.NoError(terr)
	spanID, serr := trace.SpanIDFromHex("1234567890123456")
	is.NoError(serr)
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))
	err := newBuilder().WithContext(ctx).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("12345678901234567890123456789012", err.(OopsError).trace)
	is.Equal("1234567890123456", err.(OopsError).span)
}

func TestOopsWith(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().With("user_id", 1234).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{"user_id": 1234}, err.(OopsError).context)
	err = newBuilder().With("user_id", 1234, "foo").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{"user_id": 1234}, err.(OopsError).context)
	err = newBuilder().With("user_id", 1234, "foo", "bar").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{"user_id": 1234, "foo": "bar"}, err.(OopsError).context)
}

func TestOopsWithContext(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	type test string
	const fooo test = "fooo"
	ctx := context.WithValue(context.Background(), "foo", "bar") //nolint:staticcheck,revive
	ctx = context.WithValue(ctx, fooo, "baz")
	// string
	err := newBuilder().WithContext(ctx, "foo").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{"foo": "bar"}, err.(OopsError).context)
	// type alias
	err = newBuilder().WithContext(ctx, fooo).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{"fooo": "baz"}, err.(OopsError).context)
	// multiple
	err = newBuilder().WithContext(ctx, "foo", fooo).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{"foo": "bar", "fooo": "baz"}, err.(OopsError).context)
	// not found
	err = newBuilder().WithContext(ctx, "bar").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{"bar": nil}, err.(OopsError).context)
	// none
	err = newBuilder().WithContext(ctx).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal(map[string]any{}, err.(OopsError).context)
}

func TestOopsWithLazyEvaluation(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// lazy evaluation
	err := newBuilder().With("user_id", func() int { return 1234 }, "foo", map[string]any{"bar": func() string { return "baz" }}).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(map[string]any{"user_id": 1234, "foo": map[string]any{"bar": "baz"}}, err.(OopsError).Context())
}

func TestOopsHint(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Hint("Runbook: https://doc.acme.org/doc/abcd.md").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("Runbook: https://doc.acme.org/doc/abcd.md", err.(OopsError).hint)
}

func TestOopsPublic(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Public("a public facing message").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("a public facing message", err.(OopsError).public)
}

func TestOopsOwner(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Owner("iam-team@acme.org").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("iam-team@acme.org", err.(OopsError).owner)
}

func TestOopsUser(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().User("user-123").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("user-123", err.(OopsError).userID)
	is.Equal(map[string]any{}, err.(OopsError).userData)
	err = newBuilder().User("user-123", "firstname", "john").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("user-123", err.(OopsError).userID)
	is.Equal(map[string]any{"firstname": "john"}, err.(OopsError).userData)
	err = newBuilder().User("user-123", "firstname", "john", "lastname").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("user-123", err.(OopsError).userID)
	is.Equal(map[string]any{"firstname": "john"}, err.(OopsError).userData)
	err = newBuilder().User("user-123", "firstname", "john", "lastname", "doe").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("user-123", err.(OopsError).userID)
	is.Equal(map[string]any{"firstname": "john", "lastname": "doe"}, err.(OopsError).userData)
}

func TestOopsTenant(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Tenant("workspace-123").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("workspace-123", err.(OopsError).tenantID)
	is.Equal(map[string]any{}, err.(OopsError).tenantData)
	err = newBuilder().Tenant("workspace-123", "name", "My 'hello world' project").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("workspace-123", err.(OopsError).tenantID)
	is.Equal(map[string]any{"name": "My 'hello world' project"}, err.(OopsError).tenantData)
	err = newBuilder().Tenant("workspace-123", "name", "My 'hello world' project", "date").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("workspace-123", err.(OopsError).tenantID)
	is.Equal(map[string]any{"name": "My 'hello world' project"}, err.(OopsError).tenantData)
	err = newBuilder().Tenant("workspace-123", "name", "My 'hello world' project", "date", "2023-01-01").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("workspace-123", err.(OopsError).tenantID)
	is.Equal(map[string]any{"name": "My 'hello world' project", "date": "2023-01-01"}, err.(OopsError).tenantData)
}

func TestOopsRequest(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	req, _ := http.NewRequest("POST", "http://localhost:1337/foobar", strings.NewReader("hello world"))
	err := newBuilder().Request(req, false).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.NotNil(err.(OopsError).req)
	if err.(OopsError).req != nil {
		is.Equal(req, err.(OopsError).req.A)
		is.False(err.(OopsError).req.B)
	}
	is.NotNil(err.(OopsError).Request())
	if err.(OopsError).Request() != nil {
		is.Equal(req, err.(OopsError).Request())
	}
	err = newBuilder().Request(req, true).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.NotNil(err.(OopsError).req)
	if err.(OopsError).req != nil {
		is.Equal(req, err.(OopsError).req.A)
		is.True(err.(OopsError).req.B)
	}
	is.NotNil(err.(OopsError).Request())
	if err.(OopsError).Request() != nil {
		is.Equal(req, err.(OopsError).Request())
	}
}

func TestOopsMixed(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	now := time.Now()
	req, _ := http.NewRequest("POST", "http://localhost:1337/foobar", strings.NewReader("hello world"))
	err := newBuilder().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Trace("1234").
		With("user_id", 1234).
		WithContext(context.WithValue(context.Background(), "foo", "bar"), "foo"). //nolint:staticcheck,revive
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Public("public facing message").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john", "lastname", "doe").
		Tenant("workspace-123", "name", "little project").
		Request(req, false).
		Wrapf(assert.AnError, "a message %d", 42)
	is.Error(err)
	is.Equal("iam_missing_permission", err.(OopsError).code)
	is.Equal(err.(OopsError).time, now)
	is.Equal(time.Second, err.(OopsError).duration)
	is.Equal("authz", err.(OopsError).domain)
	is.Equal("1234", err.(OopsError).trace)
	is.Equal(map[string]any{"user_id": 1234, "foo": "bar"}, err.(OopsError).context)
	is.Equal("Runbook: https://doc.acme.org/doc/abcd.md", err.(OopsError).hint)
	is.Equal("public facing message", err.(OopsError).public)
	is.Equal("authz-team@acme.org", err.(OopsError).owner)
	is.Equal("user-123", err.(OopsError).userID)
	is.Equal(map[string]any{"firstname": "john", "lastname": "doe"}, err.(OopsError).userData)
	is.Equal("workspace-123", err.(OopsError).tenantID)
	is.Equal(map[string]any{"name": "little project"}, err.(OopsError).tenantData)
	is.Equal(err.(OopsError).req, lo.ToPtr(lo.T2(req, false)))
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal("a message 42", err.(OopsError).msg)
}

func TestOopsMixedWithGetters(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	now := time.Now()
	req1, _ := http.NewRequest("POST", "http://localhost:1337/foo", strings.NewReader("hello world"))
	req2, _ := http.NewRequest("POST", "http://localhost:1337/bar", strings.NewReader("hello world"))
	err := newBuilder().
		Code("iam_authz_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Trace("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/1234.md").
		Public("public facing message").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "bob", "lastname", "martin").
		Tenant("workspace-123", "name", "little project").
		Request(req1, true).
		Wrapf(assert.AnError, "a message %d", 42)
	err = newBuilder().
		Code("iam_unknown_error").
		Time(now.Add(time.Hour)).
		Duration(2*time.Second).
		In("iam").
		Trace("abcd").
		With("workspace_id", 5678).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Public("public facing message").
		Owner("iam-team@acme.org").
		User("user-123", "firstname", "john", "lastname", "doe", "email", "john@doe.org").
		Tenant("workspace-123", "name", "little project", "deleted", false).
		Request(req2, true).
		Wrapf(err, "hello world")
	// current error
	is.Error(err)
	is.Equal("iam_authz_missing_permission", err.(OopsError).Code())
	is.Equal(err.(OopsError).Time(), now)
	is.Equal(time.Second, err.(OopsError).Duration())
	is.Equal("authz", err.(OopsError).Domain())
	is.Equal("1234", err.(OopsError).Trace())
	is.Equal(map[string]any{"user_id": 1234, "workspace_id": 5678}, err.(OopsError).Context())
	is.Equal("Runbook: https://doc.acme.org/doc/1234.md", err.(OopsError).Hint())
	is.Equal("public facing message", err.(OopsError).Public())
	is.Equal("authz-team@acme.org", err.(OopsError).Owner())
	is.Equal(lo.T2(err.(OopsError).User()), lo.T2("user-123", map[string]any{"firstname": "bob", "lastname": "martin", "email": "john@doe.org"}))
	is.Equal(lo.T2(err.(OopsError).Tenant()), lo.T2("workspace-123", map[string]any{"name": "little project", "deleted": false}))
	is.Equal(err.(OopsError).Request(), req1)
	is.Equal("hello world: a message 42: assert.AnError general error for testing", err.(OopsError).Error())
	// first-level error
	is.Error(err)
	is.Equal("iam_unknown_error", err.(OopsError).code)
	is.Equal(err.(OopsError).time, now.Add(time.Hour))
	is.Equal(2*time.Second, err.(OopsError).duration)
	is.Equal("iam", err.(OopsError).domain)
	is.Equal("abcd", err.(OopsError).trace)
	is.Equal(map[string]any{"workspace_id": 5678}, err.(OopsError).context)
	is.Equal("Runbook: https://doc.acme.org/doc/abcd.md", err.(OopsError).hint)
	is.Equal("public facing message", err.(OopsError).public)
	is.Equal("iam-team@acme.org", err.(OopsError).owner)
	is.Equal("user-123", err.(OopsError).userID)
	is.Equal(map[string]any{"email": "john@doe.org", "firstname": "john", "lastname": "doe"}, err.(OopsError).userData)
	is.Equal("workspace-123", err.(OopsError).tenantID)
	is.Equal(map[string]any{"deleted": false, "name": "little project"}, err.(OopsError).tenantData)
	is.Equal(err.(OopsError).req, lo.ToPtr(lo.T2(req2, true)))
	is.Equal("a message 42: assert.AnError general error for testing", err.(OopsError).err.Error())
	is.Equal("hello world", err.(OopsError).msg)
	// deepest error
	is.Equal("iam_authz_missing_permission", err.(OopsError).Unwrap().(OopsError).code)
	is.Equal(err.(OopsError).Unwrap().(OopsError).time, now)
	is.Equal(time.Second, err.(OopsError).Unwrap().(OopsError).duration)
	is.Equal("authz", err.(OopsError).Unwrap().(OopsError).domain)
	is.Equal("1234", err.(OopsError).Unwrap().(OopsError).trace)
	is.Equal(map[string]any{"user_id": 1234}, err.(OopsError).Unwrap().(OopsError).context)
	is.Equal("Runbook: https://doc.acme.org/doc/1234.md", err.(OopsError).Unwrap().(OopsError).hint)
	is.Equal("public facing message", err.(OopsError).Unwrap().(OopsError).public)
	is.Equal("authz-team@acme.org", err.(OopsError).Unwrap().(OopsError).owner)
	is.Equal("user-123", err.(OopsError).Unwrap().(OopsError).userID)
	is.Equal(map[string]any{"firstname": "bob", "lastname": "martin"}, err.(OopsError).Unwrap().(OopsError).userData)
	is.Equal("workspace-123", err.(OopsError).Unwrap().(OopsError).tenantID)
	is.Equal(map[string]any{"name": "little project"}, err.(OopsError).Unwrap().(OopsError).tenantData)
	is.Equal(err.(OopsError).Unwrap().(OopsError).req, lo.ToPtr(lo.T2(req1, true)))
	is.Equal(err.(OopsError).Unwrap().(OopsError).err.Error(), assert.AnError.Error())
	is.Equal("a message 42", err.(OopsError).Unwrap().(OopsError).msg)
}

func TestOopsLogValue(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	now := time.Now()
	req, _ := http.NewRequest("POST", "http://localhost:1337/foobar", strings.NewReader("hello world"))
	err := newBuilder().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Tags("iam", "authz").
		Trace("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Public("public facing message").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john").
		Tenant("workspace-123", "name", "little project").
		Request(req, true).
		Wrapf(assert.AnError, "a message %d", 42)
	is.Error(err)
	got := err.(OopsError).LogValue().Group()
	expectedAttrs := []slog.Attr{
		slog.String("message", "a message 42"),
		slog.String("err", "a message 42: assert.AnError general error for testing"),
		slog.String("code", "iam_missing_permission"),
		slog.Time("time", now.UTC()),
		slog.Duration("duration", time.Second),
		slog.String("domain", "authz"),
		slog.Any("tags", []string{"iam", "authz"}),
		slog.String("trace", "1234"),
		slog.String("hint", "Runbook: https://doc.acme.org/doc/abcd.md"),
		slog.String("public", "public facing message"),
		slog.String("owner", "authz-team@acme.org"),
		slog.Group(
			"context",
			slog.Int("user_id", 1234),
		),
		slog.Group(
			"user",
			slog.String("id", "user-123"),
			slog.String("firstname", "john"),
		),
		slog.Group(
			"tenant",
			slog.String("id", "workspace-123"),
			slog.String("name", "little project"),
		),
		slog.String("request", "POST /foobar HTTP/1.1\r\nHost: localhost:1337\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 11\r\nAccept-Encoding: gzip\r\n\r\nhello world"),
		slog.String("stacktrace", err.(OopsError).Stacktrace()),
	}
	is.Len(got, len(expectedAttrs))
	for i := range got {
		is.Equal(expectedAttrs[i].Key, got[i].Key)
		is.Equal(expectedAttrs[i].Value.Kind(), got[i].Value.Kind())
		is.Equal(expectedAttrs[i].Value.Any(), got[i].Value.Any())
	}
}

func TestOopsFormatSummary(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	now := time.Now()
	req, _ := http.NewRequest("POST", "http://localhost:1337/foobar", strings.NewReader("hello world"))
	err := newBuilder().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Trace("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Public("public facing message").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john", "lastname", "doe").
		Tenant("workspace-123", "name", "little project").
		Request(req, true).
		Wrapf(assert.AnError, "a message %d", 42)
	expected := "a message 42: assert.AnError general error for testing"
	is.Equal(expected, fmt.Sprintf("%v", err.(OopsError)))
}

func TestOopsFormatVerbose(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	now, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-05-02 05:26:48.570837 +0000 UTC")
	req, _ := http.NewRequest("POST", "http://localhost:1337/foobar", strings.NewReader("hello world"))
	err := newBuilder().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Trace("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Public("public facing message").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john").
		Tenant("workspace-123", "name", "little project").
		Request(req, true).
		Wrapf(assert.AnError, "a message %d", 42)
	expected := `Oops: a message 42: assert.AnError general error for testing
Code: iam_missing_permission
Time: 2023-05-02 05:26:48.570837 +0000 UTC
Duration: 1s
Domain: authz
Trace: 1234
Hint: Runbook: https://doc.acme.org/doc/abcd.md
Owner: authz-team@acme.org
Context:
  * user_id: 1234
User:
  * id: user-123
  * firstname: john
Tenant:
  * id: workspace-123
  * name: little project
Request:
  * POST /foobar HTTP/1.1
  * Host: localhost:1337
  * User-Agent: Go-http-client/1.1
  * Content-Length: 11
  * Accept-Encoding: gzip
  * 
  * hello world
`
	got := fmt.Sprintf("%+v", withoutStacktrace(err.(OopsError)))
	got = strings.ReplaceAll(got, "\r", "") // remove \r from request
	is.Equal(expected, got)
}

func TestOopsMarshalJSON(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	now, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-05-02 05:26:48.570837 +0200 UTC")
	req, _ := http.NewRequest("POST", "http://localhost:1337/foobar", strings.NewReader("hello world"))
	err := newBuilder().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Trace("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Public("public facing message").
		User("user-123", "firstname", "john", "lastname", "doe").
		Tenant("workspace-123", "name", "little project").
		Request(req, true).
		Wrapf(assert.AnError, "a message %d", 42)
	expected := `{"code":"iam_missing_permission","context":{"user_id":1234},"domain":"authz","duration":"1s","error":"a message 42: assert.AnError general error for testing","hint":"Runbook: https://doc.acme.org/doc/abcd.md","public":"public facing message","request":"POST /foobar HTTP/1.1\r\nHost: localhost:1337\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 11\r\nAccept-Encoding: gzip\r\n\r\nhello world","tenant":{"id":"workspace-123","name":"little project"},"time":"2023-05-02T05:26:48.570837Z","trace":"1234","user":{"firstname":"john","id":"user-123","lastname":"doe"}}`
	got, err := json.Marshal(withoutStacktrace(err.(OopsError)))
	is.NoError(err)
	is.Equal(expected, string(got))
}

func TestOopsGetPublic(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Public("public message").Wrap(assert.AnError)
	is.Equal("public message", GetPublic(err, "default"))
	err = newBuilder().Wrap(assert.AnError)
	is.Equal("default", GetPublic(err, "default"))
	is.Equal("default", GetPublic(nil, "default"))
}

func TestOopsAssert(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test successful assertion
	err := newBuilder().Assert(true).Wrap(assert.AnError)
	is.Error(err)
	// Test failed assertion
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(OopsError)
			is.True(ok)
			is.Equal("assertion failed", err.Error())
		} else {
			t.Fatal("Expected panic for failed assertion")
		}
	}()
	newBuilder().Assert(false)
}

func TestOopsAssertf(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test successful assertion
	err := newBuilder().Assertf(true, "test %d", 42).Wrap(assert.AnError)
	is.Error(err)
	// Test failed assertion
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(OopsError)
			is.True(ok)
			is.Equal("test 42", err.Error())
		} else {
			t.Fatal("Expected panic for failed assertion")
		}
	}()
	newBuilder().Assertf(false, "test %d", 42)
}

func TestOopsSpan(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Span("test-span").Wrap(assert.AnError)
	is.Error(err)
	is.Equal("test-span", err.(OopsError).span)
	// Test that span is set automatically if not provided
	err = newBuilder().Wrap(assert.AnError)
	is.NotEmpty(err.(OopsError).span)
}

func TestOopsResponse(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	resp := &http.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
	err := newBuilder().Response(resp, true).Wrap(assert.AnError)
	is.Error(err)
	// The actual value is a Tuple2, so check the .A field
	is.Equal(resp, err.(OopsError).res.A)
	// Test with nil response
	err = newBuilder().Response(nil, false).Wrap(assert.AnError)
	is.Error(err)
	is.Nil(err.(OopsError).res.A)
}

func TestOopsStackFrames(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Wrap(assert.AnError)
	is.NotNil(err.(OopsError).stacktrace)
	frames := err.(OopsError).StackFrames()
	if frames != nil {
		is.NotEmpty(frames)
	}
	// Test with nil stacktrace
	err2 := OopsError{err: assert.AnError}
	frames2 := err2.StackFrames()
	is.Nil(frames2)
}

func TestOopsLogValuer(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := newBuilder().Wrap(assert.AnError)
	logValuer := err.(OopsError).LogValuer()
	is.NotNil(logValuer)
	// Test that LogValue returns the same as LogValuer()
	logValue := err.(OopsError).LogValue()
	is.Equal(logValue, logValuer)
}

func TestOopsErrorMethods(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test Span method
	err := newBuilder().Span("test-span").Wrap(assert.AnError)
	is.Equal("test-span", err.(OopsError).Span())
	// Test with empty span
	err2 := newBuilder().Wrap(assert.AnError)
	is.NotEmpty(err2.(OopsError).Span())
	// Test Response method
	resp := &http.Response{StatusCode: 500}
	err3 := newBuilder().Response(resp, false).Wrap(assert.AnError)
	is.Equal(resp, err3.(OopsError).Response())
	// Test with nil response
	err4 := newBuilder().Wrap(assert.AnError)
	is.Nil(err4.(OopsError).Response())
}

func TestOopsRecoverEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test recover with non-error panic
	err := newBuilder().Recover(func() {
		panic("string panic")
	})
	is.Error(err)
	is.Contains(err.Error(), "string panic")
	// Test recover with nil panic
	err = newBuilder().Recover(func() {
		panic(nil)
	})
	is.Error(err)
	is.Contains(err.Error(), "panic")
	// Test recover with struct panic
	type testStruct struct {
		Field string
	}
	err = newBuilder().Recover(func() {
		panic(testStruct{Field: "test"})
	})
	is.Error(err)
	is.Contains(err.Error(), "test")
}

func TestOopsWithContextEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test with nil context
	err := newBuilder().WithContext(nil, "key1", "key2").Errorf("test") //nolint:staticcheck
	is.Error(err)
	// Test with empty keys
	err = newBuilder().WithContext(context.Background()).Errorf("test")
	is.Error(err)
	// Test with odd number of keys
	err = newBuilder().WithContext(context.Background(), "key1").Errorf("test")
	is.Error(err)
	// Test with context values that don't exist
	ctx := context.Background()
	err = newBuilder().WithContext(ctx, "nonexistent", "value").Errorf("test")
	is.Error(err)
	is.Equal(map[string]any{"nonexistent": nil, "value": nil}, err.(OopsError).Context())
}

func TestOopsErrorFormatEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test error with nil underlying error
	err := OopsError{msg: "test message"}
	is.Equal("test message", err.Error())
	// Test error with empty message
	err2 := OopsError{err: assert.AnError}
	is.Equal("assert.AnError general error for testing", err2.Error())
	// Test error with both message and underlying error
	err3 := OopsError{err: assert.AnError, msg: "test message"}
	is.Equal("test message: assert.AnError general error for testing", err3.Error())
}

func TestOopsRequestEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test with nil request
	err := newBuilder().Request(nil, false).Wrap(assert.AnError)
	is.Error(err)
	is.Nil(err.(OopsError).Request())
	// Test request method
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err2 := newBuilder().Request(req, false).Wrap(assert.AnError)
	is.Equal(req, err2.(OopsError).Request())
}

func TestOopsSourcesEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test with nil stacktrace
	err := OopsError{err: assert.AnError}
	sources := err.Sources()
	is.Empty(sources)
	// Test with empty frames
	err2 := OopsError{err: assert.AnError, stacktrace: &oopsStacktrace{frames: []oopsStacktraceFrame{}}}
	sources2 := err2.Sources()
	is.Empty(sources2)
}

func TestOopsFormatVerboseEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test with nil error
	err := OopsError{}
	formatted := err.formatVerbose()
	is.NotEmpty(formatted)
	// Test with minimal error
	err2 := OopsError{err: assert.AnError}
	formatted2 := err2.formatVerbose()
	is.NotEmpty(formatted2)
	is.Contains(formatted2, "assert.AnError general error for testing")
}

func TestOopsRecursiveEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test with nil error - recursive is a void function, so we can't test it directly
	// Instead, test the behavior through other means
	err := OopsError{err: nil}
	is.Empty(err.Context())
	// Test with non-OopsError
	err2 := OopsError{err: assert.AnError}
	is.Empty(err2.Context())
}

func TestOopsGetDeepestErrorAttributeEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test with nil error
	result := getDeepestErrorAttribute(OopsError{err: nil}, func(o OopsError) string {
		return o.Code()
	})
	is.Empty(result)
	// Test with non-OopsError
	result2 := getDeepestErrorAttribute(OopsError{err: assert.AnError}, func(o OopsError) string {
		return o.Code()
	})
	is.Empty(result2)
	// Test with OopsError but no context
	err := OopsError{err: assert.AnError}
	result3 := getDeepestErrorAttribute(err, func(o OopsError) string {
		return o.Code()
	})
	is.Empty(result3)
}

func TestOopsMergeNestedErrorMapEdgeCases(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test with OopsError but no context
	err := OopsError{err: assert.AnError}
	result := mergeNestedErrorMap(err, func(o OopsError) map[string]any {
		return o.Context()
	})
	is.Empty(result)
}

func TestOopsMainFunctions(t *testing.T) { //nolint:paralleltest
	is := assert.New(t)

	// Test New function
	err := New("test error")
	is.Error(err)
	is.Equal("test error", err.(OopsError).err.Error())
	// Test Assert function
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(OopsError)
			is.True(ok)
			is.Equal("assertion failed", err.Error())
		} else {
			t.Fatal("Expected panic for failed assertion")
		}
	}()
	Assert(false)
	// Test Assertf function
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(OopsError)
			is.True(ok)
			is.Equal("test assertion", err.Error())
		} else {
			t.Fatal("Expected panic for failed assertion")
		}
	}()
	Assertf(false, "test assertion")
	// Test Code function
	err2 := Code("test_code").Wrap(assert.AnError)
	is.Error(err2)
	is.Equal("test_code", err2.(OopsError).Code())
	// Test Time function
	now := time.Now()
	err3 := Time(now).Wrap(assert.AnError)
	is.Error(err3)
	is.Equal(now, err3.(OopsError).Time())
	// Test Since function
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	err4 := Since(start).Wrap(assert.AnError)
	is.Error(err4)
	is.Positive(err4.(OopsError).Duration())
	// Test Duration function
	duration := 5 * time.Second
	err5 := Duration(duration).Wrap(assert.AnError)
	is.Error(err5)
	is.Equal(duration, err5.(OopsError).Duration())
	// Test In function
	err6 := In("test_domain").Wrap(assert.AnError)
	is.Error(err6)
	is.Equal("test_domain", err6.(OopsError).Domain())
	// Test Tags function
	err7 := Tags("tag1", "tag2").Wrap(assert.AnError)
	is.Error(err7)
	is.Equal([]string{"tag1", "tag2"}, err7.(OopsError).Tags())
	// Test Trace function
	err8 := Trace("test_trace").Wrap(assert.AnError)
	is.Error(err8)
	is.Equal("test_trace", err8.(OopsError).Trace())
	// Test Span function
	err9 := Span("test_span").Wrap(assert.AnError)
	is.Error(err9)
	is.Equal("test_span", err9.(OopsError).Span())
	// Test WithContext function
	ctx := context.WithValue(context.Background(), "key", "value") //nolint:staticcheck,revive
	err10 := WithContext(ctx, "key", "value").Wrap(assert.AnError)
	is.Error(err10)
	is.Equal("value", err10.(OopsError).Context()["key"])
	// Test Hint function
	err11 := Hint("test hint").Wrap(assert.AnError)
	is.Error(err11)
	is.Equal("test hint", err11.(OopsError).Hint())
	// Test Public function
	err12 := Public("public message").Wrap(assert.AnError)
	is.Error(err12)
	is.Equal("public message", err12.(OopsError).Public())
	// Test Owner function
	err13 := Owner("test owner").Wrap(assert.AnError)
	is.Error(err13)
	is.Equal("test owner", err13.(OopsError).Owner())
	// Test User function
	userData := map[string]any{"name": "test"}
	err14 := User("user123", userData).Wrap(assert.AnError)
	is.Error(err14)
	userID, userDataResult := err14.(OopsError).User()
	is.Equal("user123", userID)
	is.Equal(userData, userDataResult)
	// Test Tenant function
	tenantData := map[string]any{"org": "test"}
	err15 := Tenant("tenant123", tenantData).Wrap(assert.AnError)
	is.Error(err15)
	tenantID, tenantDataResult := err15.(OopsError).Tenant()
	is.Equal("tenant123", tenantID)
	is.Equal(tenantData, tenantDataResult)
	// Test Request function
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err16 := Request(req, false).Wrap(assert.AnError)
	is.Error(err16)
	is.Equal(req, err16.(OopsError).Request())
	// Test Response function
	resp := &http.Response{StatusCode: 404}
	err17 := Response(resp, false).Wrap(assert.AnError)
	is.Error(err17)
	is.Equal(resp, err17.(OopsError).Response())
}
