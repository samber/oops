package oops

import (
	"errors"
	"fmt"
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
