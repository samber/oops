package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/samber/oops"
	oopszap "github.com/samber/oops/loggers/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// go run examples/zap/example.go

func d() error {
	req, _ := http.NewRequest("POST", "http://localhost:1337/foobar", strings.NewReader("hello world"))

	return oops.
		Code("iam_authz_missing_permission").
		In("authz").
		Time(time.Now()).
		With("user_id", 1234).
		With("permission", "post.create").
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		User("user-123", "firstname", "john", "lastname", "doe").
		Request(req, true).
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
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, _ := config.Build()
	defer logger.Sync()

	err := a()

	logger.Error(err.Error(),
		zap.Object("error", oopszap.OopsMarshalFunc(err)),
		zap.String("stacktrace", oopszap.OopsStackMarshaller(err)),
	)
}
