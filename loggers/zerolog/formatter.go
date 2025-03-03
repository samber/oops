package oopszerolog

import (
	"github.com/rs/zerolog"
	"github.com/samber/oops"
)

func OopsStackMarshaller(err error) interface{} {
	if typedErr, ok := oops.AsOops(err); ok {
		return typedErr.Stacktrace()
	}
	return err
}

func OopsMarshalFunc(err error) interface{} {
	if typedErr, ok := oops.AsOops(err); ok {
		return zerologErrorMarshaller{err: typedErr}
	}
	return err
}

type zerologErrorMarshaller struct {
	err oops.OopsError
}

func (m zerologErrorMarshaller) MarshalZerologObject(e *zerolog.Event) {
	payload := m.err.ToMap()
	for k, v := range payload {
		switch k {
		case "stacktrace":
		case "context":
			if context, ok := v.(map[string]any); ok && len(context) > 0 {
				dict := zerolog.Dict()
				for k, v := range context {
					switch vTyped := v.(type) {
					case nil:
					case error:
						dict = dict.Str(k, vTyped.Error())
					default:
						dict = dict.Interface(k, vTyped)
					}
				}
				e.Dict(k, dict)
			}
		default:
			e.Any(k, v)
		}
	}
}
