package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/oops"
	oopszerolog "github.com/samber/oops/loggers/zerolog"
)

// go run examples/zerolog/example.go 2>&1 | jq
// go run examples/zerolog/example.go 2>&1 | jq .stack -r

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

func main() {
	zerolog.ErrorStackMarshaler = oopszerolog.OopsStackMarshaller
	zerolog.ErrorMarshalFunc = oopszerolog.OopsMarshalFunc
	logger := zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Logger()

	err := a()

	logger.Error().Stack().Err(err).Msg(err.Error())
}
