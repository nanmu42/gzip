# Command

All benchmarks were running on the same machine.

```bash
go test -benchmem -cpuprofile cpu.prof -memprofile mem.prof -bench=.
```

# v0.1.0

v0.1.0 gets things working.

go version go1.13.4 linux/amd64

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

v0.2.0 uses ahocorasick in substring matching.

go version go1.13.4 linux/amd64

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

# v0.3.0

v0.3.0 reuses writerWrapper to gain less allocations and GC pressure.

go version go1.13.4 linux/amd64

```
goos: linux
goarch: amd64
pkg: github.com/nanmu42/gzip
BenchmarkSoleGin_SmallPayload-12                         7376715               194 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_SmallPayload-12           2475199               466 ns/op              96 B/op          3 allocs/op
BenchmarkSoleGin_BigPayload-12                           6572848               191 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_BigPayload-12             2991879               398 ns/op              96 B/op          3 allocs/op
PASS
ok      github.com/nanmu42/gzip 6.425s
```

# v0.4.0

v0.4.0 fixes panic on Gin's no route error.

# v0.5.0

v0.5.0 fixes panic on second calling to Write().

```
goos: linux
goarch: amd64
pkg: github.com/nanmu42/gzip
BenchmarkSoleGin_SmallPayload-12                         7490284               201 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_SmallPayload-12           2292319               501 ns/op              96 B/op          3 allocs/op
BenchmarkSoleGin_BigPayload-12                           6403441               190 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_BigPayload-12             2951451               410 ns/op              96 B/op          3 allocs/op
PASS
ok      github.com/nanmu42/gzip 6.620s
```

# v0.6.0

v0.6.0 fixes wrong status code handling CORS OPTIONS request by gin's other middleware.

# v0.7.0

* writerWrapper: buffer writes to decide whether use gzip or not
* writerWrapper: detect Content-Type if there's none
* ginGzipWriter: full implementation for gin.ResponseWriter excluding Pusher()


```
goos: linux
goarch: amd64
pkg: github.com/nanmu42/gzip
BenchmarkSoleGin_SmallPayload-12                         7900057               184 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_SmallPayload-12           2171088               510 ns/op              96 B/op          3 allocs/op
BenchmarkSoleGin_BigPayload-12                           7402651               184 ns/op              64 B/op          2 allocs/op
BenchmarkGinWithDefaultHandler_BigPayload-12             2911062               404 ns/op              96 B/op          3 allocs/op
PASS
ok      github.com/nanmu42/gzip 6.634s
```