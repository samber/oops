# Oops API Reference

This document provides a comprehensive reference for the Oops error handling library API.

## Core Functions

### Error Creation

#### `oops.New(message string) error`
Creates a new error with the given message.

```go
err := oops.New("something went wrong")
```

#### `oops.Errorf(format string, args ...any) error`
Creates a new error with formatted message.

```go
err := oops.Errorf("failed to process user %d: %s", userID, reason)
```

### Error Wrapping

#### `oops.Wrap(err error) error`
Wraps an existing error with additional context.

```go
err := oops.Wrap(originalError)
```

#### `oops.Wrapf(err error, format string, args ...any) error`
Wraps an existing error with formatted message.

```go
err := oops.Wrapf(originalError, "failed to process request: %s", requestID)
```

### Error Joining

#### `oops.Join(err1 error, err2 error, ...) error`
Joins multiple errors into a single error.

```go
err := oops.Join(err1, err2, err3)
```

## Builder Pattern Methods

The Oops library uses a fluent builder pattern. All methods return an `*oops.OopsErrorBuilder` that can be chained.

### Context Methods

#### `.Code(code string) OopsErrorBuilder`
Sets an error code for machine-readable identification.

```go
err := oops.Code("auth_failed").Errorf("authentication failed")
```

#### `.In(domain string) OopsErrorBuilder`
Sets the domain or feature category where the error occurred.

```go
err := oops.In("authentication").Errorf("user not found")
```

#### `.Tags(tags ...string) OopsErrorBuilder`
Adds tags for categorization and filtering.

```go
err := oops.Tags("security", "auth", "user").Errorf("permission denied")
```

#### `.Trace(trace string) OopsErrorBuilder`
Sets a transaction ID, trace ID, or correlation ID.

```go
err := oops.Trace("e76031ee-a0c4-4a80-88cb-17086fdd19c0").Errorf("request failed")
```

#### `.Span(span string) OopsErrorBuilder`
Sets a unit of work or operation identifier.

```go
err := oops.Span("user_creation").Errorf("failed to create user")
```

### Time and Duration

#### `.Time(time time.Time) OopsErrorBuilder`
Sets the error timestamp.

```go
err := oops.Time(time.Now()).Errorf("timeout occurred")
```

#### `.Since(time time.Time) OopsErrorBuilder`
Sets the error duration from a start time.

```go
start := time.Now()
// ... do work ...
err := oops.Since(start).Errorf("operation took too long")
```

#### `.Duration(duration time.Duration) OopsErrorBuilder`
Sets the error duration directly.

```go
err := oops.Duration(5 * time.Second).Errorf("operation timeout")
```

### Attributes and Context

#### `.With(kv ...any) OopsErrorBuilder`
Adds key-value pairs as attributes.

```go
err := oops.With("user_id", 123, "attempt", 3).Errorf("login failed")
```

#### `.WithContext(ctx context.Context, keys ...any) OopsErrorBuilder`
Extracts values from Go context.

```go
err := oops.WithContext(ctx, "request_id", "user_id").Errorf("context error")
```

### User and Tenant Information

#### `.User(userID string, data map[string]any) OopsErrorBuilder`
Adds user information and attributes.

```go
err := oops.User("user-123", map[string]any{
    "email": "john@example.com",
    "role":  "admin",
}).Errorf("user operation failed")
```

#### `.Tenant(tenantID string, data map[string]any) OopsErrorBuilder`
Adds tenant information and attributes.

```go
err := oops.Tenant("tenant-456", map[string]any{
    "name": "Acme Corp",
    "plan": "premium",
}).Errorf("tenant operation failed")
```

### HTTP Context

#### `.Request(req *http.Request, withBody bool) OopsErrorBuilder`
Adds HTTP request information.

```go
err := oops.Request(req, false).Errorf("request processing failed")
```

#### `.Response(res *http.Response, withBody bool) OopsErrorBuilder`
Adds HTTP response information.

```go
err := oops.Response(res, false).Errorf("response processing failed")
```

### Debugging and Hints

#### `.Hint(hint string) OopsErrorBuilder`
Adds debugging hints for developers.

```go
err := oops.Hint("Check database connection and user permissions").Errorf("database error")
```

#### `.Public(public string) OopsErrorBuilder`
Sets a user-safe error message.

```go
err := oops.Public("Unable to process your request. Please try again later.").Errorf("internal server error")
```

#### `.Owner(owner string) OopsErrorBuilder`
Sets the team or person responsible for handling the error.

```go
err := oops.Owner("backend-team@company.com").Errorf("service unavailable")
```

## Panic Handling

### `oops.Recover(cb func()) error`
Recovers from panics and converts them to errors.

```go
err := oops.Recover(func() {
    // code that might panic
    panic("something went wrong")
})
```

### `oops.Recoverf(cb func(), format string, args ...any) error`
Recovers from panics with custom error message.

```go
err := oops.Recoverf(func() {
    panic("internal error")
}, "operation failed: %s", "timeout")
```

## Assertions

### `oops.Assert(condition bool) OopsErrorBuilder`
Asserts a condition and panics if false.

```go
oops.Assert(userID > 0).Errorf("invalid user ID")
```

### `oops.Assertf(condition bool, format string, args ...any) OopsErrorBuilder`
Asserts a condition with custom error message.

```go
oops.Assertf(userID > 0, "user ID must be positive, got %d", userID)
```

## Context Integration

### `oops.FromContext(ctx context.Context) OopsErrorBuilder`
Retrieves an error builder from Go context.

```go
builder := oops.FromContext(ctx)
err := builder.Errorf("context error")
```

## Utility Functions

### `oops.GetPublic(err error, defaultPublicMessage string) string`
Extracts user-safe message from error.

```go
publicMsg := oops.GetPublic(err, "An unexpected error occurred")
```

## Error Interface

The `OopsError` type implements the standard `error` interface and provides additional methods:

```go
type OopsError interface {
    error
    Code() string
    Time() time.Time
    Duration() time.Duration
    Domain() string
    Tags() []string
    Trace() string
    Span() string
    Attributes() map[string]any
    Hint() string
    Public() string
    Owner() string
    User() (string, map[string]any)
    Tenant() (string, map[string]any)
    Request() *http.Request
    Response() *http.Response
    StackTrace() []Frame
    SourceFragments() []SourceFragment
}
```

## Integration Examples

### With Zap Logger

```go
logger, _ := zap.NewProduction()
err := oops.Errorf("database connection failed")
logger.Error("error occurred", zap.Any("error", err))
```

### With Zerolog

```go
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
err := oops.Errorf("authentication failed")
logger.Error().Interface("error", err).Msg("auth error")
```

### With Logrus

```go
logger := logrus.New()
err := oops.Errorf("file not found")
logger.WithField("error", err).Error("file operation failed")
```

### With Slog

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
err := oops.Errorf("network timeout")
logger.Error("network error", slog.Any("error", err))
``` 