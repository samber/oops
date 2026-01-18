package oopszap

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/samber/oops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type jsonLogEntryError struct {
	Error   string         `json:"error"`
	Time    string         `json:"time"`
	Domain  string         `json:"domain"`
	Trace   string         `json:"trace"`
	Context map[string]any `json:"context"`
}

type jsonLogEntry struct {
	Level      string            `json:"level"`
	Stacktrace string            `json:"stacktrace"`
	Message    string            `json:"msg"`
	Error      jsonLogEntryError `json:"error"`
}

func TestZapFormatter(t *testing.T) {
	is := assert.New(t)

	// Setup Zap logger buffer
	buffer := &bytes.Buffer{}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "" // disable timestamp for easier testing or keep it? zerolog test kept it but cleared it.
	// We want to test the payload structure.

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(buffer),
		zap.ErrorLevel,
	)
	logger := zap.New(core)

	err := oops.
		In("test").
		With("driver", "postgresql").
		Errorf("could not fetch user")

	logger.Error("something went wrong",
		zap.Object("error", OopsMarshalFunc(err)),
		zap.String("stacktrace", OopsStackMarshaller(err)),
	)

	var loggedError jsonLogEntry
	decErr := json.Unmarshal(buffer.Bytes(), &loggedError)
	require.NoError(t, decErr)

	// Assertions
	is.Contains(loggedError.Stacktrace, "Oops: could not fetch user\n  --- at ")
	is.NotEmpty(loggedError.Error.Time)
	is.NotEmpty(loggedError.Error.Trace)

	// Clear dynamic fields for equality check
	loggedError.Stacktrace = ""
	loggedError.Error.Time = ""
	loggedError.Error.Trace = ""

	expected := jsonLogEntry{
		Level:   "error",
		Message: "something went wrong",
		Error: jsonLogEntryError{
			Error:  "could not fetch user",
			Domain: "test",
			Context: map[string]any{
				"driver": "postgresql",
			},
		},
	}
	is.Equal(expected, loggedError)
}

func BenchmarkZapFormatter(b *testing.B) {
	err := oops.
		In("test").
		With("driver", "postgresql").
		Errorf("could not fetch user")

	// Discard output
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(bytes.NewBuffer(nil)),
		zap.ErrorLevel,
	)
	logger := zap.New(core)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Error("something went wrong",
			zap.Object("error", OopsMarshalFunc(err)),
		)
	}
}
