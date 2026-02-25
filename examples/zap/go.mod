module github.com/samber/oops/examples/zap

go 1.21

replace (
	github.com/samber/oops => ../..
	github.com/samber/oops/loggers/zap => ../../loggers/zap
)

require (
	github.com/samber/oops v0.0.0
	github.com/samber/oops/loggers/zap v0.0.0
	go.uber.org/zap v1.26.0
)

require (
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/samber/lo v1.52.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
