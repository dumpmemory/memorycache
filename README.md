<div align="center">
    <h1>MemoryCache</h1>
    <img src="assets/logo.png" alt="logo" width="300px">
    <h5>To the time to life, rather than to life in time.</h5>
</div>


[中文](README_CN.md)

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main
[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain
[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT
[4]: https://codecov.io/gh/lxzan/memorycache

### Description

Minimalist in-memory KV storage, powered by `HashMap` and `Minimal Quad Heap`, without optimizations for GC.

**Cache Elimination Policy:**

1. Set method cleans up overflowed keys
2. Active cycle cleans up expired keys

### Principle

-   Storage Data Limit: Limited by maximum capacity
-   Expiration Time: Supported
-   Cache Elimination Policy: LRU
-   GC Optimization: None
-   Persistent: None
-   Locking Mechanism: Slicing + Mutual Exclusion Locking

### Advantage

-   Simple and easy to use
-   No third-party dependencies
-   High performance
-   Low memory usage
-   Use quadruple heap to maintain the expiration time, effectively reduce the height of the tree, and improve the insertion performance

### Methods

-   [x] **Set** : Set key-value pair with expiring time. If the key already exists, the value will be updated. Also the expiration time will be updated.
-   [x] **SetWithCallback** : Set key-value pair with expiring time and callback function. If the key already exists, the value will be updated. Also the expiration time will be updated.
-   [x] **Get** : Get value by key. If the key does not exist, the second return value will be false.
-   [x] **GetWithTTL** : Get value by key. If the key does not exist, the second return value will be false. When return value, method will refresh the expiration time.
-   [x] **Delete** : Delete key-value pair by key.
-   [x] **GetOrCreate** : Get value by key. If the key does not exist, the value will be created.
-   [x] **GetOrCreateWithCallback** : Get value by key. If the key does not exist, the value will be created. Also the callback function will be called.

### Example

```go
package main

import (
	"fmt"
	"github.com/lxzan/memorycache"
	"time"
)

func main() {
	mc := memorycache.New[string, any](
		memorycache.WithBucketNum(128),                          // Bucket number, recommended to be a prime number.
		memorycache.WithBucketSize(1000, 10000),                 // Bucket size, initial size and maximum capacity.
		memorycache.WithInterval(5*time.Second, 30*time.Second), // Active cycle cleanup interval and expiration time.
	)

	mc.SetWithCallback("xxx", 1, time.Second, func(element *memorycache.Element[string, any], reason memorycache.Reason) {
		fmt.Printf("callback: key=%s, reason=%v\n", element.Key, reason)
	})

	val, exist := mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)

	time.Sleep(2 * time.Second)

	val, exist = mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)
}

```

### Benchmark

-   1,000,000 elements

```
go test -benchmem -run=^$ -bench . github.com/lxzan/memorycache/benchmark
goos: linux
goarch: amd64
pkg: github.com/lxzan/memorycache/benchmark
cpu: AMD Ryzen 5 PRO 4650G with Radeon Graphics
BenchmarkMemoryCache_Set-12             18891738               109.5 ns/op            11 B/op          0 allocs/op
BenchmarkMemoryCache_Get-12             21813127                48.21 ns/op            0 B/op          0 allocs/op
BenchmarkMemoryCache_SetAndGet-12       22530026                52.14 ns/op            0 B/op          0 allocs/op
BenchmarkRistretto_Set-12               13786928               140.6 ns/op           116 B/op          2 allocs/op
BenchmarkRistretto_Get-12               26299240                45.87 ns/op           16 B/op          1 allocs/op
BenchmarkRistretto_SetAndGet-12         11360748               103.0 ns/op            27 B/op          1 allocs/op
BenchmarkTheine_Set-12                   3527848               358.2 ns/op            19 B/op          0 allocs/op
BenchmarkTheine_Get-12                  23234760                49.37 ns/op            0 B/op          0 allocs/op
BenchmarkTheine_SetAndGet-12             6755134               176.3 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/lxzan/memorycache/benchmark  65.498s
```
