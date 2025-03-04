# Zerolog formatter for Oops

```go
import oopszerolog "github.com/samber/oops/loggers/zerolog"

func init() {
	zerolog.ErrorStackMarshaler = oopszerolog.OopsStackMarshaller
	zerolog.ErrorMarshalFunc = oopszerolog.OopsMarshalFunc
}

func main() {
	err := oops.
		With("driver", "postgresql").
		With("query", query).
		With("query.duration", queryDuration).
		Errorf("could not fetch user")

	if err != nil {
		zerolog.New(os.Stderr).Error().Stack().Err(err).Msg(err.Error())
	}
}
```
