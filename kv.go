package oops

import (
	"fmt"
	"reflect"
	"time"
)

// dereferencePointers recursively dereferences pointer values in a map
// to provide more useful logging and debugging information.
//
// This function processes all values in the input map and recursively
// dereferences pointer types to extract their underlying values. This
// is particularly useful for logging where you want to see the actual
// data rather than pointer addresses.
//
// The function respects the global DereferencePointers configuration
// flag. If this flag is false, the function returns the input map
// unchanged.
//
// Example usage:
//
//	data := map[string]any{
//	  "user": &User{Name: "John"},
//	  "count": lo.ToPtr(42),
//	}
//	result := dereferencePointers(data)
//	// result["user"] will be User{Name: "John"} instead of *User
//	// result["count"] will be 42 instead of *int
func dereferencePointers(data map[string]any) map[string]any {
	if !DereferencePointers {
		return data
	}

	for key, value := range data {
		// Fast path: only use reflect for types that could be pointers
		switch value.(type) {
		case nil, string, int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64, bool, []byte,
			map[string]any, []any:
			continue // not a pointer, skip
		}
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr {
			data[key] = dereferencePointerRecursive(val, 0)
		}
	}

	return data
}

// dereferencePointerRecursive recursively dereferences pointer values
// to extract their underlying data, with protection against infinite
// recursion and nil pointers.
//
// This function handles complex pointer structures by recursively
// following pointer chains until it reaches a non-pointer value or
// hits safety limits. It includes several safety mechanisms:
// - Depth limiting to prevent infinite recursion (max 10 levels)
// - Nil pointer handling to prevent panics
// - Invalid value handling for edge cases
//
// Example usage:
//
//	var ptr1 *int = lo.ToPtr(42)
//	var ptr2 **int = &ptr1
//	var ptr3 ***int = &ptr2
//
//	val := reflect.ValueOf(ptr3)
//	result := dereferencePointerRecursive(val, 0)
//	// result will be 42 (int), not ***int
func dereferencePointerRecursive(val reflect.Value, depth int) (ret any) {
	defer func() {
		if r := recover(); r != nil {
			ret = nil
		}
	}()

	if !val.IsValid() {
		return nil
	}
	if val.Kind() != reflect.Ptr {
		return val.Interface()
	}

	if val.IsNil() {
		return nil
	}

	// Prevent infinite recursion with circular references
	if depth > 10 {
		return val.Interface()
	}

	elem := val.Elem()
	if !elem.IsValid() {
		return nil
	}

	// Recursively handle nested pointers
	if elem.Kind() == reflect.Ptr {
		return dereferencePointerRecursive(elem, depth+1)
	}

	return elem.Interface()
}

// lazyMapEvaluation processes a map and evaluates any lazy evaluation
// functions (functions with no arguments and one return value) to
// extract their computed values.
//
// This function enables lazy evaluation of expensive operations in
// error context. Instead of computing values immediately when the
// error is created, you can provide functions that will be evaluated
// only when the error data is actually accessed (e.g., during logging).
//
// The function recursively processes nested maps to handle complex
// data structures. It identifies functions by checking if they have
// no input parameters and exactly one output parameter.
//
// Example usage:
//
//	data := map[string]any{
//	  "timestamp": time.Now,
//	  "expensive": func() any { return computeExpensiveValue() },
//	  "simple": "static value",
//	}
//	result := lazyMapEvaluation(data)
//	// result["timestamp"] will be the actual time.Time value
//	// result["expensive"] will be the computed value
//	// result["simple"] will remain "static value"
func lazyMapEvaluation(data map[string]any) map[string]any {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]any:
			data[key] = lazyMapEvaluation(v)
		default:
			data[key] = lazyValueEvaluation(value)
		}
	}

	return data
}

