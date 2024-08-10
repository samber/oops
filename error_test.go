package oops

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsIs(t *testing.T) {
	is := assert.New(t)

	err := Errorf("Error: %w", fs.ErrExist)
	is.True(errors.Is(err, fs.ErrExist))

	err = Wrap(fs.ErrExist)
	is.True(errors.Is(err, fs.ErrExist))

	err = Wrapf(fs.ErrExist, "Error: %w", assert.AnError)
	is.True(errors.Is(err, fs.ErrExist))

	err = Join(fs.ErrExist, assert.AnError)
	is.True(errors.Is(err, fs.ErrExist))
	err = Join(assert.AnError, fs.ErrExist)
	is.True(errors.Is(err, fs.ErrExist))

	err = Recover(func() {
		panic(fs.ErrExist)
	})
	is.True(errors.Is(err, fs.ErrExist))

	err = Recoverf(func() {
		panic(fs.ErrExist)
	}, "Error: %w", assert.AnError)
	is.True(errors.Is(err, fs.ErrExist))
}

func TestErrorsAs(t *testing.T) {
	is := assert.New(t)

	var anError error = &fs.PathError{Err: fs.ErrExist}
	var target *fs.PathError

	err := Errorf("error: %w", anError)
	is.True(errors.As(err, &target))

	err = Wrap(anError)
	is.True(errors.As(err, &target))

	err = Wrapf(anError, "Error: %w", assert.AnError)
	is.True(errors.As(err, &target))

	err = Join(anError, assert.AnError)
	is.True(errors.As(err, &target))
	err = Join(assert.AnError, anError)
	is.True(errors.As(err, &target))

	err = Recover(func() {
		panic(anError)
	})
	is.True(errors.As(err, &target))

	err = Recoverf(func() {
		panic(anError)
	}, "Error: %w", assert.AnError)
	is.True(errors.As(err, &target))
}
