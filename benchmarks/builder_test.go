package benchmarks

// Benchmarks focused on the builder pattern before migration to functional options.
//
// The core overhead: every builder method calls copy(), which clones 3 maps
// (context, userData, tenantData) and 1 slice (tags). A chain of N methods
// pays that cost N times, regardless of whether the error is ever read.
//
// Run:
//   go test -bench=. -benchmem -count=10 ./benchmarks/ | tee bench-builder.txt
//
// Compare after refactoring:
//   benchstat bench-builder.txt bench-functional-options.txt

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/samber/oops"
)

// BenchmarkBuilderChainDepth shows how cost scales with chain length.
// Each depth-N run adds one more builder method (one more copy() call) vs depth-(N-1).
// This is the primary cost that functional options will eliminate.
func BenchmarkBuilderChainDepth(b *testing.B) {
	inner := errors.New("inner error")

	cases := []struct {
		name string
		fn   func() error
	}{
		// depth=0: baseline — newBuilder() + Wrap only, no intermediate copies
		{"depth=0", func() error {
			return oops.Wrap(inner)
		}},
		// depth=1: one copy()
		{"depth=1", func() error {
			return oops.In("database").Wrap(inner)
		}},
		// depth=2: two copies
		{"depth=2", func() error {
			return oops.In("database").Code("db_error").Wrap(inner)
		}},
		// depth=4: four copies
		{"depth=4", func() error {
			return oops.
				In("database").
				Code("db_error").
				Hint("check connection pool size").
				Owner("backend-team").
				Wrap(inner)
		}},
		// depth=7: seven copies — representative of a rich error in production
		{"depth=7", func() error {
			return oops.
				In("database").
				Code("db_error").
				Tags("critical", "database").
				With("query", "SELECT * FROM users", "duration_ms", 1500).
				Hint("check connection pool size").
				Owner("backend-team").
				Trace("req-abc-123").
				Wrap(inner)
		}},
		// depth=10: stress — uncommon but shows the slope clearly
		{"depth=10", func() error {
			return oops.
				In("database").
				Code("db_error").
				Tags("critical", "database").
				With("query", "SELECT * FROM users", "duration_ms", 1500).
				Hint("check connection pool size").
				Owner("backend-team").
				Trace("req-abc-123").
				Span("query-span").
				User("user-123", "role", "admin").
				Tenant("tenant-456", "plan", "enterprise").
				Wrap(inner)
		}},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = tc.fn()
			}
		})
	}
}

// BenchmarkBuilderMethodCost isolates the overhead of each individual builder method.
// Every benchmark applies exactly one method before Wrap, so the delta vs
// BenchmarkBuilderChainDepth/depth=0 is the true per-method cost (one copy()).
//
// Some methods do extra work beyond copy():
//   - Tags: appends to the cloned slice
//   - With: inserts into the cloned context map
//   - User/Tenant: runs identityArgsToMap + inserts into userData/tenantData map
func BenchmarkBuilderMethodCost(b *testing.B) {
	inner := errors.New("inner error")

	cases := []struct {
		name string
		fn   func() error
	}{
		// Scalar field setters — copy() + one field assignment
		{"Code", func() error { return oops.Code("db_error").Wrap(inner) }},
		{"In", func() error { return oops.In("database").Wrap(inner) }},
		{"Hint", func() error { return oops.Hint("check pool").Wrap(inner) }},
		{"Owner", func() error { return oops.Owner("backend-team").Wrap(inner) }},
		{"Public", func() error { return oops.Public("service unavailable").Wrap(inner) }},
		{"Trace", func() error { return oops.Trace("req-abc-123").Wrap(inner) }},
		{"Span", func() error { return oops.Span("op-span").Wrap(inner) }},

		// Slice mutation — copy() + slice append
		{"Tags/1", func() error { return oops.Tags("critical").Wrap(inner) }},
		{"Tags/3", func() error { return oops.Tags("critical", "database", "query").Wrap(inner) }},

		// Map mutation — copy() + map insert(s)
		{"With/1kv", func() error { return oops.With("query", "SELECT 1").Wrap(inner) }},
		{"With/4kv", func() error {
			return oops.With("query", "SELECT 1", "duration_ms", 42, "rows", 0, "db", "primary").Wrap(inner)
		}},

		// identityArgsToMap + map mutation
		{"User/0attr", func() error { return oops.User("user-123").Wrap(inner) }},
		{"User/2attr", func() error { return oops.User("user-123", "role", "admin", "plan", "pro").Wrap(inner) }},
		{"Tenant/2attr", func() error { return oops.Tenant("tenant-456", "name", "Acme", "plan", "enterprise").Wrap(inner) }},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = tc.fn()
			}
		})
	}
}

