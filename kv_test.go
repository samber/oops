//nolint:errcheck,forcetypeassert
package oops

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const anErrorStr = "assert.AnError general error for testing"

func TestDereferencePointers(t *testing.T) {
	t.Parallel()

	ptr := func(v string) *string { return &v }

	tests := []struct {
		name        string
		makeVal     func() any
		wantCtx     map[string]any
		equalValues bool
	}{
		{
			name:    "string_value",
			makeVal: func() any { return "world" },
			wantCtx: map[string]any{"hello": "world"},
		},
		{
			name:    "string_pointer",
			makeVal: func() any { return ptr("world") },
			wantCtx: map[string]any{"hello": "world"},
		},
		{
			name:    "nil_value",
			makeVal: func() any { return nil },
			wantCtx: map[string]any{"hello": nil},
		},
		{
			name:    "nil_int_pointer",
			makeVal: func() any { return (*int)(nil) },
			wantCtx: map[string]any{"hello": nil},
		},
		{
			name:    "nil_triple_pointer",
			makeVal: func() any { return (***int)(nil) },
			wantCtx: map[string]any{"hello": nil},
		},
		{
			name: "nil_through_triple_pointer",
			makeVal: func() any {
				var i **int
				return (***int)(&i) //nolint:unconvert
			},
			wantCtx:     map[string]any{"hello": nil},
			equalValues: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			err := With("hello", tt.makeVal()).Errorf(anErrorStr).(OopsError)
			if tt.equalValues {
				is.EqualValues(tt.wantCtx, err.Context()) //nolint:testifylint
			} else {
				is.Equal(tt.wantCtx, err.Context())
			}
		})
	}
}

func TestDereferencePointerEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "deeply_nested_pointers",
			run: func(is *assert.Assertions) {
				deepValue := "deep"
				ptr1 := &deepValue
				ptr2 := &ptr1
				ptr3 := &ptr2
				ptr4 := &ptr3
				ptr5 := &ptr4
				err := With("deep", ptr5).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"deep": "deep"}, err.Context())
			},
		},
		{
			name: "nil_single_pointer",
			run: func(is *assert.Assertions) {
				var nilPtr *string
				err := With("nil1", nilPtr).Errorf(anErrorStr).(OopsError)
				is.EqualValues(map[string]any{"nil1": nil}, err.Context()) //nolint:testifylint
			},
		},
		{
			name: "nil_double_pointer",
			run: func(is *assert.Assertions) {
				var nilPtr2 **string
				err := With("nil2", nilPtr2).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"nil2": nil}, err.Context())
			},
		},
		{
			name: "mixed_nil_and_non_nil",
			run: func(is *assert.Assertions) {
				value := "test"
				valuePtr := &value
				mixedPtr := &valuePtr
				err := With("mixed", mixedPtr).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"mixed": "test"}, err.Context())
			},
		},
		{
			name: "int_pointer",
			run: func(is *assert.Assertions) {
				intValue := 42
				intPtr := &intValue
				err := With("int", intPtr).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"int": 42}, err.Context())
			},
		},
		{
			name: "struct_pointer",
			run: func(is *assert.Assertions) {
				type testStruct struct{ Field string }
				structValue := testStruct{Field: "test"}
				structPtr := &structValue
				err := With("struct", structPtr).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"struct": testStruct{Field: "test"}}, err.Context())
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

func TestDereferencePointerSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "invalid_pointer",
			run: func(is *assert.Assertions) {
				var invalidPtr unsafe.Pointer
				err := With("invalid", invalidPtr).Errorf(anErrorStr).(OopsError)
				is.NotNil(err)
			},
		},
		{
			name: "function_pointer",
			run: func(is *assert.Assertions) {
				testFunc := func() string { return "test" }
				err := With("func", &testFunc).Errorf(anErrorStr).(OopsError)
				is.NotNil(err)
			},
		},
		{
			name: "channel_pointer",
			run: func(is *assert.Assertions) {
				ch := make(chan int)
				err := With("channel", &ch).Errorf(anErrorStr).(OopsError)
				is.NotNil(err)
			},
		},
		{
			name: "slice_pointer",
			run: func(is *assert.Assertions) {
				slice := []int{1, 2, 3}
				err := With("slice", &slice).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"slice": []int{1, 2, 3}}, err.Context())
			},
		},
		{
			name: "map_pointer",
			run: func(is *assert.Assertions) {
				m := map[string]int{"a": 1, "b": 2}
				err := With("map", &m).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"map": map[string]int{"a": 1, "b": 2}}, err.Context())
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

func TestDereferencePointerComplexTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "interface_pointer",
			run: func(is *assert.Assertions) {
				var iface any = "interface_value"
				ifacePtr := &iface
				err := With("interface", ifacePtr).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"interface": "interface_value"}, err.Context())
			},
		},
		{
			name: "array_pointer",
			run: func(is *assert.Assertions) {
				arr := [3]int{1, 2, 3}
				arrPtr := &arr
				err := With("array", arrPtr).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"array": [3]int{1, 2, 3}}, err.Context())
			},
		},
		{
			name: "nested_struct_pointer",
			run: func(is *assert.Assertions) {
				type nestedStruct struct {
					Inner struct{ Value string }
				}
				nested := nestedStruct{Inner: struct{ Value string }{Value: "nested"}}
				nestedPtr := &nested
				err := With("nested", nestedPtr).Errorf(anErrorStr).(OopsError)
				is.Equal(map[string]any{"nested": nested}, err.Context())
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

func TestDereferencePointerWithDisabled(t *testing.T) { //nolint:paralleltest
	is := assert.New(t)

	// Save original setting
	original := DereferencePointers
	defer func() { DereferencePointers = original }()

	// Test with dereferencing disabled
	DereferencePointers = false
	value := "test"
	ptr := &value
	err := With("test", ptr).Errorf(anErrorStr).(OopsError)
	is.Equal(map[string]any{"test": ptr}, err.Context())

	// Re-enable and test
	DereferencePointers = true
	err = With("test", ptr).Errorf(anErrorStr).(OopsError)
	is.Equal(map[string]any{"test": "test"}, err.Context())
}

func TestDereferencePointerMultipleValues(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test multiple pointer values in the same context
	strValue := "string"
	intValue := 42
	boolValue := true

	err := With(
		"str", &strValue,
		"int", &intValue,
		"bool", &boolValue,
		"nil", (*string)(nil),
	).Errorf(anErrorStr).(OopsError)

	expected := map[string]any{
		"str":  "string",
		"int":  42,
		"bool": true,
		"nil":  nil,
	}
	is.Equal(expected, err.Context())
}

func TestDereferencePointerPanicPrevention(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			// This would have caused a stack overflow in the original implementation.
			name: "deeply_nested_exceeds_limit",
			run: func(is *assert.Assertions) {
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
				ptr11 := &ptr10 // exceeds the 10-level limit
				err := With("very_deep", ptr11).Errorf(anErrorStr).(OopsError)
				is.NotNil(err)
				is.Contains(err.Context(), "very_deep")
			},
		},
		{
			name: "function_type",
			run: func(is *assert.Assertions) {
				var invalidValue any = func() {}
				err := With("invalid", &invalidValue).Errorf(anErrorStr).(OopsError)
				is.NotNil(err)
			},
		},
		{
			name: "unexported_struct_fields",
			run: func(is *assert.Assertions) {
				type unexportedStruct struct{ unexported string }
				unexported := unexportedStruct{unexported: "test"}
				err := With("unexported", &unexported).Errorf(anErrorStr).(OopsError)
				is.NotNil(err)
			},
		},
		{
			name: "nil_interface",
			run: func(is *assert.Assertions) {
				var nilInterface any
				err := With("nil_interface", &nilInterface).Errorf(anErrorStr).(OopsError)
				is.NotNil(err)
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

func TestDereferencePointerRecursiveEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "nil_pointer",
			run: func(is *assert.Assertions) {
				var nilPtr *string
				result := dereferencePointerRecursive(reflect.ValueOf(nilPtr), 0)
				is.Nil(result)
			},
		},
		{
			name: "max_depth_exceeded",
			run: func(is *assert.Assertions) {
				value := "test"
				ptr := &value
				result := dereferencePointerRecursive(reflect.ValueOf(ptr), 100)
				is.Equal(ptr, result)
			},
		},
		{
			name: "invalid_reflect_value",
			run: func(is *assert.Assertions) {
				var invalid reflect.Value
				result := dereferencePointerRecursive(invalid, 0)
				is.Nil(result)
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

func TestLazyMapEvaluation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]any
		want  map[string]any
	}{
		{
			name:  "static values pass through unchanged",
			input: map[string]any{"s": "hello", "n": 42, "b": true},
			want:  map[string]any{"s": "hello", "n": 42, "b": true},
		},
		{
			name:  "zero-arg single-return func is evaluated",
			input: map[string]any{"fn": func() any { return "computed" }},
			want:  map[string]any{"fn": "computed"},
		},
		{
			name: "nested map is recursively evaluated",
			input: map[string]any{
				"outer": map[string]any{
					"fn": func() any { return 99 },
				},
			},
			want: map[string]any{
				"outer": map[string]any{"fn": 99},
			},
		},
		{
			name:  "func with wrong signature is left unevaluated",
			input: map[string]any{"fn": func() (string, error) { return "x", nil }},
			want:  map[string]any{"fn": func() (string, error) { return "x", nil }},
		},
		{
			name:  "nil value is preserved",
			input: map[string]any{"n": nil},
			want:  map[string]any{"n": nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			got := lazyMapEvaluation(tt.input)
			is.Len(got, len(tt.want))
			// Compare func entries by kind only (functions aren't directly comparable).
			for k, wantVal := range tt.want {
				is.Contains(got, k, "key %q missing from result", k)
				if wantVal != nil && reflect.TypeOf(wantVal).Kind() == reflect.Func {
					is.Equal(reflect.Func, reflect.TypeOf(got[k]).Kind(), "key %q", k)
				} else {
					is.Equal(wantVal, got[k], "key %q", k)
				}
			}
		})
	}
}

func TestLazyValueEvaluationEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "nil_value",
			run: func(is *assert.Assertions) {
				result := lazyValueEvaluation(nil)
				is.Nil(result)
			},
		},
		{
			name: "non_function_value",
			run: func(is *assert.Assertions) {
				result := lazyValueEvaluation("string")
				is.Equal("string", result)
			},
		},
		{
			name: "function_returning_error",
			run: func(is *assert.Assertions) {
				fn := func() (string, error) { return "test", assert.AnError }
				result := lazyValueEvaluation(fn)
				is.Equal(reflect.Func, reflect.TypeOf(result).Kind())
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

func TestRecursive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         OopsError
		stopAfter   int // 0 = never stop early; N = stop after N visits
		wantVisited []string
	}{
		{
			name:        "single error visits once",
			err:         Code("single").Errorf("single").(OopsError),
			wantVisited: []string{"single"},
		},
		{
			name:        "chain of three visited outer to inner",
			err:         Code("outer").Wrapf(Code("mid").Wrapf(Code("inner").Errorf("inner"), "mid"), "outer").(OopsError),
			wantVisited: []string{"outer", "mid", "inner"},
		},
		{
			name:        "early stop halts traversal after first visit",
			err:         Code("outer").Wrapf(Code("mid").Wrapf(Code("inner").Errorf("inner"), "mid"), "outer").(OopsError),
			stopAfter:   1,
			wantVisited: []string{"outer"},
		},
		{
			name:        "non-oops wrapped error stops at boundary",
			err:         Code("oops").Wrapf(errors.New("plain"), "oops").(OopsError),
			wantVisited: []string{"oops"},
		},
		// Mixed: OopsError -> plain -> OopsError
		// fmt.Errorf is transparent to errors.As, so the inner OopsError is still reached.
		{
			name: "oops wrapping fmt.Errorf wrapping oops visits both oops layers",
			err: func() OopsError {
				inner := Code("inner").Errorf("inner")
				mid := fmt.Errorf("mid: %w", inner) // plain wrapper, transparent to errors.As
				return Code("outer").Wrapf(mid, "outer").(OopsError)
			}(),
			wantVisited: []string{"outer", "inner"},
		},
		// Mixed: OopsError -> plain -> OopsError -> plain
		// The second plain boundary stops traversal after the second OopsError.
		{
			name: "oops wrapping fmt.Errorf wrapping oops wrapping plain visits two oops layers",
			err: func() OopsError {
				inner := Code("inner").Wrapf(errors.New("plain"), "inner")
				mid := fmt.Errorf("mid: %w", inner) // transparent
				return Code("outer").Wrapf(mid, "outer").(OopsError)
			}(),
			wantVisited: []string{"outer", "inner"},
		},
		// Mixed: OopsError -> fmt.Errorf -> plain (no inner OopsError)
		// errors.As finds no OopsError through the plain chain, so traversal stops at outer.
		{
			name: "oops wrapping fmt.Errorf wrapping plain stops at boundary",
			err: func() OopsError {
				mid := fmt.Errorf("mid: %w", errors.New("plain"))
				return Code("outer").Wrapf(mid, "outer").(OopsError)
			}(),
			wantVisited: []string{"outer"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			var (
				visited []string
				count   int
			)
			recursive(tt.err, func(e OopsError) bool {
				visited = append(visited, e.code.(string))
				count++
				return tt.stopAfter == 0 || count < tt.stopAfter
			})
			is.Equal(tt.wantVisited, visited)
		})
	}
}

func TestGetDeepestErrorAttribute(t *testing.T) {
	t.Parallel()

	getter := func(e OopsError) any { return e.code }

	tests := []struct {
		name string
		err  func() OopsError
		want any
	}{
		{
			name: "single error returns its own code",
			err:  func() OopsError { return Code("single").Errorf("single").(OopsError) },
			want: "single",
		},
		{
			name: "deepest code wins when only inner is set",
			err: func() OopsError {
				return Wrapf(Code("inner").Errorf("inner"), "outer").(OopsError)
			},
			want: "inner",
		},
		{
			name: "outer code used as fallback when inner has none",
			err: func() OopsError {
				return Code("outer").Wrap(Errorf("inner")).(OopsError)
			},
			want: "outer",
		},
		{
			name: "deepest code wins when both levels are set",
			err: func() OopsError {
				return Code("outer").Wrap(Code("inner").Errorf("inner")).(OopsError)
			},
			want: "inner",
		},
		{
			name: "outer code used as fallback when inner is plain error",
			err: func() OopsError {
				return Code("fallback").Wrap(errors.New("plain")).(OopsError)
			},
			want: "fallback",
		},
		// Zero value
		{
			name: "no code set anywhere returns nil",
			err:  func() OopsError { return Wrap(Errorf("inner")).(OopsError) },
			want: nil,
		},
		// Three-level chains
		{
			name: "three-level: only deepest has code",
			err: func() OopsError {
				return Wrapf(Wrap(Code("deep").Errorf("inner")), "outer").(OopsError)
			},
			want: "deep",
		},
		{
			name: "three-level: only middle has code",
			err: func() OopsError {
				return Wrapf(Code("mid").Wrap(Errorf("inner")), "outer").(OopsError)
			},
			want: "mid",
		},
		{
			name: "three-level: only outer has code",
			err: func() OopsError {
				return Code("outer").Wrapf(Wrap(Errorf("inner")), "mid").(OopsError)
			},
			want: "outer",
		},
		{
			name: "three-level: all set returns deepest",
			err: func() OopsError {
				return Code("outer").Wrapf(Code("mid").Wrap(Code("inner").Errorf("inner")), "mid").(OopsError)
			},
			want: "inner",
		},
		// Non-string comparable type
		{
			name: "integer code is returned correctly",
			err:  func() OopsError { return Code(42).Errorf("err").(OopsError) },
			want: 42,
		},
		// Mixed: OopsError -> plain -> OopsError
		// fmt.Errorf is transparent to errors.As, so the inner OopsError is found.
		{
			name: "plain wrapper between two oops layers is transparent: deepest code wins",
			err: func() OopsError {
				inner := Code("inner").Errorf("inner")
				mid := fmt.Errorf("mid: %w", inner)
				return Code("outer").Wrapf(mid, "outer").(OopsError)
			},
			want: "inner",
		},
		{
			name: "plain wrapper between two oops layers is transparent: falls back to outer when inner has no code",
			err: func() OopsError {
				inner := Errorf("inner")
				mid := fmt.Errorf("mid: %w", inner)
				return Code("outer").Wrapf(mid, "outer").(OopsError)
			},
			want: "outer",
		},
		// Mixed: OopsError -> plain -> OopsError -> plain
		// Second plain stops recursion; the middle OopsError code is the deepest reachable.
		{
			name: "oops wrapping plain wrapping oops wrapping plain: middle oops code is deepest",
			err: func() OopsError {
				inner := Code("mid").Wrapf(errors.New("plain"), "mid")
				outer_plain := fmt.Errorf("wrap: %w", inner)
				return Code("outer").Wrapf(outer_plain, "outer").(OopsError)
			},
			want: "mid",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			is.Equal(tt.want, getDeepestErrorAttribute(tt.err(), getter))
		})
	}
}

