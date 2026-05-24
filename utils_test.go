package oops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoalesceOrEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "non_empty_values",
			run: func(is *assert.Assertions) {
				is.Equal("test", coalesceOrEmpty("", "test", "another"))
			},
		},
		{
			name: "all_empty",
			run: func(is *assert.Assertions) {
				is.Empty(coalesceOrEmpty("", "", ""))
			},
		},
		{
			name: "no_values",
			run: func(is *assert.Assertions) {
				is.Empty(coalesceOrEmpty[string]())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			tt.run(is)
		})
	}
}

func TestContextValueOrNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "value_found",
			run: func(is *assert.Assertions) {
				ctx := context.WithValue(context.Background(), "key", "value") //nolint:staticcheck
				is.Equal("value", contextValueOrNil(ctx, "key"))
			},
		},
		{
			name: "key_not_found",
			run: func(is *assert.Assertions) {
				ctx := context.WithValue(context.Background(), "key", "value") //nolint:staticcheck
				is.Nil(contextValueOrNil(ctx, "nonexistent"))
			},
		},
		{
			name: "nil_context",
			run: func(is *assert.Assertions) {
				is.Nil(contextValueOrNil(nil, "key")) //nolint:staticcheck
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			tt.run(is)
		})
	}
}
