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

	// Two independently created OopsErrors must not match each other.
	// Previously, Is() returned true for any OopsError target regardless of identity.
	err1 := Errorf("error 1")
	err2 := Errorf("error 2")
	is.False(errors.Is(err1, err2)) //nolint:testifylint
	is.False(errors.Is(err2, err1)) //nolint:testifylint
}

func TestLayers(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// single layer
	oopsErr, ok := AsOops(New("hello"))
	is.True(ok)
	layers := oopsErr.Layers()
	is.Len(layers, 1)

	// multiple layers with distinct attributes
	inner := Code("inner_code").Public("inner public").New("inner")
	outer := Code("outer_code").Public("outer public").Wrap(inner)
	oopsErr, ok = AsOops(outer)
	is.True(ok)
	layers = oopsErr.Layers()
	is.Len(layers, 2)
	is.Equal("outer_code", layers[0].Code)
	is.Equal("outer public", layers[0].Public)
	is.Equal("inner_code", layers[1].Code)
	is.Equal("inner public", layers[1].Public)

	// non-OopsError root is skipped
	root := errors.New("plain error")
	wrapped := Wrap(root)
	oopsErr, ok = AsOops(wrapped)
	is.True(ok)
	layers = oopsErr.Layers()
	is.Len(layers, 1)

	// three layers deep
	l1 := Code("l1").New("level 1")
	l2 := Code("l2").Wrap(l1)
	l3 := Code("l3").Wrap(l2)
	oopsErr, ok = AsOops(l3)
	is.True(ok)
	layers = oopsErr.Layers()
	is.Len(layers, 3)
	is.Equal("l3", layers[0].Code)
	is.Equal("l2", layers[1].Code)
	is.Equal("l1", layers[2].Code)

	// layers are pointers (not shared)
	is.NotSame(layers[0], layers[1])
	is.NotSame(layers[1], layers[2])
}

func TestErrorsAs(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	var anError error = &fs.PathError{Err: fs.ErrExist}
	var target *fs.PathError

	err := Errorf("Error: %w", anError)
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
