
# Example of panic handling

Playground: https://go.dev/play/p/uGwrFj9mII8

```sh
go run examples/panic/example.go 2>&1 | jq
go run examples/panic/example.go 2>&1 | jq .stacktrace -r
```
