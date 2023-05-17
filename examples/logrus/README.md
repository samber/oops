
# Example of Logrus logger

Playground: https://go.dev/play/p/IVeDMt4ouP-

```sh
go run examples/logrus/example.go 2>&1 | jq
go run examples/logrus/example.go 2>&1 | jq .stacktrace -r
```