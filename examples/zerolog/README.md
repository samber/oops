
# Example of Zerolog logger

Playground: https://go.dev/play/p/DaHzR4Zc-jj

```sh
go run examples/zerolog/example.go 2>&1 | jq
go run examples/zerolog/example.go 2>&1 | jq .stacktrace -r
```
