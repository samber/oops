package oops

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestOopsWrap(t *testing.T) {
	is := assert.New(t)

	err := new().Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).msg, "")

	err = new().Wrap(nil)
	is.Nil(err)
}

func TestOopsWrapf(t *testing.T) {
	is := assert.New(t)

	err := new().Wrapf(assert.AnError, "a message %d", 42)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).msg, "a message 42")

	err = new().Wrapf(nil, "a message %d", 42)
	is.Nil(err)
}

func TestOopsErrorf(t *testing.T) {
	is := assert.New(t)

	err := new().Errorf("a message %d", 42)
	is.Error(err)
	is.Equal(err.(OopsError).err, nil)
	is.Equal(err.(OopsError).msg, "a message 42")
}

func TestOopsCode(t *testing.T) {
	is := assert.New(t)

	err := new().Code("iam_missing_permission").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).code, "iam_missing_permission")
}

func TestOopsTime(t *testing.T) {
	is := assert.New(t)

	now := time.Now()

	err := new().Time(now).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).time, now)
}

func TestOopsSince(t *testing.T) {
	is := assert.New(t)

	start := time.Now()
	time.Sleep(10 * time.Millisecond)

	err := new().Since(start).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.True(err.(OopsError).duration.Milliseconds() >= 10)
}

func TestOopsDuration(t *testing.T) {
	is := assert.New(t)

	start := time.Now()
	time.Sleep(10 * time.Millisecond)

	err := new().Duration(time.Since(start)).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.True(err.(OopsError).duration.Milliseconds() >= 10)
}

func TestOopsIn(t *testing.T) {
	is := assert.New(t)

	err := new().In("authz").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).domain, "authz")
}

func TestOopsTags(t *testing.T) {
	is := assert.New(t)

	err := new().Tags("iam", "authz", "iam").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).tags, []string{"iam", "authz", "iam"}) // not deduplicated
	is.Equal(err.(OopsError).Tags(), []string{"iam", "authz"})      // deduplicated
}

func TestOopsTx(t *testing.T) {
	is := assert.New(t)

	err := new().Tx("1234").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).transactionID, "1234")
}

func TestOopsWith(t *testing.T) {
	is := assert.New(t)

	err := new().With("user_id", 1234).Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).context, map[string]any{"user_id": 1234})

	err = new().With("user_id", 1234, "foo").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).context, map[string]any{"user_id": 1234})

	err = new().With("user_id", 1234, "foo", "bar").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).context, map[string]any{"user_id": 1234, "foo": "bar"})
}

func TestOopsHint(t *testing.T) {
	is := assert.New(t)

	err := new().Hint("Runbook: https://doc.acme.org/doc/abcd.md").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).hint, "Runbook: https://doc.acme.org/doc/abcd.md")
}

func TestOopsOwner(t *testing.T) {
	is := assert.New(t)

	err := new().Owner("iam-team@acme.org").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).owner, "iam-team@acme.org")
}

func TestOopsUser(t *testing.T) {
	is := assert.New(t)

	err := new().User("user-123").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).userID, "user-123")
	is.Equal(err.(OopsError).userData, map[string]any{})

	err = new().User("user-123", "firstname", "john").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).userID, "user-123")
	is.Equal(err.(OopsError).userData, map[string]any{"firstname": "john"})

	err = new().User("user-123", "firstname", "john", "lastname").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).userID, "user-123")
	is.Equal(err.(OopsError).userData, map[string]any{"firstname": "john"})

	err = new().User("user-123", "firstname", "john", "lastname", "doe").Wrap(assert.AnError)
	is.Error(err)
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).userID, "user-123")
	is.Equal(err.(OopsError).userData, map[string]any{"firstname": "john", "lastname": "doe"})
}

