package benchmarks

import (
	"errors"
	"testing"

	"github.com/samber/oops"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = oops.New("an error")
	}
}

func BenchmarkErrorfSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = oops.Errorf("an error: %s", "details")
	}
}

func BenchmarkErrorfWrap(b *testing.B) {
	inner := errors.New("inner")
	for i := 0; i < b.N; i++ {
		_ = oops.Errorf("wrapped: %w", inner)
	}
}

func BenchmarkWrap(b *testing.B) {
	inner := errors.New("inner")
	for i := 0; i < b.N; i++ {
		_ = oops.Wrap(inner)
	}
}

func BenchmarkWrapf(b *testing.B) {
	inner := errors.New("inner")
	for i := 0; i < b.N; i++ {
		_ = oops.Wrapf(inner, "context: %s", "details")
	}
}

func BenchmarkBuilderWithContext(b *testing.B) {
	inner := errors.New("inner")
	for i := 0; i < b.N; i++ {
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
	for i := 0; i < b.N; i++ {
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
	for i := 0; i < b.N; i++ {
		_ = oopsErr.ToMap()
	}
}

func BenchmarkWrapNil(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = oops.Wrap(nil)
	}
}
