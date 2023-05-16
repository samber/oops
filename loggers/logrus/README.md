# Logrus formatter for Oops

```go
import "github.com/samber/oops/loggers/logrus"

func init() {
	logrus.SetFormatter(
        oopslogrus.NewOopsFormatter(
            &logrus.JSONFormatter{},
        )
    )
}

func main() {
    err := oops.
        With("driver", "postgresql").
        With("query", query).
        With("query.duration", queryDuration).
        Errorf("could not fetch user")

	logrus.WithError(err).Error(err)
}
```
