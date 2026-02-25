package oopszap

import (
	"github.com/samber/oops"
	"go.uber.org/zap/zapcore"
)

// OopsStackMarshaller returns the stack trace string for use in zap.
// Usage: zap.String("stacktrace", oopszap.OopsStackMarshaller(err))
func OopsStackMarshaller(err error) string {
	if typedErr, ok := oops.AsOops(err); ok {
		return typedErr.Stacktrace()
	}
	// For normal errors, we might not want to return anything or just empty string,
	// but to be safe/useful let's return nothing or handle it at call site.
	// Actually, matching zerolog implementation logic:
	return ""
}

// OopsMarshalFunc returns a zapcore.ObjectMarshaler that logs the error details.
// Usage: zap.Object("error", oopszap.OopsMarshalFunc(err))
func OopsMarshalFunc(err error) zapcore.ObjectMarshaler {
	if typedErr, ok := oops.AsOops(err); ok {
		return &zapErrorMarshaller{err: typedErr}
	}
	return &simpleErrorMarshaller{err: err}
}

type simpleErrorMarshaller struct {
	err error
}

func (m *simpleErrorMarshaller) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m.err != nil {
		enc.AddString("message", m.err.Error())
	}
	return nil
}

type zapErrorMarshaller struct {
	err oops.OopsError
}

func (m *zapErrorMarshaller) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	payload := m.err.ToMap()

	for k, v := range payload {
		switch k {
		case "stacktrace":
			// Skip stacktrace in the main object - handled separately if desired
		case "context":
			if contextMap, ok := v.(map[string]any); ok && len(contextMap) > 0 {
				enc.AddObject(k, zapcore.ObjectMarshalerFunc(func(innerEnc zapcore.ObjectEncoder) error {
					for ctxK, ctxV := range contextMap {
						if errVal, ok := ctxV.(error); ok {
							innerEnc.AddString(ctxK, errVal.Error())
						} else {
							innerEnc.AddReflected(ctxK, ctxV)
						}
					}
					return nil
				}))
			}
		default:
			enc.AddReflected(k, v)
		}
	}
	return nil
}
