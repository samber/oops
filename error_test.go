package oops

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsIs(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Use a variable for the format string to avoid go vet false-positive:
	// oops.Errorf and oops.Wrapf support %w but vet cannot verify this
	// through the conditional fmt.Errorf path used for performance.
	wrapFormat := "Error: %w"

	err := Errorf(wrapFormat, fs.ErrExist)
	is.ErrorIs(err, fs.ErrExist)

	err = Wrap(fs.ErrExist)
	is.ErrorIs(err, fs.ErrExist)

	err = Wrap(fs.ErrExist)
	is.ErrorIs(err, err) //nolint:testifylint

	err = Wrapf(fs.ErrExist, wrapFormat, assert.AnError)
	is.ErrorIs(err, fs.ErrExist)

	err = Join(fs.ErrExist, assert.AnError)
	is.ErrorIs(err, fs.ErrExist)
	err = Join(assert.AnError, fs.ErrExist)
	is.ErrorIs(err, fs.ErrExist)

	err = Recover(func() {
		panic(fs.ErrExist)
	})
	is.ErrorIs(err, fs.ErrExist)

	err = Recoverf(func() {
		panic(fs.ErrExist)
	}, wrapFormat, assert.AnError)
	is.ErrorIs(err, fs.ErrExist)

	// Two independently created OopsErrors must not match each other.
	// Previously, Is() returned true for any OopsError target regardless of identity.
	err1 := Errorf("error 1")
	err2 := Errorf("error 2")
	is.False(errors.Is(err1, err2)) //nolint:testifylint
	is.False(errors.Is(err2, err1)) //nolint:testifylint
}

func TestErrorsAs(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	var anError error = &fs.PathError{Err: fs.ErrExist}
	var target *fs.PathError

	// Use a variable for the format string to avoid go vet false-positive:
	// oops.Errorf and oops.Wrapf support %w but vet cannot verify this
	// through the conditional fmt.Errorf path used for performance.
	wrapFormat := "Error: %w"

	err := Errorf(wrapFormat, anError)
	is.ErrorAs(err, &target)

	err = Wrap(anError)
	is.ErrorAs(err, &target)

	err = Wrapf(anError, wrapFormat, assert.AnError)
	is.ErrorAs(err, &target)

	err = Join(anError, assert.AnError)
	is.ErrorAs(err, &target)
	err = Join(assert.AnError, anError)
	is.ErrorAs(err, &target)

	err = Recover(func() {
		panic(anError)
	})
	is.ErrorAs(err, &target)

	err = Recoverf(func() {
		panic(anError)
	}, wrapFormat, assert.AnError)
	is.ErrorAs(err, &target)
}
