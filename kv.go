package oops

import (
	"reflect"

	"github.com/samber/lo"
)

func dereferencePointers(data map[string]any) map[string]any {
	if !DereferencePointers {
		return data
	}

	for key, value := range data {
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr {
			data[key] = dereferencePointer(val)
		}
	}

	return data
}

func dereferencePointer(val reflect.Value) any {
	if !val.IsValid() {
		return nil
	}

	if val.IsNil() {
		return nil
	}

	return dereferencePointerRecursive(val, 0)
}

func dereferencePointerRecursive(val reflect.Value, depth int) any {
	if !val.IsValid() {
		return nil
	}

	if val.IsNil() {
		return nil
	}

	if depth > 10 {
		return val.Interface()
	}

	elem := val.Elem()
	if !elem.IsValid() {
		return nil
	}

	if elem.Kind() == reflect.Ptr {
		return dereferencePointerRecursive(elem, depth+1)
	}

	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	return elem.Interface()
}

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

func lazyValueEvaluation(value any) any {
	v := reflect.ValueOf(value)
	if !v.IsValid() || v.Kind() != reflect.Func {
		return value
	}

	if v.Type().NumIn() != 0 || v.Type().NumOut() != 1 {
		return value
	}

	return v.Call([]reflect.Value{})[0].Interface()
}

func getDeepestErrorAttribute[T comparable](err OopsError, getter func(OopsError) T) T {
	if err.err == nil {
		return getter(err)
	}

	if child, ok := AsOops(err.err); ok {
		return coalesceOrEmpty(getDeepestErrorAttribute(child, getter), getter(err))
	}

	return getter(err)
}

func mergeNestedErrorMap(err OopsError, getter func(OopsError) map[string]any) map[string]any {
	if err.err == nil {
		return getter(err)
	}

	if child, ok := AsOops(err.err); ok {
		return lo.Assign(map[string]any{}, getter(err), mergeNestedErrorMap(child, getter))
	}

	return getter(err)
}
