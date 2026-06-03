module github.com/samber/oops/examples/zerolog

go 1.25.0

replace (
	github.com/samber/oops => ../..
	github.com/samber/oops/loggers/zerolog => ../../loggers/zerolog
)

require (
	github.com/rs/zerolog v1.31.0
	github.com/samber/oops v0.0.0
	github.com/samber/oops/loggers/zerolog v0.0.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/samber/lo v1.53.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
