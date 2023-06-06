package oopslogrus

import (
	"errors"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

func NewOopsFormatter(secondaryFormatter logrus.Formatter) *oopsFormatter {
	if secondaryFormatter == nil {
		secondaryFormatter = logrus.StandardLogger().Formatter
	}

	return &oopsFormatter{
		formatter: secondaryFormatter,
	}
}

type oopsFormatter struct {
	formatter logrus.Formatter
}

func (f *oopsFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	errField, ok := entry.Data["error"]
	if ok {
		switch err := errField.(type) {
		case error:
			var oopsError oops.OopsError
			if errors.As(err, &oopsError) {
				oopsErrorToEntryData(&oopsError, entry)
			}
		case any:
		}
	}

	return f.formatter.Format(entry)
}

func oopsErrorToEntryData(err *oops.OopsError, entry *logrus.Entry) {
	entry.Time = err.Time()

	payload := err.ToMap()

	if entry.Level < logrus.ErrorLevel {
		delete(payload, "stacktrace")
		delete(payload, "sources")
	}

	for k, v := range payload {
		entry.Data[k] = v
	}
}
