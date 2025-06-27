# Oops Examples

This document provides comprehensive examples of how to use the Oops error handling library in various scenarios.

## Basic Usage

### Simple Error Creation

```go
package main

import (
    "fmt"
    "github.com/samber/oops"
)

func main() {
    // Basic error
    err := oops.New("something went wrong")
    fmt.Println(err)
    
    // Formatted error
    err = oops.Errorf("failed to process user %d", 123)
    fmt.Println(err)
}
```

### Error Wrapping

```go
func processUser(userID int) error {
    // Simulate some operation that might fail
    if userID <= 0 {
        return oops.Errorf("invalid user ID: %d", userID)
    }
    
    // Simulate database operation
    if err := databaseOperation(userID); err != nil {
        return oops.Wrapf(err, "failed to process user %d", userID)
    }
    
    return nil
}

func databaseOperation(userID int) error {
    // Simulate database error
    return fmt.Errorf("connection timeout")
}
```

## Context and Attributes

### Adding Rich Context

```go
func authenticateUser(username, password string) error {
    if username == "" {
        return oops.
            Code("auth_missing_username").
            In("authentication").
            Tags("auth", "validation").
            With("username", username).
            Hint("Username is required for authentication").
            Errorf("missing username")
    }
    
    if password == "" {
        return oops.
            Code("auth_missing_password").
            In("authentication").
            Tags("auth", "validation").
            With("username", username).
            Hint("Password is required for authentication").
            Errorf("missing password for user %s", username)
    }
    
    // Simulate authentication failure
    return oops.
        Code("auth_invalid_credentials").
        In("authentication").
        Tags("auth", "security").
        With("username", username).
        With("attempt_count", 1).
        Hint("Check if user exists and password is correct").
        Public("Invalid username or password").
        Errorf("authentication failed for user %s", username)
}
```

### User and Tenant Context

```go
func processOrder(userID string, orderID string, tenantID string) error {
    userData := map[string]any{
        "email": "john@example.com",
        "role":  "customer",
        "plan":  "premium",
    }
    
    tenantData := map[string]any{
        "name": "Acme Corp",
        "plan": "enterprise",
        "region": "us-west",
    }
    
    // Simulate order processing error
    return oops.
        Code("order_processing_failed").
        In("order_management").
        Tags("order", "processing").
        Trace("e76031ee-a0c4-4a80-88cb-17086fdd19c0").
        Span("order_creation").
        User(userID, userData).
        Tenant(tenantID, tenantData).
        With("order_id", orderID).
        With("amount", 99.99).
        Hint("Check inventory and payment processing").
        Owner("orders-team@company.com").
        Errorf("failed to process order %s for user %s", orderID, userID)
}
```

## HTTP Context

### Web Handler with Request Context

```go
package main

import (
    "net/http"
    "github.com/samber/oops"
)

func userHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("user_id")
    
    if userID == "" {
        err := oops.
            Code("missing_user_id").
            In("user_api").
            Tags("api", "validation").
            Request(r, false).
            With("method", r.Method).
            With("path", r.URL.Path).
            Hint("User ID is required as query parameter").
            Public("User ID is required").
            Errorf("missing user_id parameter")
        
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process user request
    if err := processUserRequest(userID); err != nil {
        // Wrap with HTTP context
        httpErr := oops.
            Request(r, false).
            Wrapf(err, "failed to process user request")
        
        http.Error(w, httpErr.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("User processed successfully"))
}

func processUserRequest(userID string) error {
    // Simulate processing error
    return oops.Errorf("user %s not found in database", userID)
}
```

## Panic Handling

### Recovering from Panics

```go
func safeOperation() error {
    return oops.Recover(func() {
        // This might panic
        panic("unexpected error occurred")
    })
}

func safeOperationWithContext() error {
    return oops.Recoverf(func() {
        // This might panic
        panic("database connection lost")
    }, "operation failed due to %s", "system error")
}

// Usage in HTTP handler
func safeHandler(w http.ResponseWriter, r *http.Request) {
    if err := safeOperation(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}
```

## Assertions

### Input Validation

```go
func validateUser(userID int, email string) error {
    // Assert user ID is positive
    oops.Assert(userID > 0)
    
    oops.
        With("email", email).
        With("user_id", userID).
        // Assert email is not empty
        Assertf(email != "", "email cannot be empty").
        // Assert email format (simplified)
        Assertf(len(email) > 5 && len(email) < 100, "email length must be between 5 and 100 characters, got %d", len(email))
    
    return nil
}

// Usage
func main() {
    if err := validateUser(0, ""); err != nil {
        fmt.Println("Validation failed:", err)
    }
}
```

## Time and Duration Tracking

### Performance Monitoring

```go
func expensiveOperation() error {
    start := time.Now()
    
    // Simulate expensive operation
    time.Sleep(100 * time.Millisecond)
    
    // Check if operation took too long
    if time.Since(start) > 50*time.Millisecond {
        return oops.
            Code("operation_timeout").
            In("performance").
            Tags("performance", "timeout").
            Since(start).
            Hint("Consider optimizing the operation or increasing timeout").
            Errorf("operation took longer than expected")
    }
    
    return nil
}

func timedOperation() error {
    return oops.
        Duration(5 * time.Second).
        Time(time.Now()).
        Errorf("operation completed")
}
```

