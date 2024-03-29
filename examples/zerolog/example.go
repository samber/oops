package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/oops"
)

func d() error {
	return oops.
		Code("iam_authz_missing_permission").
		In("authz").
		Time(time.Now()).
		With("user_id", 1234).
		With("permission", "post.create").
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		User("user-123", "firstname", "john", "lastname", "doe").
		Errorf("permission denied")
}

func c() error {
	return d()
}

func b() error {
	return oops.
		In("iam").
		Trace("6710668a-2b2a-4de6-b8cf-3272a476a1c9").
		With("hello", "world").
		Wrapf(c(), "something failed")
}

func a() error {
	return b()
}

func WithOops(l zerolog.Logger, err error) *zerolog.Logger {
	if oopsErr, ok := oops.AsOops(err); ok {
		ctx := l.With()
		for k, v := range oopsErr.ToMap() {
			ctx = ctx.Interface(k, v)
		}

		logger := ctx.Err(oopsErr).Logger()
		return &logger
	}

	// Recursively call into ourself so we can at least get a stack trace for any error
	return WithOops(l, oops.Wrap(err))
}

func main() {
	logger := zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Logger()

	err := a()

	WithOops(logger, err).Error().Send()
}
