
# Example of Zerolog logger

Playground: https://go.dev/play/p/aalqQ6wEDyx

```sh
go run examples/zerolog/example.go 2>&1 | jq
go run examples/zerolog/example.go 2>&1 | jq .stack -r
```