// BenchmarkBuilderCopyScaling shows how the cost of each copy() grows as context
// maps are already populated. A chain of N methods on a builder that already has
// M context entries pays O(M) per copy — this compounds with chain depth.
func BenchmarkBuilderCopyScaling(b *testing.B) {
	inner := errors.New("inner error")

	cases := []struct {
		name    string
		entries int
	}{
		{"context=0kv", 0},
		{"context=5kv", 5},
		{"context=10kv", 10},
		{"context=20kv", 20},
	}

	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			// Build the base builder once with tc.entries in its context map.
			// Each iteration then adds one more method (one copy() of the populated map).
			kv := make([]any, 0, tc.entries*2)
			for i := 0; i < tc.entries; i++ {
				kv = append(kv, "key"+string(rune('a'+i)), i)
			}
			// Pre-build a reusable base builder (not measured).
			base := oops.With(kv...)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// One extra builder step triggers one copy() of the populated context map.
				_ = base.In("database").Wrap(inner)
			}
		})
	}
}

// BenchmarkBuilderNilError measures the no-op path when the wrapped error is nil.
// This is common in code like: return oops.In("db").Wrap(possiblyNilErr)
// The builder chain still pays copy() costs even when the result is discarded.
func BenchmarkBuilderNilError(b *testing.B) {
	cases := []struct {
		name string
		fn   func() error
	}{
		{"chain=0", func() error { return oops.Wrap(nil) }},
		{"chain=3", func() error { return oops.In("db").Code("err").Hint("hint").Wrap(nil) }},
		{"chain=7", func() error {
			return oops.
				In("database").
				Code("db_error").
				Tags("critical").
				With("query", "SELECT 1").
				Hint("check pool").
				Owner("team").
				Trace("req-id").
				Wrap(nil)
		}},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = tc.fn()
			}
		})
	}
}

// BenchmarkBuilderCreatedNotRead models the dominant production pattern:
// the error is fully built and returned but never inspected by the caller
// (e.g., propagated up via `if err != nil { return oops.In("svc").Wrap(err) }`).
//
// This is the highest-frequency path in production. Every allocation here is waste
// for errors that are eventually logged (rare) but never introspected (common).
func BenchmarkBuilderCreatedNotRead(b *testing.B) {
	inner := errors.New("db: connection refused")

	cases := []struct {
		name string
		fn   func() error
	}{
		// Minimal — single Wrap, no chain
		{"simple/Wrap", func() error {
			return oops.Wrap(inner)
		}},
		// Typical service boundary — 3-method chain
		{"typical/chain=3", func() error {
			return oops.In("database").Code("db_error").Wrap(inner)
		}},
		// Rich error — full context, 7-method chain
		{"rich/chain=7", func() error {
			return oops.
				In("database").
				Code("db_error").
				Tags("critical", "database").
				With("query", "SELECT * FROM users", "duration_ms", 1500).
				Hint("check connection pool size").
				Owner("backend-team").
				Trace("req-abc-123").
				Wrap(inner)
		}},
	}

	for _, tc := range cases {
		// Simulate the "not read" path: create the error and immediately discard it.
		// The caller only checked err != nil; no fields were accessed.
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				err := tc.fn()
				// Prevent dead-code elimination without accessing any fields.
				if err == nil {
					b.Fatal("unexpected nil")
				}
			}
		})
	}
}