## Context Integration

### Using Go Context

```go
package main

import (
    "context"
    "github.com/samber/oops"
)

func processWithContext(ctx context.Context, userID string) error {
    // Add context values to error
    err := oops.
        WithContext(ctx, "request_id", "user_id").
        Errorf("failed to process user %s", userID)
    
    return err
}

// Usage
func main() {
    ctx := context.WithValue(context.Background(), "request_id", "req-123")
    ctx = context.WithValue(ctx, "user_id", "user-456")
    
    if err := processWithContext(ctx, "user-456"); err != nil {
        fmt.Println("Error:", err)
    }
}
```

## Logger Integration

### With Zap Logger

```go
package main

import (
    "github.com/samber/oops"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    err := oops.
        Code("database_error").
        In("database").
        Tags("database", "connection").
        With("host", "localhost").
        With("port", 5432).
        Hint("Check database server status and network connectivity").
        Errorf("failed to connect to database")
    
    logger.Error(err.Error(),
        zap.Any("error", err),
    )
}
```

### With Zerolog

```go
package main

import (
    "os"
    "github.com/samber/oops"
    "github.com/rs/zerolog"
)

func main() {
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    
    err := oops.
        Code("file_not_found").
        In("file_system").
        Tags("file", "io").
        With("path", "/tmp/file.txt").
        Hint("Check if file exists and has proper permissions").
        Errorf("file not found: %s", "/tmp/file.txt")
    
    logger.Error().
        Interface("error", err).
        Msg(err.Error())
}
```

### With Logrus

```go
package main

import (
    "github.com/samber/oops"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    err := oops.
        Code("network_timeout").
        In("network").
        Tags("network", "timeout").
        With("timeout", "30s").
        With("retries", 3).
        Hint("Check network connectivity and server status").
        Errorf("network request timed out after %s", "30s")
    
    logger.WithFields(logrus.Fields{
        "error": err,
    }).Error(err.Error())
}
```

### With Slog (Go 1.21+)

```go
package main

import (
    "log/slog"
    "os"
    "github.com/samber/oops"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    err := oops.
        Code("validation_error").
        In("validation").
        Tags("validation", "input").
        With("field", "email").
        With("value", "invalid-email").
        Hint("Email must be in valid format (e.g., user@domain.com)").
        Public("Please enter a valid email address").
        Errorf("invalid email format: %s", "invalid-email")
    
    logger.Error("Validation failed",
        slog.Any("error", err),
        slog.String("operation", "validate_email"),
    )
}
```

## Error Chaining

### Complex Error Propagation

```go
func complexOperation() error {
    // Step 1: Validate input
    if err := validateInput(); err != nil {
        return oops.
            In("validation").
            Tags("validation", "input").
            Wrapf(err, "input validation failed")
    }
    
    // Step 2: Process data
    if err := processData(); err != nil {
        return oops.
            In("processing").
            Tags("processing", "data").
            Wrapf(err, "data processing failed")
    }
    
    // Step 3: Save results
    if err := saveResults(); err != nil {
        return oops.
            In("persistence").
            Tags("persistence", "database").
            Wrapf(err, "failed to save results")
    }
    
    return nil
}

func validateInput() error {
    return oops.With("key", "value").Errorf("input validation failed")
}

func processData() error {
    return oops.With("key", "value").Errorf("data processing failed")
}

func saveResults() error {
    return oops.With("key", "value").Errorf("failed to save to database")
}
```

## Best Practices

### Error Code Standards

```go
// Define error codes as constants
const (
    ErrCodeAuthFailed        = "auth_failed"
    ErrCodeValidationFailed  = "validation_failed"
    ErrCodeDatabaseError     = "database_error"
    ErrCodeNetworkTimeout    = "network_timeout"
    ErrCodeFileNotFound      = "file_not_found"
    ErrCodePermissionDenied  = "permission_denied"
)

func authenticateUser(username, password string) error {
    if username == "" {
        return oops.
            Code(ErrCodeValidationFailed).
            In("authentication").
            Tags("auth", "validation").
            With("field", "username").
            Hint("Username is required").
            Errorf("missing username")
    }
    
    // Simulate authentication failure
    return oops.
        Code(ErrCodeAuthFailed).
        In("authentication").
        Tags("auth", "security").
        With("username", username).
        Hint("Check credentials").
        Public("Invalid username or password").
        Errorf("authentication failed for user %s", username)
}
```

### Consistent Error Structure

```go
// Define common error patterns
func newValidationError(field, value, hint string) error {
    return oops.
        Code("validation_error").
        In("validation").
        Tags("validation", "input").
        With("field", field).
        With("value", value).
        Hint(hint).
        Errorf("validation failed for field %s", field)
}

func newDatabaseError(operation, table string, err error) error {
    return oops.
        Code("database_error").
        In("database").
        Tags("database", "sql").
        With("operation", operation).
        With("table", table).
        Hint("Check database connection and table structure").
        Wrapf(err, "database operation failed: %s on %s", operation, table)
}

// Usage
func validateUser(user User) error {
    if user.Email == "" {
        return newValidationError("email", user.Email, "Email is required")
    }
    
    if user.Age < 0 {
        return newValidationError("age", fmt.Sprintf("%d", user.Age), "Age must be positive")
    }
    
    return nil
}
``` 