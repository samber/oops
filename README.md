
# Oops - Error handling with context, stack trace and source fragments

[![tag](https://img.shields.io/github/tag/samber/oops.svg)](https://github.com/samber/oops/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20.3-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/oops?status.svg)](https://pkg.go.dev/github.com/samber/oops)
![Build Status](https://github.com/samber/oops/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/oops)](https://goreportcard.com/report/github.com/samber/oops)
[![Coverage](https://img.shields.io/codecov/c/github/samber/oops)](https://codecov.io/gh/samber/oops)
[![Contributors](https://img.shields.io/github/contributors/samber/oops)](https://github.com/samber/oops/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/oops)](./LICENSE)

(Yet another) error handling library: `oops.OopsError` is a dead-simple drop-in replacement for built-in `error`, adding contextual information such as stack trace, extra attributes, trigger time, error code, and bug-fixing hints...

> ⚠️ This is NOT a logging library. `oops` should be used as a complement to your existing logging toolchain (zap, zerolog, logrus, slog, go-sentry...).

<img align="right" title="Oops gopher logo" alt="logo: thanks Gimp" width="280" src="assets/logo.png">

Jump:

- [🤔 Motivations](#🤔-motivations)
- [🚀 Install](#🚀-install)
- [💡 Quick start](#💡-quick-start)
- [🧠 Spec](#🧠-spec)
  - [Error constructors](#error-constructors)
  - [Context](#context)
  - [Other helpers](#other-helpers)
  - [Stack trace](#stack-trace)
  - [Source fragments](#source-fragments)
  - [Output](#output)
- [🥷 Tips and best practices](#🥷-tips-and-best-practices)
- [📫 Loggers](#📫-loggers)
	
## 🤔 Motivations

Loggers usually allow developers to build records with contextual attributes, that describe errors (such as `zap.Infow("failed to fetch URL", "url", url)` or `logrus.WithFields("url", url).Error("failed to fetch URL")`). But Go recommends cascading error handling, so the error may be written very far from the call to the logger.

Also, the stack trace should be gathered at the `fmt.Errorf` call, instead of `logger.Error()`.

So this is why I consider the error context and stack trace need to be transported in an `error` wrapper!

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
		Tx("e76031ee-a0c4-4a80-88cb-17086fdd19c0").
		With("hello", "world").
		Wrapf(c(), "something failed")
}

func a() error {
	return b()
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := a()

	logger.Error(
		err.Error(),
		slog.Any("error", err), // unwraps and flattens error context
	)
}
```

Output:

```json
{
  "time": "2023-05-02 05:26:48.570837Z",
  "level": "ERROR",
  "msg": "something failed: permission denied",
  "error": {
    "code": "iam_missing_permission",
    "context": {
      "hello": "world",
      "permission": "post.create",
      "user_id": 1234
    },
    "domain": "authz",
    "tags": ["iam", "authz"],
    "error": "something failed: permission denied",
    "hint": "Runbook: https://doc.acme.org/doc/abcd.md",
    "stacktrace":
        "Oops: permission denied
          --- at github.com/samber/oops/loggers/slog/example.go:20 (d)
          --- at github.com/samber/oops/loggers/slog/example.go:24 (c)
        Thrown: something failed
          --- at github.com/samber/oops/loggers/slog/example.go:32 (b)
          --- at github.com/samber/oops/loggers/slog/example.go:36 (a)
          --- at github.com/samber/oops/loggers/slog/example.go:42 (main)",
    "time": "2023-05-02 05:26:48.570837Z",
    "transaction": "e76031ee-a0c4-4a80-88cb-17086fdd19c0",
    "user": {
      "firstname": "john",
      "id": "user-123",
      "lastname": "doe"
    }
  }
}
```

### Why "oops"?

Have you already heard a developer yelling for receiving badly written error messages in Sentry, with no context, just before figuring out he wrote this piece of shit by himself?

Yes. Me too.

![oops!](https://media3.giphy.com/media/v1.Y2lkPTc5MGI3NjExZDU2MjE1ZTk1ZjFmMWNkOGZlY2YyZGYzNjA4ZWIyZWU4NTI3MmE1OCZlcD12MV9pbnRlcm5hbF9naWZzX2dpZklkJmN0PWc/mvyvXwL26FfAtRCLPk/giphy.gif)

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

```go
// simple error with stacktrace
err1 := oops.Errorf("could not fetch user")

// with optional domain
err2 := oops.
    In("repository").
    Tags("database", "sql").
    Errorf("could not fetch user")

// with custom attributes
err3 := oops.
    With("driver", "postgresql").
    With("query", query).
    With("query.duration", queryDuration).
    Errorf("could not fetch user")

// with transaction
err4 := oops.
    Tx(traceID).
    Errorf("could not fetch user")

// with hint and ownership, for helping developer to solve the issue
err5 := oops.
    Hint("The user could have been removed. Please check deleted_at column.").
    Owner("api-team@acme.org").
    Errorf("could not fetch user")

// with optional userID
err6 := oops.
    By(userID).
    Errorf("could not fetch user")

// with optional user data
err7 := oops.
    By(userID, "firstname", "Samuel").
    Errorf("could not fetch user")

// with error wrapping
err8 := oops.
    In("repository").
    Tags("database", "sql").
    By(userID).
    Time(queryTime).
    With("driver", "postgresql").
    With("query", query).
    With("query.duration", time.Since(queryTime)).
    Wrapf(sql.Exec(query), "could not fetch user")  // Wrapf returns nil when sql.Exec() is nil
```

## 🧠 Spec

GoDoc: [https://godoc.org/github.com/samber/oops](https://godoc.org/github.com/samber/oops)

### Error constructors

| Builder method                                        | Description                                                                                        |
| ----------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `.Errorf(format string, args ...any) error`           | Formats an error and returns `oops.OopsError` object that satisfies `error`                        |
| `.Wrap(err error) error`                              | Wraps an error into an `oops.OopsError` object that satisfies `error`                              |
| `.Wrapf(err error, format string, args ...any) error` | Wraps an error into an `oops.OopsError` object that satisfies `error` and formats an error message |

### Context

The library provides an error builder. Each method can be used standalone (eg: `oops.With(...)`) or from a previous builder instance (eg: `oops.In("iam").User("user-42")`).

The `oops.OopsError` builder must finish with either `.Errorf(...)`, `.Wrap(...)` or `.Wrapf(...)`.

| Builder method             | Getter                                | Description                                                                                                                                                                                |
| -------------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `.With(string, any)`       | `err.Context() map[string]any`        | Supply a list of attributes key+value                                                                                                                                                      |
| `.Code(string)`            | `err.Code() string`                   | Set a code or slug that describes the error. Error messages are intented to be read by humans, but such code is expected to be read by machines and be transported over different services |
| `.Time(time.Time)`         | `err.Time() time.Time`                | Set the error time (default: `time.Now()`)                                                                                                                                                 |
| `.Since(time.Time)`        | `err.Duration() time.Duration`        | Set the error duration                                                                                                                                                                     |
| `.Duration(time.Duration)` | `err.Duration() time.Duration`        | Set the error duration                                                                                                                                                                     |
| `.In(string)`              | `err.Domain() string`                 | Set the feature category or domain                                                                                                                                                         |
| `.Tags(...string)`         | `err.Tags() []string`                 | Add multiple tags, describing the feature returning an error                                                                                                                               |
| `.Tx(string)`              | `err.Transaction() string`            | Add a transaction id, trace id, correlation id...                                                                                                                                          |
| `.Hint(string)`            | `err.Hint() string`                   | Set a hint for faster debugging                                                                                                                                                            |
| `.Owner(string)`           | `err.Owner() (string)`                | Set the name/email of the collegue/team responsible for handling this error. Useful for alerting purpose                                                                                   |
| `.User(string, any...)`    | `err.User() (string, map[string]any)` | Supply user id and a chain of key/value                                                                                                                                                    |

### Other helpers

- `oops.AsError(error) (oops.OopsError, bool)` as an alias to `errors.As(...)`

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

err.(oops.OopsError).Stacktrace()
// Oops: permission denied
//   --- at github.com/samber/oops/loggers/slog/example.go:20 (d)
//   --- at github.com/samber/oops/loggers/slog/example.go:24 (c)
//   --- at github.com/samber/oops/loggers/slog/example.go:32 (b)
//   --- at github.com/samber/oops/loggers/slog/example.go:36 (a)
//   --- at github.com/samber/oops/loggers/slog/example.go:42 (main)
```

Wrapping errors will be reported as an annotated stack trace:

```go
err1 := oops.Errorf("permission denied")
// ...
err2 := oops.Wrapf(err, "something failed")

err2.(oops.OopsError).Stacktrace()
// Oops: permission denied
//   --- at github.com/samber/oops/loggers/slog/example.go:20 (d)
//   --- at github.com/samber/oops/loggers/slog/example.go:24 (c)
// Thrown: something failed
//   --- at github.com/samber/oops/loggers/slog/example.go:32 (b)
//   --- at github.com/samber/oops/loggers/slog/example.go:36 (a)
//   --- at github.com/samber/oops/loggers/slog/example.go:42 (main)
```

### Source fragments

The exact error location can be provided in a Go file extract.

Source fragments are hidden by default. You must run `oops.SourceFragmentsHidden = false` to enable this feature. Go source files being read at run time, you have to keep the source code at the same location.

In a future release, this library is expected to output a colorized extract. Please contribute!

```go
oops.SourceFragmentsHidden = false

err1 := oops.Errorf("permission denied")
// ...
err2 := oops.Wrapf(err, "something failed")

err2.(oops.OopsError).Sources()
```

Output:

```txt
Oops: permission denied
github.com/samber/oops/examples/sources/example.go:22 d()
17                      Time(time.Now()).
18                      With("user_id", 1234).
19                      With("permission", "post.create").
20                      Hint("Runbook: https://doc.acme.org/doc/abcd.md").
21                      User("user-123", "firstname", "john", "lastname", "doe").
22                      Errorf("permission denied")
                        ^^^^^^^^^^^^^^^^^^^^^^^^^^^
23      }
24
25      func c() error {
26              return d()
27      }

Thrown: something failed
github.com/samber/oops/examples/sources/example.go:34 b()
29      func b() error {
30              return oops.
31                      In("iam").
32                      Tx("6710668a-2b2a-4de6-b8cf-3272a476a1c9").
33                      With("hello", "world").
34                      Wrapf(c(), "something failed")
                        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
35      }
36
37      func a() error {
38              return b()
39      }
```

### Output

#### Errorf `%w`

```go
str := fmt.Errorf("something failed: %w", oops.Errorf("permission denied"))

err.Error()
// Output:
// something failed: permission denied
```

#### printf `%v`

```go
str := fmt.Sprintf("%+v", oops.Errorf("permission denied"))

// Output:
// permission denied
```

#### printf `%+v`

```go
str := fmt.Sprintf("%+v", oops.Errorf("permission denied"))

// Output:
// Oops: permission denied
// Code: "iam_missing_permission"
// At: 2023-05-02 05:26:48.570837 +0000 UTC
// Duration: 42ms
// Domain: authz
// Tags: iam, authz
// Transaction: 092abdf7-a0ad-40cd-bfdc-d25a9435a87d
// Hint: Runbook: https://doc.acme.org/doc/abcd.md
// Owner: authz-team@acme.org
// Context:
//   * user_id: 1234
// User:
//   * id: user-123
//   * firstname: john
//   * lastname: doe
// Oops: permission denied
//     --- at github.com/samber/oops/loggers/slog/example.go:20 (d)
//     --- at github.com/samber/oops/loggers/slog/example.go:24 (c)
//     --- at github.com/samber/oops/loggers/slog/example.go:32 (b)
//     --- at github.com/samber/oops/loggers/slog/example.go:36 (a)
//     --- at github.com/samber/oops/loggers/slog/example.go:42 (main)
```

#### JSON Marshal

```go
b := json.MarshalIndent(err, "", "  ")

// Output:
// {
//   "code": "iam_missing_permission",
//   "context": {
//     "user_id": 1234
//   },
//   "domain": "authz",
//   "tags": [
//     "iam",
//     "authz"
//   ],
//   "error": "Permission denied",
//   "hint": "Runbook: https://doc.acme.org/doc/abcd.md",
//   "time": "2023-05-02T05:26:48.570837Z",
//   "duration": "42ms",
//   "stacktrace": "Oops: permission denied
//     --- at github.com/samber/oops/loggers/slog/example.go:20 (d)
//     --- at github.com/samber/oops/loggers/slog/example.go:24 (c)
//     --- at github.com/samber/oops/loggers/slog/example.go:32 (b)
//     --- at github.com/samber/oops/loggers/slog/example.go:36 (a)
//     --- at github.com/samber/oops/loggers/slog/example.go:42 (main)",
//   "transaction": "4ab0e35e-8414-4d76-b09e-cba80c983e4b",
//   "user": {
//     "firstname": "john",
//     "id": "user-123",
//     "lastname": "doe"
//   }
// }

```

#### slog.Valuer

```go
attr := slog.Any("error")

// Output:
// slog.Group("error", ...)
```

## 🥷 Tips and best practices

### Wrap/Wrapf shortcut

`oops.Wrap(...)` and `oops.Wrapf(...)` return nil if the provided `error` is nil.

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
        Tx("77cb6664").
        With("hello", "world").
        Wrap(err)
}

err = mayFail2()
if err != nil {
    return oops.
        In("iam").
        Tx("77cb6664").
        With("hello", "world").
        Wrap(err)
}

return oops.
    In("iam").
    Tx("77cb6664").
    With("hello", "world").
    Wrap(mayFail3())
```

✅ but write:

```go
errorBuilder := oops.
    In("iam").
    Tx("77cb6664").
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
		Tx("4ea76885-a371-46b0-8ce0-b72b277fa9af").
		Time(time.Now()).
		With("hello", "world").
		With("user_id", 1234).
		With("permission", "post.create").
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		By("user-123", "firstname", "john", "lastname", "doe").
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
		Tx("4ea76885-a371-46b0-8ce0-b72b277fa9af").
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
		With("user_id", 1234).
		With("permission", "post.create").
		Hint("Runbook: https://doc.acme.org/doc/abcd.md").
		By("user-123", "firstname", "john", "lastname", "doe").
		Errorf("permission denied")
}
```

## 📫 Loggers

Some loggers may need a custom formatter to extract attributes from `oops.OopsError`.

Available loggers:
- slog: [example](https://github.com/samber/oops/examples/slog)
- logrus: [formatter](https://github.com/samber/oops/loggers/logrus) + [example](https://github.com/samber/oops/examples/logrus)

We are looking for contributions and examples for:
- zap
- zerolog
- go-sentry
- other?

Examples of formatters can be found in `Format()`, `Marshal()` and `LogValuer` methods of `oops.OopsError`.

## 🤝 Contributing

- Ping me on twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
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
