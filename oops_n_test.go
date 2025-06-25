package oops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapFunctions(t *testing.T) {
	is := assert.New(t)

	// Test Wrap2
	result, err := Wrap2("test", assert.AnError)
	is.Error(err)
	is.Equal("test", result)
	is.Equal(assert.AnError, err.(OopsError).err)

	// Test Wrap3
	result1, result2, err2 := Wrap3("test", "domain", assert.AnError)
	is.Error(err2)
	is.Equal("test", result1)
	is.Equal("domain", result2)
	is.Equal(assert.AnError, err2.(OopsError).err)

	// Test Wrap4
	r1, r2, r3, err3 := Wrap4("test", "domain", "code", assert.AnError)
	is.Error(err3)
	is.Equal("test", r1)
	is.Equal("domain", r2)
	is.Equal("code", r3)
	is.Equal(assert.AnError, err3.(OopsError).err)

	// Test Wrap5
	val1, val2, val3, val4, err4 := Wrap5("test", "domain", "code", "hint", assert.AnError)
	is.Error(err4)
	is.Equal("test", val1)
	is.Equal("domain", val2)
	is.Equal("code", val3)
	is.Equal("hint", val4)
	is.Equal(assert.AnError, err4.(OopsError).err)

	// Test Wrap6
	v1, v2, v3, v4, v5, err5 := Wrap6("test", "domain", "code", "hint", "user", assert.AnError)
	is.Error(err5)
	is.Equal("test", v1)
	is.Equal("domain", v2)
	is.Equal("code", v3)
	is.Equal("hint", v4)
	is.Equal("user", v5)
	is.Equal(assert.AnError, err5.(OopsError).err)

	// Test Wrap7
	w1, w2, w3, w4, w5, w6, err6 := Wrap7("test", "domain", "code", "hint", "user", "tenant", assert.AnError)
	is.Error(err6)
	is.Equal("test", w1)
	is.Equal("domain", w2)
	is.Equal("code", w3)
	is.Equal("hint", w4)
	is.Equal("user", w5)
	is.Equal("tenant", w6)
	is.Equal(assert.AnError, err6.(OopsError).err)

	// Test Wrap8
	x1, x2, x3, x4, x5, x6, x7, err7 := Wrap8("test", "domain", "code", "hint", "user", "tenant", "request", assert.AnError)
	is.Error(err7)
	is.Equal("test", x1)
	is.Equal("domain", x2)
	is.Equal("code", x3)
	is.Equal("hint", x4)
	is.Equal("user", x5)
	is.Equal("tenant", x6)
	is.Equal("request", x7)
	is.Equal(assert.AnError, err7.(OopsError).err)

	// Test Wrap9
	y1, y2, y3, y4, y5, y6, y7, y8, err8 := Wrap9("test", "domain", "code", "hint", "user", "tenant", "request", "response", assert.AnError)
	is.Error(err8)
	is.Equal("test", y1)
	is.Equal("domain", y2)
	is.Equal("code", y3)
	is.Equal("hint", y4)
	is.Equal("user", y5)
	is.Equal("tenant", y6)
	is.Equal("request", y7)
	is.Equal("response", y8)
	is.Equal(assert.AnError, err8.(OopsError).err)

	// Test Wrap10
	z1, z2, z3, z4, z5, z6, z7, z8, z9, err9 := Wrap10("test", "domain", "code", "hint", "user", "tenant", "request", "response", "stack", assert.AnError)
	is.Error(err9)
	is.Equal("test", z1)
	is.Equal("domain", z2)
	is.Equal("code", z3)
	is.Equal("hint", z4)
	is.Equal("user", z5)
	is.Equal("tenant", z6)
	is.Equal("request", z7)
	is.Equal("response", z8)
	is.Equal("stack", z9)
	is.Equal(assert.AnError, err9.(OopsError).err)
}

