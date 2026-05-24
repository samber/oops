package oops

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsOops(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "oops_error_found",
			run: func(is *assert.Assertions) {
				err := In("test").With("key", "value").Errorf("oops error")
				oopsErr, ok := AsOops(err)
				is.True(ok)
				is.Equal("test", oopsErr.Domain())
				is.Equal(map[string]any{"key": "value"}, oopsErr.Context())
			},
		},
		{
			name: "non_oops_returns_false",
			run: func(is *assert.Assertions) {
				_, ok := AsOops(errors.New("plain error"))
				is.False(ok)
			},
		},
		{
			name: "nil_returns_false",
			run: func(is *assert.Assertions) {
				_, ok := AsOops(nil)
				is.False(ok)
			},
		},
		{
			name: "wrapped_oops_found",
			run: func(is *assert.Assertions) {
				err := In("test").With("key", "value").Errorf("oops error")
				wrapped := fmt.Errorf("wrapper: %w", err)
				oopsErr, ok := AsOops(wrapped)
				is.True(ok)
				is.Equal("test", oopsErr.Domain())
			},
		},
		{
			name: "deeply_nested_oops_found",
			run: func(is *assert.Assertions) {
				err := In("test").With("key", "value").Errorf("oops error")
				deepWrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", err))
				oopsErr, ok := AsOops(deepWrapped)
				is.True(ok)
				is.Equal("test", oopsErr.Domain())
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

func TestAsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "matches_concrete_type",
			run: func(is *assert.Assertions) {
				fsErr, ok := AsError[*fs.PathError](fmt.Errorf("path: %w", &fs.PathError{Op: "open", Path: "/tmp", Err: fs.ErrExist}))
				is.True(ok)
				is.Equal("open", fsErr.Op)
			},
		},
		{
			name: "oops_via_AsError",
			run: func(is *assert.Assertions) {
				err := Wrapf(fs.ErrExist, "wrapped")
				oopsErr, ok := AsError[OopsError](err)
				is.True(ok)
				is.NotEmpty(oopsErr.Span())
			},
		},
		{
			name: "non_matching_returns_false",
			run: func(is *assert.Assertions) {
				_, ok := AsError[*fs.PathError](errors.New("not a path error"))
				is.False(ok)
			},
		},
		{
			name: "nil_returns_false",
			run: func(is *assert.Assertions) {
				_, ok := AsError[OopsError](nil)
				is.False(ok)
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

// https://github.com/samber/oops/issues/95
func TestErrorsIsSemanticsWithOopsErrorTargets(t *testing.T) {
	t.Parallel()

	t.Run("no_panic_when_target_is_non_comparable_oops_error", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped := With("k", "v").Wrap(base)

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("errors.Is panicked: %v", r)
			}
		}()

		_ = errors.Is(wrapped, base)
	})

	t.Run("oops_error_is_reflexive", func(t *testing.T) {
		t.Parallel()

		err := New("boom")
		if !errors.Is(err, err) {
			t.Fatalf("expected errors.Is(err, err) to be true")
		}
	})

	t.Run("different_oops_errors_should_not_match_each_other_even_if_same_cause", func(t *testing.T) {
		t.Parallel()

		err1 := Wrap(io.EOF)
		err2 := With("ctx", "x").Wrap(io.EOF)

		if errors.Is(err1, err2) || errors.Is(err2, err1) {
			t.Fatalf("expected errors.Is to be false between distinct oops errors; got true")
		}
	})
}
