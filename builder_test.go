package oops

import (
	"context"
	"testing"
)

type stringerContextKey int

func (k stringerContextKey) String() string { return "request_id" }

func TestWithContextPreservesStringerKeyIdentity(t *testing.T) {
	t.Parallel()
	key := stringerContextKey(1)
	ctx := context.WithValue(context.Background(), key, "typed value")
	// A string key with the same display name must not shadow the typed key.
	ctx = context.WithValue(ctx, "request_id", "string value") //nolint:staticcheck,revive
	wrapped := WithContext(ctx, key).Errorf("failed")
	err, ok := wrapped.(OopsError)
	if !ok {
		t.Fatalf("unexpected error type: %T", wrapped)
	}
	if got := err.context["request_id"]; got != "typed value" {
		t.Fatalf("context value = %v", got)
	}
}
