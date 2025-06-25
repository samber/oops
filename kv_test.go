package oops

import (
	"testing"
	"unsafe"

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

func TestDereferencePointerEdgeCases(t *testing.T) {
	is := assert.New(t)

	// Test deeply nested pointers (should not panic)
	deepValue := "deep"
	ptr1 := &deepValue
	ptr2 := &ptr1
	ptr3 := &ptr2
	ptr4 := &ptr3
	ptr5 := &ptr4

	err := With("deep", ptr5).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"deep": "deep"}, err.Context())

	// Test nil pointers at different levels
	var nilPtr *string
	err = With("nil1", nilPtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"nil1": nil}, err.Context())

	var nilPtr2 **string
	err = With("nil2", nilPtr2).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"nil2": nil}, err.Context())

	// Test mixed nil and non-nil pointers
	value := "test"
	valuePtr := &value
	mixedPtr := &valuePtr
	err = With("mixed", mixedPtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"mixed": "test"}, err.Context())

	// Test with different types
	intValue := 42
	intPtr := &intValue
	err = With("int", intPtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"int": 42}, err.Context())

	// Test with struct pointers
	type testStruct struct {
		Field string
	}
	structValue := testStruct{Field: "test"}
	structPtr := &structValue
	err = With("struct", structPtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"struct": testStruct{Field: "test"}}, err.Context())
}

func TestDereferencePointerSafety(t *testing.T) {
	is := assert.New(t)

	// Test with invalid reflect values (should not panic)
	var invalidPtr unsafe.Pointer
	err := With("invalid", invalidPtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	// Should handle gracefully without panic
	is.NotNil(err)

	// Test with function pointers
	testFunc := func() string { return "test" }
	err = With("func", &testFunc).Errorf(anErrorStr).(OopsError) //nolint:govet
	// Should handle function pointers gracefully
	is.NotNil(err)

	// Test with channel pointers
	ch := make(chan int)
	err = With("channel", &ch).Errorf(anErrorStr).(OopsError) //nolint:govet
	// Should handle channel pointers gracefully
	is.NotNil(err)

	// Test with slice pointers
	slice := []int{1, 2, 3}
	err = With("slice", &slice).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"slice": []int{1, 2, 3}}, err.Context())

	// Test with map pointers
	m := map[string]int{"a": 1, "b": 2}
	err = With("map", &m).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"map": map[string]int{"a": 1, "b": 2}}, err.Context())
}

func TestDereferencePointerComplexTypes(t *testing.T) {
	is := assert.New(t)

	// Test with interface pointers
	var iface interface{} = "interface_value"
	ifacePtr := &iface
	err := With("interface", ifacePtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"interface": "interface_value"}, err.Context())

	// Test with array pointers
	arr := [3]int{1, 2, 3}
	arrPtr := &arr
	err = With("array", arrPtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"array": [3]int{1, 2, 3}}, err.Context())

	// Test with nested struct pointers
	type nestedStruct struct {
		Inner struct {
			Value string
		}
	}
	nested := nestedStruct{Inner: struct{ Value string }{Value: "nested"}}
	nestedPtr := &nested
	err = With("nested", nestedPtr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"nested": nested}, err.Context())
}

func TestDereferencePointerWithDisabled(t *testing.T) {
	is := assert.New(t)

	// Save original setting
	original := DereferencePointers
	defer func() { DereferencePointers = original }()

	// Test with dereferencing disabled
	DereferencePointers = false
	value := "test"
	ptr := &value
	err := With("test", ptr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"test": ptr}, err.Context())

	// Re-enable and test
	DereferencePointers = true
	err = With("test", ptr).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.EqualValues(map[string]any{"test": "test"}, err.Context())
}

func TestDereferencePointerMultipleValues(t *testing.T) {
	is := assert.New(t)

	// Test multiple pointer values in the same context
	strValue := "string"
	intValue := 42
	boolValue := true

	err := With(
		"str", &strValue,
		"int", &intValue,
		"bool", &boolValue,
		"nil", (*string)(nil),
	).Errorf(anErrorStr).(OopsError) //nolint:govet

	expected := map[string]any{
		"str":  "string",
		"int":  42,
		"bool": true,
		"nil":  nil,
	}
	is.EqualValues(expected, err.Context())
}

func TestDereferencePointerPanicPrevention(t *testing.T) {
	is := assert.New(t)

	// Test that the function doesn't panic with extremely deep nesting
	// This would have caused a stack overflow in the original implementation
	deepValue := "deep"
	ptr1 := &deepValue
	ptr2 := &ptr1
	ptr3 := &ptr2
	ptr4 := &ptr3
	ptr5 := &ptr4
	ptr6 := &ptr5
	ptr7 := &ptr6
	ptr8 := &ptr7
	ptr9 := &ptr8
	ptr10 := &ptr9
	ptr11 := &ptr10 // This exceeds the 10-level limit

	// This should not panic and should return the original value
	err := With("very_deep", ptr11).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.NotNil(err)

	// The context should contain the pointer as-is since it exceeds the depth limit
	context := err.Context()
	is.Contains(context, "very_deep")

	// Test with invalid reflect values that could cause panics
	var invalidValue interface{} = func() {}                            // Function type
	err = With("invalid", &invalidValue).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.NotNil(err)

	// Test with unexported fields that might cause panics when accessing
	type unexportedStruct struct {
		unexported string
	}
	unexported := unexportedStruct{unexported: "test"}
	err = With("unexported", &unexported).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.NotNil(err)

	// Test with nil interface
	var nilInterface interface{} = nil
	err = With("nil_interface", &nilInterface).Errorf(anErrorStr).(OopsError) //nolint:govet
	is.NotNil(err)
}
