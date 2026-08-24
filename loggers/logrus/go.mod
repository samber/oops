module github.com/samber/oops/loggers/logrus

go 1.25.0

replace github.com/samber/oops => ../..

require (
	github.com/samber/oops v0.0.0
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.11.1
	go.uber.org/goleak v1.3.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/oklog/ulid/v2 v2.1.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/samber/lo v1.53.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
