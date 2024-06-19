
# Oops - Error handling with context, assertion, stack trace and source fragments

[![tag](https://img.shields.io/github/tag/samber/oops.svg)](https://github.com/samber/oops/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20.3-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/oops?status.svg)](https://pkg.go.dev/github.com/samber/oops)
![Build Status](https://github.com/samber/oops/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/oops)](https://goreportcard.com/report/github.com/samber/oops)
[![Coverage](https://img.shields.io/codecov/c/github/samber/oops)](https://codecov.io/gh/samber/oops)
[![Contributors](https://img.shields.io/github/contributors/samber/oops)](https://github.com/samber/oops/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/oops)](./LICENSE)

(Yet another) error handling library: `oops.Errorf` is a dead-simple drop-in replacement for built-in `error`, adding contextual information such as stack trace, extra attributes, error code, and bug-fixing hints...

> ⚠️ This is NOT a logging library. `oops` should be used as a complement to your existing logging toolchain (zap, zerolog, logrus, slog, go-sentry...).

🥷 Start hacking `oops` with this [playground](https://go.dev/play/p/-_7EBnceJ_A).

<img align="right" title="Oops gopher logo" alt="logo: thanks Gimp" width="280" src="assets/logo.png">

Jump:

- [🤔 Motivations](#-motivations)
- [🚀 Install](#-install)
- [💡 Quick start](#-quick-start)
- [🧠 Spec](#-spec)
  - [Error constructors](#error-constructors)
  - [Context](#context)
  - [Other helpers](#other-helpers)
  - [Stack trace](#stack-trace)
  - [Source fragments](#source-fragments)
  - [Panic handling](#panic-handling)
  - [Assertions](#assertions)
  - [Output](#output)
- [📫 Loggers](#-loggers)
- [🥷 Tips and best practices](#-tips-and-best-practices)
	
## 🤔 Motivations

Loggers usually allow developers to build records with contextual attributes, that describe errors, such as:
- `zap.Infow("failed to fetch URL", "url", url)`
- `logrus.WithFields("url", url).Error("failed to fetch URL")`).

Go recommends cascading error handling, which can cause the error to be triggered far away from the call to the logger. Returning context over X callers is painful, and to be meaningful, the stack trace must be gathered by the error builder instead of the logger.

This is why we need an `error` wrapper!

🥵 Why develop yet another library?
- drop-in replacement to `error`
- easy to integrate without large refactoring
- separation of concern (logger vs error)
- extra attributes
- developer-friendly error builder
- no extra code for [output](#output): can be used with loggers, printf syntax...
- out-of-the-box stack trace and source fragments
- one-line panic handling
- one-line assertion

### ❌ Before samber/oops

In the following example, we try to propagate an error with contextual information and stack trace, to the caller function `handler()`:

```go
func c(token string) error {
    userID := ...   // <-- How do I transport `userID` and `role` from here...
    role := ...

    // ...

    return fmt.Errorf("an error")
}

func b() error {
    // ...
    return c()
}

func a() {
    err := b()
    if err != nil {
        // print log
        slog.Error(err.Error(),
            slog.String("user.id", "????"),      // <-- ...to here ??
            slog.String("user.role", "????"),    // <-- ...and here ??
            slog.String("stracktrace", generateStacktrace()))  // <-- this won't contain the exact error location 😩
    }
}
```

### ✅ Using samber/oops

I would rather write something like that:

```go
func d() error {
    return oops.
        Code("iam_missing_permission").
        In("authz").
        Tags("authz").
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
    // add more context
    return oops.
        In("iam").
        Tags("iam").
        Trace("e76031ee-a0c4-4a80-88cb-17086fdd19c0").
        With("hello", "world").
        Wrapf(c(), "something failed")
}

func a() error {
	return b()
}

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    err := a()
    if err != nil {
        logger.Error(
            err.Error(),
            slog.Any("error", err), // unwraps and flattens error context
        )
    }
}
```

<div style="text-align:center;">
    <img alt="Why 'oops'?" src="./assets/motivation.png" style="max-width: 650px;">
</div>

### Why "oops"?

Have you already heard a developer yelling at unclear error messages in Sentry, with no context, just before figuring out he wrote this piece of shit by himself?

Yes. Me too.

<div style="text-align:center;">
    <img alt="oops!" src="https://media3.giphy.com/media/v1.Y2lkPTc5MGI3NjExZDU2MjE1ZTk1ZjFmMWNkOGZlY2YyZGYzNjA4ZWIyZWU4NTI3MmE1OCZlcD12MV9pbnRlcm5hbF9naWZzX2dpZklkJmN0PWc/mvyvXwL26FfAtRCLPk/giphy.gif">
</div>

## 🚀 Install

```sh
go get github.com/samber/oops
```

This library is v1 and follows SemVer strictly.

No breaking changes will be made to APIs before v2.0.0.

This library has no dependencies outside the Go standard library.

## 💡 Quick start

This library provides a simple `error` builder for composing structured errors, with contextual attributes and stack trace.

Since `oops.OopsError` implements the `error` interface, you will be able to compose and wrap native errors with `oops.OopsError`.

🥷 Start hacking `oops` with this [playground](https://go.dev/play/p/-_7EBnceJ_A).

## 🧠 Spec

GoDoc: [https://godoc.org/github.com/samber/oops](https://godoc.org/github.com/samber/oops)

### Error constructors

| Constructor                                                             | Description                                                                                           |
| ----------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `.Errorf(format string, args ...any) error`                             | Formats an error and returns `oops.OopsError` object that satisfies `error`                           |
| `.Wrap(err error) error`                                                | Wraps an error into an `oops.OopsError` object that satisfies `error`                                 |
| `.Wrapf(err error, format string, args ...any) error`                   | Wraps an error into an `oops.OopsError` object that satisfies `error` and formats an error message    |
| `.Recover(cb func()) error`                                             | Handle panic and returns `oops.OopsError` object that satisfies `error`.                              |
| `.Recoverf(cb func(), format string, args ...any) error`                | Handle panic and returns `oops.OopsError` object that satisfies `error` and formats an error message. |
| `.Assert(condition bool) OopsErrorBuilder`                              | Panics if condition is false. Assertions can be chained.                                              |
| `.Assertf(condition bool, format string, args ...any) OopsErrorBuilder` | Panics if condition is false and formats an error message. Assertions can be chained.                 |

#### Examples

```go
// with error wrapping
err0 := oops.
    In("repository").
    Tags("database", "sql").
    Wrapf(sql.Exec(query), "could not fetch user")  // Wrapf returns nil when sql.Exec() is nil

// with panic recovery
err1 := oops.
    In("repository").
    Tags("database", "sql").
    Recover(func () {
        panic("caramba!")
    })

// with assertion
err2 := oops.
    In("repository").
    Tags("database", "sql").
    Recover(func () {
        // ...
        oops.Assertf(time.Now().Weekday() == 1, "This code should run on Monday only.")
        // ...
    })
```

### Context

The library provides an error builder. Each method can be used standalone (eg: `oops.With(...)`) or from a previous builder instance (eg: `oops.In("iam").User("user-42")`).

The `oops.OopsError` builder must finish with either `.Errorf(...)`, `.Wrap(...)` or `.Wrapf(...)`.

| Builder method                          | Getter                                  | Description                                                                                                                                                                                |
| --------------------------------------- | --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `.With(string, any)`                    | `err.Context() map[string]any`          | Supply a list of attributes key+value. Values of type `func() any {}` are accepted and evaluated lazily.                                                                                   |
| `.WithContext(context.Context, ...any)` | `err.Context() map[string]any`          | Supply a list of values declared in context. Values of type `func() any {}` are accepted and evaluated lazily.                                                                             |
| `.Code(string)`                         | `err.Code() string`                     | Set a code or slug that describes the error. Error messages are intented to be read by humans, but such code is expected to be read by machines and be transported over different services |
| `.Time(time.Time)`                      | `err.Time() time.Time`                  | Set the error time (default: `time.Now()`)                                                                                                                                                 |
| `.Since(time.Time)`                     | `err.Duration() time.Duration`          | Set the error duration                                                                                                                                                                     |
| `.Duration(time.Duration)`              | `err.Duration() time.Duration`          | Set the error duration                                                                                                                                                                     |
| `.In(string)`                           | `err.Domain() string`                   | Set the feature category or domain                                                                                                                                                         |
| `.Tags(...string)`                      | `err.Tags() []string`                   | Add multiple tags, describing the feature returning an error                                                                                                                               |
| `.Trace(string)`                        | `err.Trace() string`                    | Add a transaction id, trace id, correlation id... (default: ULID)                                                                                                                          |
| `.Span(string)`                         | `err.Span() string`                     | Add a span representing a unit of work or operation... (default: ULID)                                                                                                                     |
| `.Hint(string)`                         | `err.Hint() string`                     | Set a hint for faster debugging                                                                                                                                                            |
| `.Owner(string)`                        | `err.Owner() (string)`                  | Set the name/email of the collegue/team responsible for handling this error. Useful for alerting purpose                                                                                   |
| `.User(string, any...)`                 | `err.User() (string, map[string]any)`   | Supply user id and a chain of key/value                                                                                                                                                    |
| `.Tenant(string, any...)`               | `err.Tenant() (string, map[string]any)` | Supply tenant id and a chain of key/value                                                                                                                                                  |
| `.Request(*http.Request, bool)`         | `err.Request() *http.Request`           | Supply http request                                                                                                                                                                        |
| `.Response(*http.Response, bool)`       | `err.Response() *http.Response`         | Supply http response                                                                                                                                                                       |

#### Examples

```go
// simple error with stacktrace
err1 := oops.Errorf("could not fetch user")

// with optional domain
err2 := oops.
    In("repository").
    Tags("database", "sql").
    Errorf("could not fetch user")

// with custom attributes
ctx := context.WithContext(context.Background(), "a key", "value")
err3 := oops.
    With("driver", "postgresql").
    With("query", query).
    With("query.duration", queryDuration).
    With("lorem", func() string { return "ipsum" }).	// lazy evaluation
    WithContext(ctx, "a key", "another key").
    Errorf("could not fetch user")

// with trace+span
err4 := oops.
    Trace(traceID).
    Span(spanID).
    Errorf("could not fetch user")

// with hint and ownership, for helping developer to solve the issue
err5 := oops.
    Hint("The user could have been removed. Please check deleted_at column.").
    Owner("Slack: #api-gateway").
    Errorf("could not fetch user")

// with optional userID
err6 := oops.
    User(userID).
    Errorf("could not fetch user")

// with optional user data
err7 := oops.
    User(userID, "firstname", "Samuel").
    Errorf("could not fetch user")

// with optional user and tenant
err8 := oops.
    User(userID, "firstname", "Samuel").
    Tenant(workspaceID, "name", "my little project").
    Errorf("could not fetch user")

// with optional http request and response
err9 := oops.
    Request(req, false).
    Response(res, true).
    Errorf("could not fetch user")
```

### Other helpers

- `oops.AsError[MyError](error) (MyError, bool)` as an alias to `errors.As(...)`

### Stack trace

This library provides a pretty printed stack trace for each generated error.

The stack trace max depth can be set using:

```go
// default: 10
oops.StackTraceMaxDepth = 42
```

The stack trace will be printed this way:

```go
err := oops.Errorf("permission denied")

fmt.Println(err.(oops.OopsError).Stacktrace())
```

<div style="text-align:center;">
    <img alt="Stacktrace" src="./assets/stacktrace1.png" style="max-width: 650px;">
</div>

Wrapped errors will be reported as an annotated stack trace:

```go
err1 := oops.Errorf("permission denied")
// ...
err2 := oops.Wrapf(err, "something failed")

fmt.Println(err2.(oops.OopsError).Stacktrace())
```

<div style="text-align:center;">
    <img alt="Stacktrace" src="./assets/stacktrace2.png" style="max-width: 650px;">
</div>

### Source fragments

The exact error location can be provided in a Go file extract.

Source fragments are hidden by default. You must run `oops.SourceFragmentsHidden = false` to enable this feature. Go source files being read at run time, you have to keep the source code at the same location.

In a future release, this library is expected to output a colorized extract. Please contribute!

```go
oops.SourceFragmentsHidden = false

err1 := oops.Errorf("permission denied")
// ...
err2 := oops.Wrapf(err, "something failed")

fmt.Println(err2.(oops.OopsError).Sources())
```

<div style="text-align:center;">
    <img alt="Sources" src="./assets/sources1.png" style="max-width: 650px;">
</div>

### Panic handling

`oops` library is delivered with a try/catch -ish error handler. 2 handlers variants are available: `oops.Recover()` and `oops.Recoverf()`. Both can be used in the `oops` error builder with usual methods.

🥷 Start hacking `oops.Recover()` with this [playground](https://go.dev/play/p/uGwrFj9mII8).

```go
func mayPanic() {
	panic("permission denied")
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
            mayPanic()
            // ...
        }, "unexpected error %d", 42)
}
```

### Assertions

Assertions may be considered an anti-pattern for Golang since we only call `panic()` for unexpected and critical errors. In this situation, assertions might help developers to write safer code.

```go
func mayPanic() {
    x := 42

    oops.
        Trace("6710668a-2b2a-4de6-b8cf-3272a476a1c9").
        Hint("Runbook: https://doc.acme.org/doc/abcd.md").
        Assertf(time.Now().Weekday() == 1, "This code should run on Monday only.").
        With("x", x).
        Assertf(x == 42, "expected x to be equal to 42, but got %d", x)

    oops.Assert(re.Match(email))

    // ...
}

func handlePanic() error {
    return oops.
        Code("iam_authz_missing_permission").
        In("authz").
        Recover(func() {
            // ...
            mayPanic()
            // ...
        })
}
```

### Output

Errors can be printed in many ways. Logger formatters provided in this library use these methods.

#### Errorf `%w`

```go
str := fmt.Errorf("something failed: %w", oops.Errorf("permission denied"))

fmt.Println(err.Error())
// Output:
// something failed: permission denied
```

#### printf `%v`

```go
err := oops.Errorf("permission denied")

fmt.Printf("%v", err)
// Output:
// permission denied
```

#### printf `%+v`

```go
err := oops.Errorf("permission denied")

fmt.Printf("%+v", err)
```

<div style="text-align:center;">
    <img alt="Output" src="./assets/output-printf-plusv.png" style="max-width: 650px;">
</div>

#### JSON Marshal

```go
b := json.MarshalIndent(err, "", "  ")
```

<div style="text-align:center;">
    <img alt="Output" src="./assets/output-json.png" style="max-width: 650px;">
</div>

#### slog.Valuer

```go
err := oops.Errorf("permission denied")

attr := slog.Error(err.Error(),
    slog.Any("error", err))

// Output:
// slog.Group("error", ...)
```

#### Custom timezone

```go
loc, _ := time.LoadLocation("Europe/Paris")
oops.Local = loc
```

## 📫 Loggers

Some loggers may need a custom formatter to extract attributes from `oops.OopsError`.

Available loggers:
- log: [playground](https://go.dev/play/p/uNx3CcT-X40) - [example](https://github.com/samber/oops/tree/master/examples/log)
- slog: [playground](https://go.dev/play/p/-X2ZnqjyDLu) - [example](https://github.com/samber/oops/examples/slog)
- logrus: [formatter](https://github.com/samber/oops/tree/master/loggers/logrus) - [playground](https://go.dev/play/p/-_7EBnceJ_A) - [example](https://github.com/samber/oops/tree/master/examples/logrus)

We are looking for contributions and examples for:
- zap
- zerolog
- go-sentry
- otel
- other?

Examples of formatters can be found in `ToMap()`, `Format()`, `Marshal()` and `LogValuer` methods of `oops.OopsError`.

## 🥷 Tips and best practices

### Wrap/Wrapf shortcut

`oops.Wrap(...)` and `oops.Wrapf(...)` returns nil if the provided `error` is nil.

❌ So don't write:

```go
err := mayFail()
if err != nil {
    return oops.Wrapf(err, ...)
}

return nil
```

✅ but write:

```go
return oops.Wrapf(mayFail(), ...)
```

### Reuse error builder

Writing a full contextualized error can be painful and very repetitive. But a single context can be used for multiple errors in a single function:

❌ So don't write:

```go
err := mayFail1()
if err != nil {
    return oops.
        In("iam").
        Trace("77cb6664").
        With("hello", "world").
        Wrap(err)
}

err = mayFail2()
if err != nil {
    return oops.
        In("iam").
        Trace("77cb6664").
        With("hello", "world").
        Wrap(err)
}

return oops.
    In("iam").
    Trace("77cb6664").
    With("hello", "world").
    Wrap(mayFail3())
```

✅ but write:

```go
errorBuilder := oops.
    In("iam").
    Trace("77cb6664").
    With("hello", "world")

err := mayFail1()
if err != nil {
    return errorBuilder.Wrap(err)
}

err = mayFail2()
if err != nil {
    return errorBuilder.Wrap(err)
}

return errorBuilder.Wrap(mayFail3())
```

### Caller/callee attributes

Also, think about feeding error context in every caller, instead of adding extra information at the last moment.

❌ So don't write:

```go
func a() error {
    return b()
}

func b() error {
    return c()
}

func c() error {
    return d()
}

func d() error {
    return oops.
        Code("iam_missing_permission").
        In("authz").
        Trace("4ea76885-a371-46b0-8ce0-b72b277fa9af").
        Time(time.Now()).
        With("hello", "world").
        With("permission", "post.create").
        Hint("Runbook: https://doc.acme.org/doc/abcd.md").
        User("user-123", "firstname", "john", "lastname", "doe").
        Tenant("organization-123", "name", "Microsoft").
        Errorf("permission denied")
}
```

✅ but write:

```go
func a() error {
	return b()
}

func b() error {
    return oops.
        In("iam").
        Trace("4ea76885-a371-46b0-8ce0-b72b277fa9af").
        With("hello", "world").
        Wrapf(c(), "something failed")
}

func c() error {
    return d()
}

func d() error {
    return oops.
        Code("iam_missing_permission").
        In("authz").
        Time(time.Now()).
        With("permission", "post.create").
        Hint("Runbook: https://doc.acme.org/doc/abcd.md").
        User("user-123", "firstname", "john", "lastname", "doe").
        Tenant("organization-123", "name", "Microsoft").
        Errorf("permission denied")
}
```

## 🤝 Contributing

- Ping me on Twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/oops)
- Fix [open issues](https://github.com/samber/oops/issues) or request new features

Don't hesitate ;)

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/oops)

## 💫 Show your support

Give a ⭐️ if this project helped you!

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2023 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
