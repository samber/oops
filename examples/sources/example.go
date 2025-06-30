package main

import (
	"fmt"
	"time"

	"github.com/samber/oops"
)

// go run examples/sources/example.go | jq .error.sources -r

func f() error {
	return fmt.Errorf("permission denied")
}

func e() error {
	return oops.
		Code("iam_authz_missing_permission").
		In("authz").
		Time(time.Now()).
		With("user_id", 1234).
		With("permission", "post.create").
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		User("user-123", "firstname", "john", "lastname", "doe").
		Wrap(f())
}

func d() error {
	return oops.Wrap(e())
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

	err := a()
	fmt.Println(err.(oops.OopsError).Sources())
}
