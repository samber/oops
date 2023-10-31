package main

import (
	"github.com/samber/oops"
	oopslogrus "github.com/samber/oops/loggers/logrus"
	"github.com/sirupsen/logrus"
)

// go run examples/segfault/example.go 2>&1 | jq
// go run examples/segfault/example.go 2>&1 | jq .stacktrace -r

func nilPointerException() {
	var a *int
	*a = 42
}

func handlePanic() error {
	return oops.
		Code("iam_authz_missing_permission").
		In("authz").
		With("permission", "post.create").
		Trace("6710668a-2b2a-4de6-b8cf-3272a476a1c9").
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		Recoverf(func() {
			// ...
			nilPointerException()
			// ...
		}, "unexpected error")
}

func main() {
	logrus.SetFormatter(oopslogrus.NewOopsFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	}))

	err := handlePanic()
	if err != nil {
		logrus.WithError(err).Error(err)
	}
}