// BenchmarkBuilderVsStdlib compares oops builder cost against the stdlib baseline.
// Use this to quantify the overhead of oops vs plain errors.New / fmt.Errorf.
func BenchmarkBuilderVsStdlib(b *testing.B) {
	inner := errors.New("inner")

	b.Run("stdlib/errors.New", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.New("an error occurred")
		}
	})

	b.Run("stdlib/fmt.Errorf", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.New("wrapped: " + inner.Error())
		}
	})

	b.Run("oops/Wrap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = oops.Wrap(inner)
		}
	})

	b.Run("oops/chain=3/Wrap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = oops.In("database").Code("db_error").Wrap(inner)
		}
	})

	b.Run("oops/chain=7/Wrap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = oops.
				In("database").
				Code("db_error").
				Tags("critical", "database").
				With("query", "SELECT 1", "duration_ms", 42).
				Hint("check pool").
				Owner("backend-team").
				Trace("req-abc-123").
				Wrap(inner)
		}
	})
}

// BenchmarkOopsErrorMemoryFootprint measures the live heap cost of holding N errors
// simultaneously. Two sub-benchmarks per count:
//
//   - alloc: standard benchmark — B/op = total bytes to allocate N errors per iteration;
//     divide by "errors/op" custom metric to get per-error cost.
//   - live: single snapshot using runtime.MemStats after GC, reports "live-B/error"
//     and "live-B/total" custom metrics showing actual retained heap.
//
// Variants: simple (Wrap only) and rich (full 7-method chain with context).
func BenchmarkOopsErrorMemoryFootprint(b *testing.B) {
	inner := errors.New("root cause")

	type variant struct {
		name string
		make func(j int) error
	}

	variants := []variant{
		{
			"simple",
			func(j int) error { return oops.Wrap(inner) },
		},
		{
			"rich",
			func(j int) error {
				return oops.
					In("database").
					Code("db_error").
					Tags("critical", "database").
					With("query", "SELECT * FROM users", "user_id", j).
					Hint("check connection pool").
					Owner("backend-team").
					Trace(fmt.Sprintf("req-%d", j)).
					Wrap(inner)
			},
		},
	}

	counts := []int{1, 10, 100, 1_000, 10_000}

	for _, v := range variants {
		for _, count := range counts {
			v, count := v, count

			// alloc: measures allocation throughput — B/op covers N errors per iteration.
			b.Run(fmt.Sprintf("%s/alloc/count=%d", v.name, count), func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(count), "errors/op")
				var sink []error
				for i := 0; i < b.N; i++ {
					errs := make([]error, count)
					for j := 0; j < count; j++ {
						errs[j] = v.make(j)
					}
					sink = errs
				}
				runtime.KeepAlive(sink)
			})

			// live: single heap snapshot — reports retained bytes after GC.
			b.Run(fmt.Sprintf("%s/live/count=%d", v.name, count), func(b *testing.B) {
				// Establish a clean heap baseline.
				runtime.GC()
				runtime.GC()
				var before runtime.MemStats
				runtime.ReadMemStats(&before)

				// Allocate and hold N errors — these must stay live through the GC below.
				errs := make([]error, count)
				for j := 0; j < count; j++ {
					errs[j] = v.make(j)
				}

				// Collect garbage; only the live slice survives.
				runtime.GC()
				runtime.GC()
				var after runtime.MemStats
				runtime.ReadMemStats(&after)

				liveTotal := int64(after.HeapInuse) - int64(before.HeapInuse)
				if liveTotal < 0 {
					liveTotal = 0
				}
				b.ReportMetric(float64(liveTotal), "live-B/total")
				b.ReportMetric(float64(liveTotal)/float64(count), "live-B/error")

				// Prevent the compiler from collecting errs before the second GC.
				runtime.KeepAlive(errs)

				// Drive b.N to 1 — this benchmark is a snapshot, not a loop.
				for i := 0; i < b.N; i++ {
				}
			})
		}
	}
}
