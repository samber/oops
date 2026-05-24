package oops

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "Errorf",
			run: func(is *assert.Assertions) {
				err := Errorf("Error: %w", fs.ErrExist)
				is.ErrorIs(err, fs.ErrExist)
			},
		},
		{
			name: "Wrap",
			run: func(is *assert.Assertions) {
				err := Wrap(fs.ErrExist)
				is.ErrorIs(err, fs.ErrExist)
			},
		},
		{
			name: "Wrap_reflexive",
			run: func(is *assert.Assertions) {
				err := Wrap(fs.ErrExist)
				is.ErrorIs(err, err) //nolint:testifylint
			},
		},
		{
			name: "Wrapf",
			run: func(is *assert.Assertions) {
				err := Wrapf(fs.ErrExist, "Error: %w", assert.AnError)
				is.ErrorIs(err, fs.ErrExist)
			},
		},
		{
			name: "Join_first",
			run: func(is *assert.Assertions) {
				err := Join(fs.ErrExist, assert.AnError)
				is.ErrorIs(err, fs.ErrExist)
			},
		},
		{
			name: "Join_second",
			run: func(is *assert.Assertions) {
				err := Join(assert.AnError, fs.ErrExist)
				is.ErrorIs(err, fs.ErrExist)
			},
		},
		{
			name: "Recover",
			run: func(is *assert.Assertions) {
				err := Recover(func() { panic(fs.ErrExist) })
				is.ErrorIs(err, fs.ErrExist)
			},
		},
		{
			name: "Recoverf",
			run: func(is *assert.Assertions) {
				err := Recoverf(func() { panic(fs.ErrExist) }, "Error: %w", assert.AnError)
				is.ErrorIs(err, fs.ErrExist)
			},
		},
		{
			// Two independently created OopsErrors must not match each other.
			// Previously, Is() returned true for any OopsError target regardless of identity.
			name: "distinct_oops_errors_not_equal",
			run: func(is *assert.Assertions) {
				err1 := Errorf("error 1")
				err2 := Errorf("error 2")
				is.False(errors.Is(err1, err2)) //nolint:testifylint
				is.False(errors.Is(err2, err1)) //nolint:testifylint
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

func TestErrorsAs(t *testing.T) {
	t.Parallel()

	var anError error = &fs.PathError{Err: fs.ErrExist}

	tests := []struct {
		name  string
		errFn func() error
	}{
		{
			name:  "Errorf",
			errFn: func() error { return Errorf("Error: %w", anError) },
		},
		{
			name:  "Wrap",
			errFn: func() error { return Wrap(anError) },
		},
		{
			name:  "Wrapf",
			errFn: func() error { return Wrapf(anError, "Error: %w", assert.AnError) },
		},
		{
			name:  "Join_first",
			errFn: func() error { return Join(anError, assert.AnError) },
		},
		{
			name:  "Join_second",
			errFn: func() error { return Join(assert.AnError, anError) },
		},
		{
			name:  "Recover",
			errFn: func() error { return Recover(func() { panic(anError) }) },
		},
		{
			name:  "Recoverf",
			errFn: func() error { return Recoverf(func() { panic(anError) }, "Error: %w", assert.AnError) },
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			var target *fs.PathError
			is.ErrorAs(tt.errFn(), &target)
		})
	}
}
