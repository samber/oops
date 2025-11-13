package oops

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsIs(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	err := Errorf("Error: %w", fs.ErrExist)
	is.ErrorIs(err, fs.ErrExist)

	err = Wrap(fs.ErrExist)
	is.ErrorIs(err, fs.ErrExist)

	err = Wrap(fs.ErrExist)
	is.ErrorIs(err, err) //nolint:testifylint

	err = Wrapf(fs.ErrExist, "Error: %w", assert.AnError)
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
	}, "Error: %w", assert.AnError)
	is.ErrorIs(err, fs.ErrExist)
}

func TestErrorsAs(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	var anError error = &fs.PathError{Err: fs.ErrExist}
	var target *fs.PathError

	err := Errorf("error: %w", anError)
	is.ErrorAs(err, &target)

	err = Wrap(anError)
	is.ErrorAs(err, &target)

	err = Wrapf(anError, "Error: %w", assert.AnError)
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
	}, "Error: %w", assert.AnError)
	is.ErrorAs(err, &target)
}