func TestMergeNestedErrorMap(t *testing.T) {
	t.Parallel()

	getter := func(e OopsError) map[string]any { return e.context }

	tests := []struct {
		name string
		err  func() OopsError
		want map[string]any
	}{
		{
			name: "single error returns its own context",
			err:  func() OopsError { return With("a", 1).Errorf("err").(OopsError) },
			want: map[string]any{"a": 1},
		},
		{
			name: "deeper key overrides shallower key in chain",
			err: func() OopsError {
				inner := With("a", "inner", "b", 2).Errorf("inner")
				return With("a", "outer", "c", 3).Wrap(inner).(OopsError)
			},
			want: map[string]any{"a": "inner", "b": 2, "c": 3},
		},
		{
			name: "error with no context returns empty map",
			err:  func() OopsError { return Errorf("no context").(OopsError) },
			want: map[string]any{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			is.Equal(tt.want, mergeNestedErrorMap(tt.err(), getter))
		})
	}
}

func TestCollectMaps(t *testing.T) {
	t.Parallel()

	getter := func(e OopsError) map[string]any { return e.context }

	tests := []struct {
		name     string
		err      func() OopsError
		wantMaps []map[string]any
	}{
		{
			name:     "single error with context appends one map",
			err:      func() OopsError { return With("x", 1).Errorf("err").(OopsError) },
			wantMaps: []map[string]any{{"x": 1}},
		},
		{
			name: "chain appends maps in shallow to deep order",
			err: func() OopsError {
				return With("shallow", false).Wrap(With("deep", true).Errorf("inner")).(OopsError)
			},
			wantMaps: []map[string]any{{"shallow": false}, {"deep": true}},
		},
		{
			name:     "error with no context appends nothing",
			err:      func() OopsError { return Errorf("no context").(OopsError) },
			wantMaps: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			var result []map[string]any
			collectMaps(tt.err(), getter, &result)
			is.Equal(tt.wantMaps, result)
		})
	}
}

func TestSnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(is *assert.Assertions)
	}{
		{
			name: "single_error_scalar_fields_preserved",
			run: func(is *assert.Assertions) {
				now := time.Now().Truncate(time.Second)
				dur := 5 * time.Second
				err := Code("E001").
					Time(now).
					Duration(dur).
					In("payments").
					Hint("check config").
					Public("payment failed").
					Owner("team-payments").
					With("key", "val").
					Tags("tag1", "tag2").
					Errorf("base error").(OopsError)
				s := snapshot(err)
				is.Equal("E001", s.code)
				is.Equal(now, s.time)
				is.Equal(dur, s.duration)
				is.Equal("payments", s.domain)
				is.Equal("check config", s.hint)
				is.Equal("payment failed", s.public)
				is.Equal("team-payments", s.owner)
				is.Equal(map[string]any{"key": "val"}, s.context)
				is.ElementsMatch([]string{"tag1", "tag2"}, s.tags)
			},
		},
		{
			name: "chained_innermost_scalar_wins",
			run: func(is *assert.Assertions) {
				inner := Code("INNER").In("inner-domain").Hint("inner-hint").Errorf("inner")
				outer := Code("OUTER").In("outer-domain").Hint("outer-hint").Wrap(inner).(OopsError)
				s := snapshot(outer)
				is.Equal("INNER", s.code)
				is.Equal("inner-domain", s.domain)
				is.Equal("inner-hint", s.hint)
			},
		},
		{
			name: "chained_inner_code_with_outer_fallback",
			run: func(is *assert.Assertions) {
				inner := Errorf("inner") // no code
				outer := Code("OUTER").Wrap(inner).(OopsError)
				s := snapshot(outer)
				is.Equal("OUTER", s.code)
			},
		},
		{
			name: "context_inner_key_wins_outer_key_merged",
			run: func(is *assert.Assertions) {
				inner := With("shared", "inner", "inner_only", 1).Errorf("inner")
				outer := With("shared", "outer", "outer_only", 2).Wrap(inner).(OopsError)
				s := snapshot(outer)
				is.Equal("inner", s.context["shared"])
				is.Equal(1, s.context["inner_only"])
				is.Equal(2, s.context["outer_only"])
			},
		},
		{
			name: "tags_deduplicated_across_chain",
			run: func(is *assert.Assertions) {
				inner := Tags("a", "b").Errorf("inner")
				outer := Tags("b", "c").Wrap(inner).(OopsError)
				s := snapshot(outer)
				is.ElementsMatch([]string{"a", "b", "c"}, s.tags)
			},
		},
		{
			name: "explicit_trace_wins_over_auto_generated",
			run: func(is *assert.Assertions) {
				// outer has explicit trace; inner gets auto-generated
				inner := Errorf("inner") // traceAutoGenerated=true
				outer := Trace("explicit-trace-id").Wrap(inner).(OopsError)
				s := snapshot(outer)
				is.Equal("explicit-trace-id", s.trace)
			},
		},
		{
			name: "auto_trace_used_when_no_explicit",
			run: func(is *assert.Assertions) {
				err := Errorf("err").(OopsError)
				s := snapshot(err)
				// AutoTraceID is on by default: a non-empty trace should be present
				is.NotEmpty(s.trace)
			},
		},
		{
			name: "context_lazy_value_evaluated",
			run: func(is *assert.Assertions) {
				err := With("computed", func() any { return "evaluated" }).Errorf("err").(OopsError)
				s := snapshot(err)
				is.Equal("evaluated", s.context["computed"])
			},
		},
		{
			name: "context_pointer_dereferenced",
			run: func(is *assert.Assertions) {
				val := "deref_me"
				err := With("ptr", &val).Errorf("err").(OopsError)
				s := snapshot(err)
				is.Equal("deref_me", s.context["ptr"])
			},
		},
		{
			name: "user_innermost_wins_data_merged",
			run: func(is *assert.Assertions) {
				inner := User("uid-inner", "role", "admin", "name", "inner").Errorf("inner")
				outer := User("uid-outer", "name", "outer").Wrap(inner).(OopsError)
				s := snapshot(outer)
				is.Equal("uid-inner", s.userID)
				is.Equal("admin", s.userData["role"])
				is.Equal("inner", s.userData["name"]) // inner wins
			},
		},
		{
			name: "tenant_innermost_wins_data_merged",
			run: func(is *assert.Assertions) {
				inner := Tenant("tid-inner", "plan", "enterprise").Errorf("inner")
				outer := Tenant("tid-outer", "plan", "free", "region", "us").Wrap(inner).(OopsError)
				s := snapshot(outer)
				is.Equal("tid-inner", s.tenantID)
				is.Equal("enterprise", s.tenantData["plan"]) // inner wins
				is.Equal("us", s.tenantData["region"])       // outer-only key preserved
			},
		},
		{
			name: "empty_error_zero_snapshot",
			run: func(is *assert.Assertions) {
				err := Errorf("bare").(OopsError)
				s := snapshot(err)
				is.Nil(s.code)
				is.Empty(s.domain)
				is.Empty(s.hint)
				is.Empty(s.public)
				is.Empty(s.owner)
				is.Empty(s.userID)
				is.Empty(s.tenantID)
				is.Empty(s.tags)
				is.Empty(s.context)
				is.Empty(s.userData)
				is.Empty(s.tenantData)
			},
		},
		{
			name: "snapshot_fields_not_cached_in_output",
			run: func(is *assert.Assertions) {
				// snapshot must not copy stacktrace, err, or cache fields
				err := Code("X").With("k", "v").Errorf("msg").(OopsError)
				s := snapshot(err)
				is.NoError(s.err)
				is.Nil(s.stacktrace)
				is.Nil(s.cacheOnce)
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
