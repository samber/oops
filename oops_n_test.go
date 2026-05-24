//nolint:errcheck,forcetypeassert
package oops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "Wrap2",
			run: func(is *assert.Assertions) {
				result, err := Wrap2("test", assert.AnError)
				is.Error(err)
				is.Equal("test", result)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap3",
			run: func(is *assert.Assertions) {
				result1, result2, err := Wrap3("test", "domain", assert.AnError)
				is.Error(err)
				is.Equal("test", result1)
				is.Equal("domain", result2)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap4",
			run: func(is *assert.Assertions) {
				r1, r2, r3, err := Wrap4("test", "domain", "code", assert.AnError)
				is.Error(err)
				is.Equal("test", r1)
				is.Equal("domain", r2)
				is.Equal("code", r3)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap5",
			run: func(is *assert.Assertions) {
				val1, val2, val3, val4, err := Wrap5("test", "domain", "code", "hint", assert.AnError)
				is.Error(err)
				is.Equal("test", val1)
				is.Equal("domain", val2)
				is.Equal("code", val3)
				is.Equal("hint", val4)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap6",
			run: func(is *assert.Assertions) {
				v1, v2, v3, v4, v5, err := Wrap6("test", "domain", "code", "hint", "user", assert.AnError)
				is.Error(err)
				is.Equal("test", v1)
				is.Equal("domain", v2)
				is.Equal("code", v3)
				is.Equal("hint", v4)
				is.Equal("user", v5)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap7",
			run: func(is *assert.Assertions) {
				w1, w2, w3, w4, w5, w6, err := Wrap7("test", "domain", "code", "hint", "user", "tenant", assert.AnError)
				is.Error(err)
				is.Equal("test", w1)
				is.Equal("domain", w2)
				is.Equal("code", w3)
				is.Equal("hint", w4)
				is.Equal("user", w5)
				is.Equal("tenant", w6)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap8",
			run: func(is *assert.Assertions) {
				x1, x2, x3, x4, x5, x6, x7, err := Wrap8("test", "domain", "code", "hint", "user", "tenant", "request", assert.AnError)
				is.Error(err)
				is.Equal("test", x1)
				is.Equal("domain", x2)
				is.Equal("code", x3)
				is.Equal("hint", x4)
				is.Equal("user", x5)
				is.Equal("tenant", x6)
				is.Equal("request", x7)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap9",
			run: func(is *assert.Assertions) {
				y1, y2, y3, y4, y5, y6, y7, y8, err := Wrap9("test", "domain", "code", "hint", "user", "tenant", "request", "response", assert.AnError)
				is.Error(err)
				is.Equal("test", y1)
				is.Equal("domain", y2)
				is.Equal("code", y3)
				is.Equal("hint", y4)
				is.Equal("user", y5)
				is.Equal("tenant", y6)
				is.Equal("request", y7)
				is.Equal("response", y8)
				is.Equal(assert.AnError, err.(OopsError).err)
			},
		},
		{
			name: "Wrap10",
			run: func(is *assert.Assertions) {
				z1, z2, z3, z4, z5, z6, z7, z8, z9, err := Wrap10("test", "domain", "code", "hint", "user", "tenant", "request", "response", "stack", assert.AnError)
				is.Error(err)
				is.Equal("test", z1)
				is.Equal("domain", z2)
				is.Equal("code", z3)
				is.Equal("hint", z4)
				is.Equal("user", z5)
				is.Equal("tenant", z6)
				is.Equal("request", z7)
				is.Equal("response", z8)
				is.Equal("stack", z9)
				is.Equal(assert.AnError, err.(OopsError).err)
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

func TestWrapfFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "Wrapf2",
			run: func(is *assert.Assertions) {
				result, err := Wrapf2("test", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", result)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf3",
			run: func(is *assert.Assertions) {
				result1, result2, err := Wrapf3("test", "domain", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", result1)
				is.Equal("domain", result2)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf4",
			run: func(is *assert.Assertions) {
				r1, r2, r3, err := Wrapf4("test", "domain", "code", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", r1)
				is.Equal("domain", r2)
				is.Equal("code", r3)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf5",
			run: func(is *assert.Assertions) {
				val1, val2, val3, val4, err := Wrapf5("test", "domain", "code", "hint", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", val1)
				is.Equal("domain", val2)
				is.Equal("code", val3)
				is.Equal("hint", val4)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf6",
			run: func(is *assert.Assertions) {
				v1, v2, v3, v4, v5, err := Wrapf6("test", "domain", "code", "hint", "user", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", v1)
				is.Equal("domain", v2)
				is.Equal("code", v3)
				is.Equal("hint", v4)
				is.Equal("user", v5)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf7",
			run: func(is *assert.Assertions) {
				w1, w2, w3, w4, w5, w6, err := Wrapf7("test", "domain", "code", "hint", "user", "tenant", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", w1)
				is.Equal("domain", w2)
				is.Equal("code", w3)
				is.Equal("hint", w4)
				is.Equal("user", w5)
				is.Equal("tenant", w6)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf8",
			run: func(is *assert.Assertions) {
				x1, x2, x3, x4, x5, x6, x7, err := Wrapf8("test", "domain", "code", "hint", "user", "tenant", "request", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", x1)
				is.Equal("domain", x2)
				is.Equal("code", x3)
				is.Equal("hint", x4)
				is.Equal("user", x5)
				is.Equal("tenant", x6)
				is.Equal("request", x7)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf9",
			run: func(is *assert.Assertions) {
				y1, y2, y3, y4, y5, y6, y7, y8, err := Wrapf9("test", "domain", "code", "hint", "user", "tenant", "request", "response", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", y1)
				is.Equal("domain", y2)
				is.Equal("code", y3)
				is.Equal("hint", y4)
				is.Equal("user", y5)
				is.Equal("tenant", y6)
				is.Equal("request", y7)
				is.Equal("response", y8)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
			},
		},
		{
			name: "Wrapf10",
			run: func(is *assert.Assertions) {
				z1, z2, z3, z4, z5, z6, z7, z8, z9, err := Wrapf10("test", "domain", "code", "hint", "user", "tenant", "request", "response", "stack", assert.AnError, "test %d", 42)
				is.Error(err)
				is.Equal("test", z1)
				is.Equal("domain", z2)
				is.Equal("code", z3)
				is.Equal("hint", z4)
				is.Equal("user", z5)
				is.Equal("tenant", z6)
				is.Equal("request", z7)
				is.Equal("response", z8)
				is.Equal("stack", z9)
				is.Equal(assert.AnError, err.(OopsError).err)
				is.Equal("test 42", err.(OopsError).msg)
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
