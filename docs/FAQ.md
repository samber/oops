# Frequently Asked Questions (FAQ)

This document answers common questions about the Oops error handling library.

## General Questions

### What is Oops?

Oops is a Go error handling library that provides structured error management with rich contextual information. It's designed as a drop-in replacement for Go's built-in error handling, adding features like stack traces, structured attributes, debugging hints, and logger integration.

### Why should I use Oops instead of standard Go errors?

Oops provides several advantages over standard Go errors:

- **Rich Context**: Add structured attributes, user info, request/response data
- **Automatic Stacktraces**: No need to manually capture stacktraces
- **Debugging Hints**: Include helpful information for faster issue resolution
- **Logger Integration**: Works seamlessly with popular logging libraries
- **Error Codes**: Machine-readable error identification
- **User-Safe Messages**: Separate internal and user-facing error messages

### Is Oops a logging library?

No, Oops is **NOT** a logging library. It's an error handling library that complements your existing logging toolchain. Oops creates rich, structured errors that can then be logged using your preferred logging library (Zap, Zerolog, Logrus, Slog, etc.).

### What are the performance implications of using Oops?

Oops has minimal performance overhead:

- **CPU**: Negligible impact on most applications
- **Dependencies**: Zero external dependencies
- **lazy** error formatting

For most applications, the benefits of rich error context far outweigh the minimal performance cost.

## Usage Questions

### How do I create a simple error?

```go
// Basic error
err := oops.New("something went wrong")

// Formatted error
err := oops.Errorf("failed to process user %d", userID)
```

### How do I wrap an existing error?

```go
// Simple wrapping
err := oops.Wrap(originalError)

// Wrapping with additional context
err := oops.Wrapf(originalError, "failed to process user %d", userID)
```

### How do I add context to an error?

```go
err := oops.
    Code("auth_failed").
    In("authentication").
    Tags("security", "auth").
    User("123", "email", "foo@bar.com").
    With("attempt_count", 3).
    Hint("Check user permissions in database").
    Errorf("authentication failed")
```

### How do I handle panics?

```go
// Recover from panic
err := oops.Recover(func() {
    // code that might panic
    panic("something went wrong")
})

// Recover with custom message
err := oops.Recoverf(func() {
    panic("internal error")
}, "operation failed: %s", "timeout")
```

### How do I use assertions?

```go
// Simple assertion
oops.Assert(userID > 0).Errorf("user ID must be positive")

// Assertion with custom message
oops.Assertf(userID > 0, "user ID must be positive, got %d", userID)
```

## Integration Questions

### How do I integrate with Zap logger?

```go
import (
    "github.com/samber/oops"
    "go.uber.org/zap"
)

logger, _ := zap.NewProduction()
defer logger.Sync()

err := oops.Errorf("database connection failed")
logger.Error("error occurred", zap.Any("error", err))
```

### How do I integrate with Zerolog?

```go
import (
    "github.com/samber/oops"
    "github.com/rs/zerolog"
)

logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

err := oops.Errorf("authentication failed")
logger.Error().Interface("error", err).Msg("auth error")
```

### How do I integrate with Logrus?

```go
import (
    "github.com/samber/oops"
    "github.com/sirupsen/logrus"
)

logger := logrus.New()

err := oops.Errorf("file not found")
logger.WithField("error", err).Error("file operation failed")
```

### How do I integrate with Slog (Go 1.21+)?

```go
import (
    "github.com/samber/oops"
    "log/slog"
)

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

err := oops.Errorf("network timeout")
logger.Error("network error", slog.Any("error", err))
```

### How do I handle HTTP requests and responses?

```go
func handler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("user_id")
    
    if userID == "" {
        err := oops.
            Code("missing_user_id").
            Request(r, false).
            With("method", r.Method).
            With("path", r.URL.Path).
            Errorf("missing user_id parameter")
        
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process request...
}
```

## Error Context Questions

### What is the difference between `Code()` and `In()`?

- **`Code()`**: Sets a machine-readable error code (e.g., "auth_failed", "validation_error")
- **`In()`**: Sets the domain or feature category where the error occurred (e.g., "authentication", "database")

```go
err := oops.
    Code("auth_failed").        // Machine-readable code
    In("authentication").       // Domain/feature category
    Errorf("authentication failed")
```

### When should I use `Tags()` vs `With()`?

- **`Tags()`**: Use for categorization and filtering (e.g., "security", "auth", "validation")
- **`With()`**: Use for specific key-value attributes (e.g., user_id, request_id, timestamp)

```go
err := oops.
    Tags("security", "auth").           // Categories for filtering
    With("user_id", 123).              // Specific attributes
    With("request_id", "req-456").
    Errorf("authentication failed")
```

### What is the purpose of `Hint()`?

`Hint()` provides debugging hints for developers to help them understand and fix issues quickly:

```go
err := oops.
    Hint("Check database connection and user permissions").
    Errorf("database error")
```

### What is the difference between the error message and `Public()`?

- **Error message**: Internal message for developers (may contain sensitive information)
- **`Public()`**: User-safe message that can be shown to end users

```go
err := oops.
    Public("Unable to process your request. Please try again later.").
    Errorf("internal server error: database connection failed")
```

## Performance Questions

### Does Oops capture stack traces by default?

Yes, Oops automatically captures stack traces when errors are created. This provides valuable debugging information without requiring manual stack trace handling.

### How much memory does an Oops error use?

