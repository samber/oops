package oops

import (
	"reflect"
)

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
