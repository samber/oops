
# Example of slog logger

Playground: https://go.dev/play/p/-X2ZnqjyDLu

```sh
go run examples/slog/example.go | jq
go run examples/slog/example.go | jq .error.stacktrace -r
```
