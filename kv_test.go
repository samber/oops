package oops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDereferencePointers(t *testing.T) {
	is := assert.New(t)

	ptr := func(v string) *string { return &v }

	err := With("hello", "world").Errorf(assert.AnError.Error()).(OopsError)
	is.EqualValues(map[string]any{"hello": "world"}, err.Context())

	err = With("hello", ptr("world")).Errorf(assert.AnError.Error()).(OopsError)
	is.EqualValues(map[string]any{"hello": "world"}, err.Context())

	err = With("hello", nil).Errorf(assert.AnError.Error()).(OopsError)
	is.EqualValues(map[string]any{"hello": nil}, err.Context())
}
