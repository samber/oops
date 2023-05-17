package main

import (
	"os"
	"time"

	"github.com/samber/oops"
	"golang.org/x/exp/slog"
)

// go run examples/sources/example.go | jq .error.sources -r

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
	oops.SourceFragmentsHidden = false

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := a()
	if err != nil {
		logger.Error(
			err.Error(),
			slog.Any("error", err),
		)
	}
}
