package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/samber/oops"
)

// go run examples/slog/example.go | jq
// go run examples/slog/example.go | jq .error.stacktrace -r

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

type myError1 struct {
	err error
}

func (e myError1) Error() string {
	return fmt.Errorf("fuck %w", e.err).Error()
}

func (c myError1) Unwrap() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

type myError2 struct {
	err error
}

func (e myError2) Error() string {
	return fmt.Errorf("fuck %w", e.err).Error()
}

func (c myError2) Unwrap() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

func main() {
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := a()
	// if err != nil {
	// 	logger.Error(
	// 		err.Error(),
	// 		slog.Any("error", err),
	// 	)
	// }
	err2 := &myError1{err: err}
	err3 := &myError2{err: err2}
	fmt.Println(lo.ErrorsAs[oops.OopsError](err3))
}
