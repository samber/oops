
# Example of print sources

Playground: https://go.dev/play/p/6dEVisE1jfq

```sh
go run examples/sources/example.go | jq
go run examples/sources/example.go | jq .error.sources -r
```
