# Command

```bash
go test -benchmem -bench .
```

# v0.1.0

```
$ go test -benchmem -bench .
goos: linux
goarch: amd64
pkg: github.com/nanmu42/gzip
BenchmarkSoleGin_SmallPayload-12                         7977210               156 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_SmallPayload-12           1242836               980 ns/op             224 B/op          6 allocs/op
BenchmarkSoleGin_BigPayload-12                           7137006               189 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_BigPayload-12             1908733               576 ns/op             224 B/op          6 allocs/op
PASS
ok      github.com/nanmu42/gzip 6.861s

```