// lazyValueEvaluation evaluates a single value, checking if it's a
// lazy evaluation function and executing it if so.
//
// This function identifies lazy evaluation functions by checking their
// signature: they must have no input parameters and exactly one output
// parameter. When such a function is found, it's executed and the
// result is returned instead of the function itself.
//
// This enables deferred computation of expensive values, which is
// particularly useful for error context where you want to capture
// the current state but defer expensive computations until they're
// actually needed.
//
// Example usage:
//
//	value := func() any { return expensiveComputation() }
//	result := lazyValueEvaluation(value)
//	// result will be the result of expensiveComputation()
//
//	value := "static string"
//	result := lazyValueEvaluation(value)
//	// result will be "static string" (unchanged)
func lazyValueEvaluation(value any) (ret any) {
	// Fast path: common types are never lazy evaluation functions
	switch value.(type) {
	case nil, string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool, []byte,
		time.Time, time.Duration:
		return value
	}

	defer func() {
		if r := recover(); r != nil {
			ret = fmt.Sprintf("panic in lazy evaluation: %v", r)
		}
	}()

	v := reflect.ValueOf(value)
	if !v.IsValid() || v.Kind() != reflect.Func {
		return value
	}

	// Check if this is a lazy evaluation function (no args, one return)
	if v.Type().NumIn() != 0 || v.Type().NumOut() != 1 {
		return value
	}

	return v.Call([]reflect.Value{})[0].Interface()
}

// recursive is a helper function that traverses the error chain
// and applies a function to each OopsError in the chain.
func recursive(err OopsError, tap func(OopsError) bool) {
	if !tap(err) {
		return
	}

	if err.err == nil {
		return
	}

	if child, ok := AsOops(err.err); ok {
		recursive(child, tap)
	}
}

// // recursive is a helper function that traverses the error chain
// // and applies a function to each OopsError in the chain.
// func recursiveBackward(err OopsError, tap func(OopsError)) {
// 	if err.err == nil {
// 		tap(err)
// 		return
// 	}

// 	if child, ok := AsOops(err.err); ok {
// 		recursiveBackward(child, tap)
// 	}

// 	tap(err)
// }

// getDeepestErrorAttribute extracts an attribute from the deepest error
// in an error chain, with fallback to the current error if the deepest
// error doesn't have the attribute set.
//
// This function traverses the error chain recursively to find the
// deepest oops.OopsError and extracts the specified attribute using
// the provided getter function. If the deepest error doesn't have
// the attribute set (returns a zero value), it falls back to the
// current error's attribute.
//
// This behavior is useful for attributes that should be set at the
// point where the error originates (like error codes, hints, or
// public messages) but can be overridden by wrapping errors.
//
// Example usage:
//
//	code := getDeepestErrorAttribute(err, func(e OopsError) string {
//	  return e.code
//	})
//	// Returns the error code from the deepest error in the chain
func getDeepestErrorAttribute[T comparable](err OopsError, getter func(OopsError) T) T {
	if err.err == nil {
		return getter(err)
	}

	if child, ok := AsOops(err.err); ok {
		return coalesceOrEmpty(getDeepestErrorAttribute(child, getter), getter(err))
	}

	return getter(err)
}

// mergeNestedErrorMap merges maps from an error chain, with deeper errors
// taking precedence over shallower ones (matching the original lo.Assign
// left-to-right semantics where the child/deeper map was the last argument).
//
// This function traverses the error chain and collects all maps via
// collectMaps, then merges them in a single pass — avoiding the O(n)
// intermediate map allocations that the recursive lo.Assign approach incurred.
//
// Example usage:
//
//	context := mergeNestedErrorMap(err, func(e OopsError) map[string]any {
//	  return e.context
//	})
//	// Returns a merged map with context from all errors in the chain,
//	// with deeper errors overriding shallower ones.
func mergeNestedErrorMap(err OopsError, getter func(OopsError) map[string]any) map[string]any {
	// Collect all maps from the chain (shallower first, deeper last).
	var maps []map[string]any
	collectMaps(err, getter, &maps)

	if len(maps) == 0 {
		return map[string]any{}
	}

	// Merge into a single result: deeper maps (appended last) overwrite
	// shallower ones, preserving the original semantics.
	// Preallocate with the first map's length as a rough capacity hint.
	result := make(map[string]any, len(maps[0]))
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// collectMaps appends maps from the error chain to result in shallow-to-deep
// order so that the final merge loop gives deeper errors higher precedence.
func collectMaps(err OopsError, getter func(OopsError) map[string]any, result *[]map[string]any) {
	if m := getter(err); len(m) > 0 {
		*result = append(*result, m)
	}
	if child, ok := AsOops(err.err); ok {
		collectMaps(child, getter, result)
	}
}
