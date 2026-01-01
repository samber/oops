# Oops Architecture

This document describes the internal architecture and design decisions of the Oops error handling library.

## Overview

Oops is designed as a fluent builder pattern library that creates rich, structured errors with contextual information. The library follows Go's error interface while providing additional capabilities for structured logging and debugging.

## Core Components

### 1. Error Builder Pattern

The library uses a fluent builder pattern to construct errors with rich context:

```go
err := oops.
    Code("auth_failed").
    In("authentication").
    Tags("security", "auth").
    With("user_id", 123).
    Hint("Check user permissions").
    Errorf("authentication failed")
```

### 2. Core Types

#### OopsErrorBuilder Interface

```go
type OopsErrorBuilder interface {
    // Context methods
    Code(code any) OopsErrorBuilder
    In(domain string) OopsErrorBuilder
    Tags(tags ...string) OopsErrorBuilder
    Trace(trace string) OopsErrorBuilder
    Span(span string) OopsErrorBuilder
    
    // Time methods
    Time(time time.Time) OopsErrorBuilder
    Since(time time.Time) OopsErrorBuilder
    Duration(duration time.Duration) OopsErrorBuilder
    
    // Attribute methods
    With(kv ...any) OopsErrorBuilder
    WithContext(ctx context.Context, keys ...any) OopsErrorBuilder
    
    // Entity methods
    User(userID string, data map[string]any) OopsErrorBuilder
    Tenant(tenantID string, data map[string]any) OopsErrorBuilder
    
    // HTTP methods
    Request(req *http.Request, withBody bool) OopsErrorBuilder
    Response(res *http.Response, withBody bool) OopsErrorBuilder
    
    // Debug methods
    Hint(hint string) OopsErrorBuilder
    Public(public string) OopsErrorBuilder
    Owner(owner string) OopsErrorBuilder
    
    // Error creation methods
    New(message string) error
    Errorf(format string, args ...any) error
    Wrap(err error) error
    Wrapf(err error, format string, args ...any) error
    Join(errs ...error) error
    
    // Assertion methods
    Assert(condition bool) OopsErrorBuilder
    Assertf(condition bool, format string, args ...any) OopsErrorBuilder
}
```

### 3. Error Wrapping

The library supports wrapping existing errors:

```go
// ❌ Bad
err := mayFail()
if err != nil {
   return oops.With("key", "value").
        Tenant("user-123").
        With("error", err).
        Errorf("an error")
}
return nil

// ✅ Good
return oops.With("key", "value").
    Tenant("user-123").
    Wrap(mayFail(), "an error")
```

With multiple parameters:

```go
// ❌ Bad
a, b, c, err := mayFail()
if err != nil {
    return a, b, c, oops.Wrap(err)
}
return a, b, c, nil

// ✅ Good
return oops.With("key", "value").
    Tenant("user-123").
    Wrap3(mayFail(), "an error")
```

## Performance Considerations

### 1. Lazy Evaluation

Some expensive operations are performed lazily:

```go
func (e *oopsError) ToMap() slog.Attr {
    if e.stackTrace != nil {
        output["stacktrace"] = e.stacktrace.format()
    }
    return output
}
```

### 2. String Formatting

String formatting is deferred until needed:

```go
func (e *oopsError) Error() string {
    if e.err != nil {
        return e.format()
    }

    return ""
}
```

## Error Serialization

### JSON Format

Errors can be serialized to JSON for logging:

```go
func (e *oopsError) MarshalJSON() ([]byte, error) {
    data := map[string]any{
        "message": e.message,
        "code":    e.code,
        "domain":  e.domain,
        "tags":    e.tags,
        "time":    e.time,
        "attributes": e.attributes,
        // ... other fields
    }
    
    if e.cause != nil {
        data["cause"] = e.cause.Error()
    }
    
    return json.Marshal(data)
}
```

### Structured Logging

The library provides structured logging support:

```go
func (e *oopsError) ToMap() map[string]any {
    return map[string]any{
        "error": map[string]any{
            "message":    e.message,
            "code":       e.code,
            "domain":     e.domain,
            "tags":       e.tags,
            "attributes": e.attributes,
            "hint":       e.hint,
            "public":     e.public,
            "owner":      e.owner,
        },
    }
}
```

## Conclusion

The Oops library is designed with a focus on:

1. **Simplicity**: Easy to use and understand
2. **Performance**: Minimal overhead
3. **Flexibility**: Extensible and customizable
4. **Integration**: Works with existing logging systems
5. **Debugging**: Rich context for faster issue resolution

The architecture supports these goals through careful design of interfaces, efficient data structures, and thoughtful integration patterns. 
