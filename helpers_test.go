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
	is := assert.New(t)

	// OopsError is found
	err := In("test").With("key", "value").Errorf("oops error")
	oopsErr, ok := AsOops(err)
	is.True(ok)
	is.Equal("test", oopsErr.Domain())
	is.Equal(map[string]any{"key": "value"}, oopsErr.Context())

	// Non-oops error returns false
	plainErr := errors.New("plain error")
	_, ok = AsOops(plainErr)
	is.False(ok)

	// Nil error returns false
	_, ok = AsOops(nil)
	is.False(ok)

	// OopsError wrapped by fmt.Errorf is found
	wrapped := fmt.Errorf("wrapper: %w", err)
	oopsErr, ok = AsOops(wrapped)
	is.True(ok)
	is.Equal("test", oopsErr.Domain())

	// Deeply nested OopsError is found
	deepWrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", err))
	oopsErr, ok = AsOops(deepWrapped)
	is.True(ok)
	is.Equal("test", oopsErr.Domain())
}

func TestAsError(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Matches concrete error type
	err := Wrapf(fs.ErrExist, "wrapped")
	fsErr, ok := AsError[*fs.PathError](fmt.Errorf("path: %w", &fs.PathError{Op: "open", Path: "/tmp", Err: fs.ErrExist}))
	is.True(ok)
	is.Equal("open", fsErr.Op)

	// OopsError matched via AsError
	oopsErr, ok := AsError[OopsError](err)
	is.True(ok)
	is.NotEmpty(oopsErr.Span())

	// Non-matching type returns false
	_, ok = AsError[*fs.PathError](errors.New("not a path error"))
	is.False(ok)

	// Nil error returns false
	_, ok = AsError[OopsError](nil)
	is.False(ok)
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
