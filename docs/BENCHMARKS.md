# Benchmark methodology

Run all correctness checks first:

```sh
go test ./...
go test -race ./...
go vet ./...
```

Then collect allocation-aware microbenchmarks:

```sh
go test ./market -run '^$' \
  -bench 'BenchmarkCandle(Ingestion|RangeScan)$' \
  -benchmem -count 5
```

The ingestion benchmark writes batches of 100 candles. The range benchmark
loads 2,000 candles in bounded setup batches and repeatedly retrieves the most
recent 252 through the composite primary index.

Record CPU, Go version, commit, sample count, and raw output whenever publishing
numbers. Do not compare results across machines as if they were the same test.
The current copy-on-write engine favors educational clarity and durability over
bulk-ingestion throughput; allocation profiles should guide Phase 6 rather than
being hidden.

## Validation snapshot

Captured on 2026-08-29 with Go 1.23.12 on an AMD EPYC 9V74 runner. These are
development-machine results, not cross-machine performance claims.

| Operation | Result | Allocations |
| --- | ---: | ---: |
| Atomic ingestion of 100 candles | 30.98 ms/op (~3.2k rows/s) | 45.9 MB, 15,458 allocs/op |
| Indexed retrieval of latest 252 candles | 2.39 ms/op | 1.86 MB, 5,260 allocs/op |

Command used `-benchtime=5x -benchmem`. The range latency is already suitable
for an interactive student demo; ingestion allocations are the clearest target
for a future profiling and optimization write-up.
