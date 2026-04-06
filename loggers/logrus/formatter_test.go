package oopslogrus

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogrusFormatter(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := logrus.New()
	logger.SetOutput(buffer)
	logger.SetFormatter(NewOopsFormatter(&logrus.JSONFormatter{}))
	logger.SetLevel(logrus.ErrorLevel)

	err := oops.
		In("test").
		With("driver", "postgresql").
		Errorf("could not fetch user")

	logger.WithError(err).Error("something went wrong")

	var loggedEntry map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &loggedEntry))

	is.Equal("something went wrong", loggedEntry["msg"])
	is.Equal("test", loggedEntry["domain"])
	is.Equal("could not fetch user", loggedEntry["error"])
	is.NotEmpty(loggedEntry["trace"])
	is.NotEmpty(loggedEntry["stacktrace"])
	is.Equal(map[string]any{"driver": "postgresql"}, loggedEntry["context"])
}

func TestLogrusFormatter_NonOopsError(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := logrus.New()
	logger.SetOutput(buffer)
	logger.SetFormatter(NewOopsFormatter(&logrus.JSONFormatter{}))
	logger.SetLevel(logrus.ErrorLevel)

	logger.WithField("error", "plain string error").Error("something went wrong")

	var loggedEntry map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &loggedEntry))

	is.Equal("something went wrong", loggedEntry["msg"])
	is.Equal("plain string error", loggedEntry["error"])
	// Should NOT have oops-specific fields
	is.Nil(loggedEntry["domain"])
	is.Nil(loggedEntry["trace"])
}

func TestLogrusFormatter_StacktraceStrippedForPanicFatal(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Logrus levels: Panic=0, Fatal=1, Error=2, Warn=3, ...
	// The formatter strips stacktrace/sources when entry.Level < ErrorLevel (i.e. Panic and Fatal)
	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Data:    logrus.Fields{},
		Level:   logrus.FatalLevel,
		Message: "fatal crash",
	}

	err := oops.
		In("test").
		Errorf("a fatal error")

	entry.Data["error"] = err

	oopsErr := err.(oops.OopsError)
	oopsErrorToEntryData(&oopsErr, entry)

	is.Equal("test", entry.Data["domain"])
	// Stacktrace and sources should be stripped for Fatal level
	is.Nil(entry.Data["stacktrace"])
	is.Nil(entry.Data["sources"])
}

func TestLogrusFormatter_StacktraceKeptForErrorLevel(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := logrus.New()
	logger.SetOutput(buffer)
	logger.SetFormatter(NewOopsFormatter(&logrus.JSONFormatter{}))
	logger.SetLevel(logrus.ErrorLevel)

	err := oops.
		In("test").
		Errorf("an error")

	logger.WithError(err).Error("something happened")

	var loggedEntry map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &loggedEntry))

	is.Equal("test", loggedEntry["domain"])
	// Stacktrace should be kept for Error level
	is.NotNil(loggedEntry["stacktrace"])
}

func TestLogrusFormatter_NilFormatter(t *testing.T) {
	t.Parallel()

	// Should not panic with nil secondary formatter
	formatter := NewOopsFormatter(nil)
	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.formatter)
}
