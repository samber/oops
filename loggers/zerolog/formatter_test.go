package oopszerolog

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/samber/oops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jsonLogEntryError struct {
	Error   string
	Time    string
	Domain  string
	Trace   string
	Context map[string]any
}

type jsonLogEntry struct {
	Level   string
	Stack   string
	Message string
	Error   jsonLogEntryError
}

func TestZerologFormatter(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	zerolog.ErrorStackMarshaler = OopsStackMarshaller
	zerolog.ErrorMarshalFunc = OopsMarshalFunc

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := zerolog.New(buffer).Level(zerolog.ErrorLevel)
	err := oops.
		In("test").
		With("driver", "postgresql").
		Errorf("could not fetch user")
	logger.Error().Stack().Err(err).Msg("something went wrong")

	loggedError := jsonLogEntry{}
	err = json.Unmarshal(buffer.Bytes(), &loggedError)
	require.NoError(t, err)

	is.Contains(loggedError.Stack, "Oops: could not fetch user\n  --- at ")
	is.NotEmpty(loggedError.Error.Time)
	is.NotEmpty(loggedError.Error.Trace)
	loggedError.Stack = ""
	loggedError.Error.Time = ""
	loggedError.Error.Trace = ""
	is.Equal(jsonLogEntry{
		Level:   "error",
		Message: "something went wrong",
		Error: jsonLogEntryError{
			Error:  "could not fetch user",
			Domain: "test",
			Context: map[string]any{
				"driver": "postgresql",
			},
		},
	}, loggedError)
}
