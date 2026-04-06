package benchmarks

import (
	"errors"
	"testing"

	"github.com/samber/lo"
	"github.com/samber/oops"
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.New("an error")
	}
}

func BenchmarkErrorfSimple(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.Errorf("an error: %s", "details")
	}
}

func BenchmarkErrorfWrap(b *testing.B) {
	inner := errors.New("inner")
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.Wrapf(inner, "wrapped")
	}
}

func BenchmarkWrap(b *testing.B) {
	inner := errors.New("inner")
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.Wrap(inner)
	}
}

func BenchmarkWrapf(b *testing.B) {
	inner := errors.New("inner")
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.Wrapf(inner, "context: %s", "details")
	}
}

func BenchmarkBuilderWithContext(b *testing.B) {
	inner := errors.New("inner")
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.
			In("database").
			Code("db_error").
			Tags("critical", "database").
			With("query", "SELECT 1", "duration_ms", 42).
			Hint("check connection pool").
			Owner("backend-team").
			Wrap(inner)
	}
}

func BenchmarkChainTraversal(b *testing.B) {
	err := oops.In("layer1").Wrap(
		oops.In("layer2").Wrap(
			oops.In("layer3").Errorf("root cause"),
		),
	)
	oopsErr, _ := oops.AsOops(err)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = oopsErr.Domain()
		_ = oopsErr.Context()
		_ = oopsErr.Trace()
		_ = oopsErr.Tags()
	}
}

func BenchmarkToMap(b *testing.B) {
	err := oops.
		In("database").
		Code("db_error").
		Tags("critical").
		With("query", "SELECT 1").
		Owner("backend-team").
		User("user-123", "name", "john").
		Errorf("connection failed")
	oopsErr, _ := oops.AsOops(err)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = oopsErr.ToMap()
	}
}

func BenchmarkWrapNil(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.Wrap(nil)
	}
}

func BenchmarkNewStacktrace(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.New("an error") // New() captures stacktrace, good proxy
	}
}

func BenchmarkLogValue(b *testing.B) {
	err := oops.
		In("database").
		Code("db_error").
		Tags("critical").
		With("query", "SELECT 1").
		Owner("backend-team").
		User("user-123", "name", "john").
		Errorf("connection failed")
	oopsErr, _ := oops.AsOops(err)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = oopsErr.LogValue()
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	err := oops.
		In("database").
		Code("db_error").
		Tags("critical").
		With("query", "SELECT 1").
		Owner("backend-team").
		Errorf("connection failed")
	oopsErr, _ := oops.AsOops(err)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = oopsErr.MarshalJSON()
	}
}

func BenchmarkErrorDeepChain(b *testing.B) {
	err := oops.In("layer1").Wrap(
		oops.In("layer2").Wrap(
			oops.In("layer3").Wrap(
				oops.In("layer4").Wrap(
					oops.In("layer5").Errorf("root cause"),
				),
			),
		),
	)
	oopsErr, _ := oops.AsOops(err)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = oopsErr.Error()
	}
}

func BenchmarkReflectionPaths(b *testing.B) {
	type MyStruct struct{ Name string }
	ptr := &MyStruct{Name: "test"}
	err := oops.With("user", ptr, "count", lo.ToPtr(42), "name", "john").Errorf("test error")
	oopsErr, _ := oops.AsOops(err)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = oopsErr.Context()
	}
}

func BenchmarkJoin(b *testing.B) {
	err1 := oops.In("service1").Errorf("error 1")
	err2 := oops.In("service2").Errorf("error 2")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.Join(err1, err2)
	}
}

func BenchmarkRecover(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = oops.Recover(func() { panic("test panic") })
	}
}
