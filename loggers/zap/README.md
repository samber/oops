# Zap formatter for Oops

```go
import "go.uber.org/zap"
import oopszap "github.com/samber/oops/loggers/zap"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	err := oops.
		With("driver", "postgresql").
		With("query", "SELECT * FROM users").
		Errorf("could not fetch user")

	if err != nil {
		logger.Error(err.Error(),
			zap.Object("error", oopszap.OopsMarshalFunc(err)),
			zap.String("stacktrace", oopszap.OopsStackMarshaller(err)),
		)
	}
}
```