An Oops error typically uses 200-500 bytes more than a standard Go error, depending on the amount of context added. This is negligible for most applications.

### Can I disable stack trace capture for performance?

Currently, stack trace capture cannot be disabled. However, the performance impact is minimal, and the debugging benefits usually outweigh the small cost.

### Does Oops use reflection?

No, Oops does not use reflection. It uses standard Go interfaces and type assertions for type safety and performance.

## Migration Questions

### How do I migrate from standard Go errors?

Replace `fmt.Errorf` with `oops.Errorf` and add context as needed:

```go
// Before
return fmt.Errorf("failed to process user %d: %w", userID, err)

// After
return oops.
    In("user_processing").
    With("user_id", userID).
    Wrapf(err, "failed to process user %d", userID)
```

### How do I migrate from pkg/errors?

Replace `errors.Wrapf` with `oops.Wrapf` and add additional context:

```go
// Before
return errors.Wrapf(err, "failed to process user %d", userID)

// After
return oops.
    In("user_processing").
    With("user_id", userID).
    Wrapf(err, "failed to process user %d", userID)
```

### Can I use Oops alongside other error libraries?

Yes, Oops can be used alongside other error libraries. It implements the standard `error` interface and can wrap errors from other libraries.

## Testing Questions

### How do I test Oops errors?

```go
func TestUserProcessing(t *testing.T) {
    err := processUser(0)
    
    var oopsErr oops.OopsError
    if !errors.As(err, &oopsErr) {
        t.Fatal("expected OopsError")
    }
    
    if oopsErr.Code() != "invalid_user_id" {
        t.Errorf("expected code 'invalid_user_id', got %s", oopsErr.Code())
    }
    
    if oopsErr.Domain() != "user_processing" {
        t.Errorf("expected domain 'user_processing', got %s", oopsErr.Domain())
    }
}
```

### How do I mock Oops errors in tests?

```go
type MockOopsError struct {
    message string
    code    any
    domain  string
}

func (m *MockOopsError) Error() string { return m.message }
func (m *MockOopsError) Code() any    { return m.code }
func (m *MockOopsError) Domain() string { return m.domain }
// ... implement other methods as needed
```

## Troubleshooting Questions

### Why isn't my error showing up in logs?

Make sure you're passing the error object to your logger, not just the error message:

```go
// Wrong - only logs the message
logger.Error(err.Error())

// Correct - logs the full error context
logger.Error(err.Error(), zap.Any("error", err))
```

### How do I see the stack trace?

The stack trace is automatically captured and can be accessed through the error object:

```go
var oopsErr oops.OopsError
if errors.As(err, &oopsErr) {
    for _, frame := range oopsErr.StackTrace() {
        fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
    }
}
```

### How do I get the error code from an error?

```go
var oopsErr oops.OopsError
if errors.As(err, &oopsErr) {
    code := oopsErr.Code()
    fmt.Printf("Error code: %v\n", code)
}
```

### How do I check if an error is an Oops error?

```go
var oopsErr oops.OopsError
if errors.As(err, &oopsErr) {
    // This is an Oops error
    fmt.Printf("Code: %v\n", oopsErr.Code())
} else {
    // This is not an Oops error
    fmt.Printf("Standard error: %s\n", err.Error())
}
```

## Best Practices Questions

### What error codes should I use?

Use consistent, descriptive error codes:

```go
const (
    ErrCodeAuthFailed        = "auth_failed"
    ErrCodeValidationFailed  = "validation_failed"
    ErrCodeDatabaseError     = "database_error"
    ErrCodeNetworkTimeout    = "network_timeout"
    ErrCodeFileNotFound      = "file_not_found"
    ErrCodePermissionDenied  = "permission_denied"
)
```

### How should I structure error messages?

- Use clear, descriptive messages
- Include relevant identifiers
- Avoid sensitive information in error messages
- Use `Public()` for user-facing messages

```go
err := oops.
    Code("user_not_found").
    In("user_management").
    With("user_id", userID).
    Public("User not found").
    Errorf("user with ID %d not found in database", userID)
```

### When should I use error wrapping?

Use error wrapping when you want to add context to an existing error:

```go
if err := databaseOperation(userID); err != nil {
    return oops.
        In("user_processing").
        With("user_id", userID).
        Wrapf(err, "failed to process user %d", userID)
}
```

### How should I handle errors in HTTP handlers?

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("user_id")
    
    if userID == "" {
        err := oops.
            Code("missing_user_id").
            Request(r, false).
            Public("User ID is required").
            Errorf("missing user_id parameter")
        
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    if err := processUser(userID); err != nil {
        // Log the full error for debugging
        logger.Error("user processing failed", zap.Any("error", err))
        
        // Return user-safe message
        publicMsg := oops.GetPublic(err, "An unexpected error occurred")
        http.Error(w, publicMsg, http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}
```

## Version Compatibility

### What Go versions are supported?

Oops requires Go 1.21 or later.

### Is Oops backward compatible?

Yes, Oops follows semantic versioning. No breaking changes will be made before v2.0.0.

### How do I update to a new version?

```bash
go get -u github.com/samber/oops
```

## Support Questions

### Where can I get help?

- **GitHub Issues**: Report bugs and request features
- **GoDoc**: API documentation
- **Examples**: Check the `examples/` directory

### How do I contribute?

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Is Oops production-ready?

Yes, Oops is production-ready and used in many production applications. It has comprehensive test coverage and follows Go best practices. 