func TestOopsMixed(t *testing.T) {
	is := assert.New(t)

	now := time.Now()

	err := new().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Tx("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john", "lastname", "doe").
		Wrapf(assert.AnError, "a message %d", 42)
	is.Error(err)
	is.Equal(err.(OopsError).code, "iam_missing_permission")
	is.Equal(err.(OopsError).time, now)
	is.Equal(err.(OopsError).duration, time.Second)
	is.Equal(err.(OopsError).domain, "authz")
	is.Equal(err.(OopsError).transactionID, "1234")
	is.Equal(err.(OopsError).context, map[string]any{"user_id": 1234})
	is.Equal(err.(OopsError).hint, "Runbook: https://doc.acme.org/doc/abcd.md")
	is.Equal(err.(OopsError).owner, "authz-team@acme.org")
	is.Equal(err.(OopsError).userID, "user-123")
	is.Equal(err.(OopsError).userData, map[string]any{"firstname": "john", "lastname": "doe"})
	is.Equal(err.(OopsError).err, assert.AnError)
	is.Equal(err.(OopsError).msg, "a message 42")
}

func TestOopsMixedWithGetters(t *testing.T) {
	is := assert.New(t)

	now := time.Now()

	err := new().
		Code("iam_authz_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Tx("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/1234.md").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "bob", "lastname", "martin").
		Wrapf(assert.AnError, "a message %d", 42)

	err = new().
		Code("iam_unknown_error").
		Time(now.Add(time.Hour)).
		Duration(2*time.Second).
		In("iam").
		Tx("abcd").
		With("workspace_id", 5678).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Owner("iam-team@acme.org").
		User("user-123", "firstname", "john", "lastname", "doe", "email", "john@doe.org").
		Wrapf(err, "hello world")

	// current error
	is.Error(err)
	is.Equal(err.(OopsError).Code(), "iam_authz_missing_permission")
	is.Equal(err.(OopsError).Time(), now)
	is.Equal(err.(OopsError).Duration(), time.Second)
	is.Equal(err.(OopsError).Domain(), "authz")
	is.Equal(err.(OopsError).Transaction(), "1234")
	is.Equal(err.(OopsError).Context(), map[string]any{"user_id": 1234, "workspace_id": 5678})
	is.Equal(err.(OopsError).Hint(), "Runbook: https://doc.acme.org/doc/1234.md")
	is.Equal(err.(OopsError).Owner(), "authz-team@acme.org")
	is.Equal(lo.T2(err.(OopsError).User()), lo.T2("user-123", map[string]any{"firstname": "bob", "lastname": "martin", "email": "john@doe.org"}))
	is.Equal(err.(OopsError).Error(), "hello world: a message 42: assert.AnError general error for testing")

	// first-level error
	is.Error(err)
	is.Equal(err.(OopsError).code, "iam_unknown_error")
	is.Equal(err.(OopsError).time, now.Add(time.Hour))
	is.Equal(err.(OopsError).duration, 2*time.Second)
	is.Equal(err.(OopsError).domain, "iam")
	is.Equal(err.(OopsError).transactionID, "abcd")
	is.Equal(err.(OopsError).context, map[string]any{"workspace_id": 5678})
	is.Equal(err.(OopsError).hint, "Runbook: https://doc.acme.org/doc/abcd.md")
	is.Equal(err.(OopsError).owner, "iam-team@acme.org")
	is.Equal(err.(OopsError).userID, "user-123")
	is.Equal(err.(OopsError).userData, map[string]any{"email": "john@doe.org", "firstname": "john", "lastname": "doe"})
	is.Equal(err.(OopsError).err.Error(), "a message 42: assert.AnError general error for testing")
	is.Equal(err.(OopsError).msg, "hello world")

	// deepest error
	is.Equal(err.(OopsError).Unwrap().(OopsError).code, "iam_authz_missing_permission")
	is.Equal(err.(OopsError).Unwrap().(OopsError).time, now)
	is.Equal(err.(OopsError).Unwrap().(OopsError).duration, time.Second)
	is.Equal(err.(OopsError).Unwrap().(OopsError).domain, "authz")
	is.Equal(err.(OopsError).Unwrap().(OopsError).transactionID, "1234")
	is.Equal(err.(OopsError).Unwrap().(OopsError).context, map[string]any{"user_id": 1234})
	is.Equal(err.(OopsError).Unwrap().(OopsError).hint, "Runbook: https://doc.acme.org/doc/1234.md")
	is.Equal(err.(OopsError).Unwrap().(OopsError).owner, "authz-team@acme.org")
	is.Equal(err.(OopsError).Unwrap().(OopsError).userID, "user-123")
	is.Equal(err.(OopsError).Unwrap().(OopsError).userData, map[string]any{"firstname": "bob", "lastname": "martin"})
	is.Equal(err.(OopsError).Unwrap().(OopsError).err.Error(), assert.AnError.Error())
	is.Equal(err.(OopsError).Unwrap().(OopsError).msg, "a message 42")
}

func TestOopsLogValuer(t *testing.T) {
	is := assert.New(t)

	now := time.Now()

	err := new().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Tags("iam", "authz").
		Tx("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john").
		Wrapf(assert.AnError, "a message %d", 42)

	is.Error(err)

	got := err.(OopsError).LogValuer().Group()
	expectedAttrs := []slog.Attr{
		slog.String("message", "a message 42"),
		slog.String("err", "a message 42: assert.AnError general error for testing"),
		slog.String("code", "iam_missing_permission"),
		slog.Time("time", now.UTC()),
		slog.Duration("duration", time.Second),
		slog.String("domain", "authz"),
		slog.Any("tags", []string{"iam", "authz"}),
		slog.String("transaction", "1234"),
		slog.String("hint", "Runbook: https://doc.acme.org/doc/abcd.md"),
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
		slog.String("stacktrace", err.(OopsError).Stacktrace()),
	}

	is.Len(got, len(expectedAttrs))
	for i := range got {
		is.Equal(expectedAttrs[i].Key, got[i].Key)
		is.Equal(expectedAttrs[i].Value.Kind(), got[i].Value.Kind())
		is.EqualValues(expectedAttrs[i].Value.Any(), got[i].Value.Any())
	}
}

func TestOopsFormatSummary(t *testing.T) {
	is := assert.New(t)

	now := time.Now()

	err := new().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Tx("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john", "lastname", "doe").
		Wrapf(assert.AnError, "a message %d", 42)

	expected := "a message 42: assert.AnError general error for testing"
	is.Equal(expected, fmt.Sprintf("%v", err.(OopsError)))
}

func TestOopsFormatVerbose(t *testing.T) {
	is := assert.New(t)

	now, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-05-02 05:26:48.570837 +0000 UTC")

	err := new().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Tx("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Owner("authz-team@acme.org").
		User("user-123", "firstname", "john").
		Wrapf(assert.AnError, "a message %d", 42)

	expected := `Oops: a message 42: assert.AnError general error for testing
Code: iam_missing_permission
At: 2023-05-02 05:26:48.570837 +0000 UTC
Duration: 1s
Domain: authz
Transaction: 1234
Hint: Runbook: https://doc.acme.org/doc/abcd.md
Owner: authz-team@acme.org
Context:
  * user_id: 1234
User:
  * id: user-123
  * firstname: john
`

	is.Equal(expected, fmt.Sprintf("%+v", withoutStacktrace(err.(OopsError))))
}

func TestOopsMarshalJSON(t *testing.T) {
	is := assert.New(t)

	now, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-05-02 05:26:48.570837 +0200 UTC")

	err := new().
		Code("iam_missing_permission").
		Time(now).
		Duration(time.Second).
		In("authz").
		Tx("1234").
		With("user_id", 1234).
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		User("user-123", "firstname", "john", "lastname", "doe").
		Wrapf(assert.AnError, "a message %d", 42)

	expected := `{"code":"iam_missing_permission","context":{"user_id":1234},"domain":"authz","duration":"1s","error":"a message 42: assert.AnError general error for testing","hint":"Runbook: https://doc.acme.org/doc/abcd.md","time":"2023-05-02T05:26:48.570837Z","transaction":"1234","user":{"firstname":"john","id":"user-123","lastname":"doe"}}`

	got, err := json.Marshal(withoutStacktrace(err.(OopsError)))
	is.NoError(err)
	is.Equal(expected, string(got))
}
