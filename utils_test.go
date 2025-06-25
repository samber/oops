package oops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoalesceOrEmpty(t *testing.T) {
	is := assert.New(t)

	// Test with non-empty values
	result := coalesceOrEmpty("", "test", "another")
	is.Equal("test", result)

	// Test with all empty values
	result2 := coalesceOrEmpty("", "", "")
	is.Equal("", result2)

	// Test with no values
	result3 := coalesceOrEmpty[string]()
	is.Equal("", result3)
}

func TestContextValueOrNil(t *testing.T) {
	is := assert.New(t)

	// Test with context containing value
	ctx := context.WithValue(context.Background(), "key", "value")
	result := contextValueOrNil(ctx, "key")
	is.Equal("value", result)

	// Test with context not containing key
	result2 := contextValueOrNil(ctx, "nonexistent")
	is.Nil(result2)

	// Test with nil context
	result3 := contextValueOrNil(nil, "key")
	is.Nil(result3)
}
