# Command

All benchmarks were running on the same machine.

```bash
go test -benchmem -cpuprofile cpu.prof -memprofile mem.prof -bench=.
```

# v0.1.0

```
goos: linux
goarch: amd64
pkg: github.com/nanmu42/gzip
BenchmarkSoleGin_SmallPayload-12                         7222698               244 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_SmallPayload-12           1000000              1076 ns/op             224 B/op          6 allocs/op
BenchmarkSoleGin_BigPayload-12                           6688381               265 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_BigPayload-12             1000000              1063 ns/op             224 B/op          6 allocs/op
PASS
ok      github.com/nanmu42/gzip 6.222s
```

# v0.2.0

```
goos: linux
goarch: amd64
pkg: github.com/nanmu42/gzip
BenchmarkSoleGin_SmallPayload-12                         6769252               182 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_SmallPayload-12           1410784               740 ns/op             224 B/op          6 allocs/op
BenchmarkSoleGin_BigPayload-12                           7300908               218 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_BigPayload-12             2312258               726 ns/op             224 B/op          6 allocs/op
PASS
ok      github.com/nanmu42/gzip 7.428s
```