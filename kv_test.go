package oops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const anErrorStr = "assert.AnError general error for testing"

func TestDereferencePointers(t *testing.T) {
	is := assert.New(t)

	ptr := func(v string) *string { return &v }

	err := With("hello", "world").Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"hello": "world"}, err.Context())

	err = With("hello", ptr("world")).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"hello": "world"}, err.Context())

	err = With("hello", nil).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"hello": nil}, err.Context())

	err = With("hello", (*int)(nil)).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"hello": nil}, err.Context())

	err = With("hello", (***int)(nil)).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"hello": nil}, err.Context())

	var i **int
	err = With("hello", (***int)(&i)).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"hello": nil}, err.Context())
}
