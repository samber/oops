
# Example of segfault handling

Playground: https://go.dev/play/p/66wkzJ-Rem1

```sh
go run examples/segfault/example.go 2>&1 | jq
go run examples/segfault/example.go 2>&1 | jq .stacktrace -r
```
