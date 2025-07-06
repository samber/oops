package oops

import (
	"reflect"

	"github.com/samber/lo"
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
func lazyValueEvaluation(value any) any {
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

// mergeNestedErrorMap merges maps from an error chain, with later
// errors taking precedence over earlier ones in the chain.
//
// This function traverses the error chain recursively and merges
// maps (like context, user data, or tenant data) from all errors
// in the chain. Later errors in the chain can override values
// from earlier errors, allowing for progressive enhancement of
// error context as the error propagates through the system.
//
// The function uses the provided getter function to extract the
// map from each error in the chain. The merging is done using
// lo.Assign, which creates a new map with all values merged.
//
// Example usage:
//
//	context := mergeNestedErrorMap(err, func(e OopsError) map[string]any {
//	  return e.context
//	})
//	// Returns a merged map with context from all errors in the chain,
//	// with later errors overriding earlier ones
func mergeNestedErrorMap(err OopsError, getter func(OopsError) map[string]any) map[string]any {
	if err.err == nil {
		return getter(err)
	}

	if child, ok := AsOops(err.err); ok {
		return lo.Assign(map[string]any{}, getter(err), mergeNestedErrorMap(child, getter))
	}

	return getter(err)
}