func TestWrapfFunctions(t *testing.T) {
	is := assert.New(t)

	// Test Wrapf2
	result, err := Wrapf2("test", assert.AnError, "test %d", 42)
	is.Error(err)
	is.Equal("test", result)
	is.Equal(assert.AnError, err.(OopsError).err)
	is.Equal("test 42", err.(OopsError).msg)

	// Test Wrapf3
	result1, result2, err2 := Wrapf3("test", "domain", assert.AnError, "test %d", 42)
	is.Error(err2)
	is.Equal("test", result1)
	is.Equal("domain", result2)
	is.Equal(assert.AnError, err2.(OopsError).err)
	is.Equal("test 42", err2.(OopsError).msg)

	// Test Wrapf4
	r1, r2, r3, err3 := Wrapf4("test", "domain", "code", assert.AnError, "test %d", 42)
	is.Error(err3)
	is.Equal("test", r1)
	is.Equal("domain", r2)
	is.Equal("code", r3)
	is.Equal(assert.AnError, err3.(OopsError).err)
	is.Equal("test 42", err3.(OopsError).msg)

	// Test Wrapf5
	val1, val2, val3, val4, err4 := Wrapf5("test", "domain", "code", "hint", assert.AnError, "test %d", 42)
	is.Error(err4)
	is.Equal("test", val1)
	is.Equal("domain", val2)
	is.Equal("code", val3)
	is.Equal("hint", val4)
	is.Equal(assert.AnError, err4.(OopsError).err)
	is.Equal("test 42", err4.(OopsError).msg)

	// Test Wrapf6
	v1, v2, v3, v4, v5, err5 := Wrapf6("test", "domain", "code", "hint", "user", assert.AnError, "test %d", 42)
	is.Error(err5)
	is.Equal("test", v1)
	is.Equal("domain", v2)
	is.Equal("code", v3)
	is.Equal("hint", v4)
	is.Equal("user", v5)
	is.Equal(assert.AnError, err5.(OopsError).err)
	is.Equal("test 42", err5.(OopsError).msg)

	// Test Wrapf7
	w1, w2, w3, w4, w5, w6, err6 := Wrapf7("test", "domain", "code", "hint", "user", "tenant", assert.AnError, "test %d", 42)
	is.Error(err6)
	is.Equal("test", w1)
	is.Equal("domain", w2)
	is.Equal("code", w3)
	is.Equal("hint", w4)
	is.Equal("user", w5)
	is.Equal("tenant", w6)
	is.Equal(assert.AnError, err6.(OopsError).err)
	is.Equal("test 42", err6.(OopsError).msg)

	// Test Wrapf8
	x1, x2, x3, x4, x5, x6, x7, err7 := Wrapf8("test", "domain", "code", "hint", "user", "tenant", "request", assert.AnError, "test %d", 42)
	is.Error(err7)
	is.Equal("test", x1)
	is.Equal("domain", x2)
	is.Equal("code", x3)
	is.Equal("hint", x4)
	is.Equal("user", x5)
	is.Equal("tenant", x6)
	is.Equal("request", x7)
	is.Equal(assert.AnError, err7.(OopsError).err)
	is.Equal("test 42", err7.(OopsError).msg)

	// Test Wrapf9
	y1, y2, y3, y4, y5, y6, y7, y8, err8 := Wrapf9("test", "domain", "code", "hint", "user", "tenant", "request", "response", assert.AnError, "test %d", 42)
	is.Error(err8)
	is.Equal("test", y1)
	is.Equal("domain", y2)
	is.Equal("code", y3)
	is.Equal("hint", y4)
	is.Equal("user", y5)
	is.Equal("tenant", y6)
	is.Equal("request", y7)
	is.Equal("response", y8)
	is.Equal(assert.AnError, err8.(OopsError).err)
	is.Equal("test 42", err8.(OopsError).msg)

	// Test Wrapf10
	z1, z2, z3, z4, z5, z6, z7, z8, z9, err9 := Wrapf10("test", "domain", "code", "hint", "user", "tenant", "request", "response", "stack", assert.AnError, "test %d", 42)
	is.Error(err9)
	is.Equal("test", z1)
	is.Equal("domain", z2)
	is.Equal("code", z3)
	is.Equal("hint", z4)
	is.Equal("user", z5)
	is.Equal("tenant", z6)
	is.Equal("request", z7)
	is.Equal("response", z8)
	is.Equal("stack", z9)
	is.Equal(assert.AnError, err9.(OopsError).err)
	is.Equal("test 42", err9.(OopsError).msg)
}
