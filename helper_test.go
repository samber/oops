package oops

func withoutStacktrace(o OopsError) OopsError {
	o.stacktrace = nil
	return o
}
