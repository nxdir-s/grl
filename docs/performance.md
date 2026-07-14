# Performance

This document records the performance work done on the `performance` branch: what was slow, why each change was made, the alternatives that were rejected, and the measured impact. The raw benchmark runs backing every number here are committed under [`docs/benchmarks/`](benchmarks/) so the comparison can be reproduced with `benchstat`.

## Overview

Profiling the request lifecycle surfaced two dominant costs:

1. **Every request send performed ~4 disk file operations plus ~5 synchronous log writes.** Sending re-read `config.json` and the active environment file from disk, then re-read and fully rewrote `history.json` — each operation logging to an unbuffered file.
2. **The TUI redid expensive work it had already done.** A drag-resize re-colorized and re-wrapped the entire response body once per intermediate width; response search lowercased and rescanned the whole body on every keystroke; the key/value editor built its rows with O(n²) string concatenation every frame.

The work was done in benchstat-gated phases: benchmarks first, a baseline capture, then one independently verifiable change set per area. Several correctness bugs surfaced by the profiling (a dead timeout config, a retry that replayed empty bodies, a history-ordering bug) were folded into the phases that touched the same code.

### Headline results

| Hot path | Before | After | Delta |
| --- | --- | --- | --- |
| Drag-resize with a 1 MB response (30 resize events) | 3.40 s | 112 ms | −96.7% |
| Save history, 100 entries | 315 µs | 119 µs | −62% time, −65% bytes written |
| Load history, 100 entries | 433 µs | 325 µs | −25% |
| Config resolution on the send path | 13.8 µs + disk read | 25 ns (cache hit) | ~550× |
| Environment resolution on the send path | 20.0 µs + disk read | 221 ns (cache hit) | ~90× |
| List 10 collections (sidebar refresh) | 522 µs + 10 file reads | 6.6 µs | ~80× |
| `Substitute` — URL with 5 variables | 1.84 µs | 1.01 µs | −45% |
| `Substitute` — 10 KB body, no placeholders | 3.31 µs | 1.28 µs, 0 allocs | −61% |
| `SubstituteRequest` — full request | 18.2 µs | 11.3 µs | −38% |
| KV editor render, 100 rows | 1122 KiB/op | 322 KiB/op | −71% allocations |

## Methodology

- **Hardware/toolchain:** Apple M2 Pro, darwin/arm64, Go 1.26.2.
- **Benchmarks** use the modern `b.Loop()` idiom with `b.ReportAllocs()` everywhere and `b.SetBytes()` where throughput is meaningful. Coverage was extended before optimizing: the domain layer (`Substitute`, `ColorizeJSON`), TUI components (KV editor, viewport wrapping, search keystrokes, drag-resize), and a full root-model frame render (`BenchmarkTUIView`) all gained benchmarks in addition to the pre-existing secondary-adapter ones.
- **Capture:** `-count=10` repetitions per benchmark, before any production change ([`benchmarks/bench-baseline.txt`](benchmarks/bench-baseline.txt)) and after all changes ([`benchmarks/bench-after.txt`](benchmarks/bench-after.txt)). Deltas quoted here are benchstat results significant at p < 0.05; `~` rows (no significant difference) are reported as neutral.
- **Acceptance rule per phase:** targeted benchmarks must improve significantly; untargeted benchmarks must not regress more than 5%.
- All phases keep `go test -race ./...` green; races are covered by dedicated concurrent tests, not just luck.

> Note: on this development machine `go env GOOS` is persistently set to `linux`, so local test/bench commands are prefixed with `GOOS=darwin` (matching `.github/unit_tests.sh`).

## 1. Storage I/O on the request send path

### Problem

Tracing one `Send` through `internal/core/domain/requests.go`:

1. `environments.ActiveVars` → `configs.Get` → **read `config.json` + unmarshal**
2. …then `service.Get(activeEnvID)` → **read the environment file + unmarshal**
3. After the response, `History.Append` → **read all of `history.json` + unmarshal**
4. …append one entry → **`MarshalIndent` the whole slice + rewrite the whole file**

Nothing was cached, and every one of those operations emitted an `Info` log line to an **unbuffered** `os.File` — one `write` syscall each. Separately, `ListCollections`/`ListEnvironments` re-read and fully unmarshaled *every* file on *every* sidebar refresh, deserializing entire request bodies when the sidebar only renders names.

### Change

`internal/adapters/secondary/cache.go` adds `CachedStorage`, a write-through in-memory cache implementing `ports.Storage` and wrapping the JSON adapter (wired in `cmd/grl/main.go`). Warm sends now perform **zero** disk reads for config/environment resolution, and list calls serve clones from memory until a write invalidates them.

**Why a decorator on `ports.Storage` rather than memoization in the domain layer:**

- Every read *and* write of config/env/history/collections flows through the one storage instance, so write-through invalidation lives in a single type. Domain-level caches would have to know about each other — `Environments.SetActive` and `Environments.Delete` write *config*, so an `Environments` cache and a `Configs` cache would be coupled.
- The core stays free of sync primitives; `ports` and `service` interfaces are unchanged.
- bubbletea runs every `Cmd` in its own goroutine against this shared instance, so one `sync.RWMutex` in one place makes the whole surface race-safe.

**Why clone-on-read is mandatory, not defensive gold-plating:** callers mutate what storage hands them. `domain.Configs.Get` writes defaults into the returned config, `Environments.SetActive` mutates it and saves it back, and `domain.History.Load` reorders its slice in place. Serving interior pointers from the cache is an immediate `-race` failure, verified by tests that mutate returned values and assert the cache is unaffected. Collections are deep-cloned (they contain nested header/param slices); history needs only a shallow slice clone because entry fields are never mutated, only slice order.

**Deliberate non-goals:** `LoadCollection` passes through uncached — its result is mutated and re-saved by `domain.Collections`, and deep-cloning nested request slices on a user-action path isn't worth it. External edits to `~/.config/grl` mid-session aren't picked up until restart (acceptable staleness for a config directory owned by the app).

**Logging:** the 17 per-operation `Info` logs in the JSON adapter were demoted to `Debug` (`GRL_LOG=debug` re-enables them). This was chosen over wrapping the log file in a `bufio.Writer` because it needs no flush-on-exit/panic lifecycle — and cache hits eliminate most of the calls anyway. A warm send now writes zero log lines at the default level.

### Measured

| Operation | Disk (`JSONAdapter`) | Cached (`CachedStorage`) |
| --- | --- | --- |
| `LoadConfig` | 13.8 µs, 17 allocs | 24.7 ns, 1 alloc |
| `LoadEnvironment` (20 vars) | 20.0 µs, 85 allocs | 221 ns, 5 allocs |
| `LoadHistory` (100 entries) | 325 µs, 1424 allocs | 695 ns, 1 alloc |
| `ListCollections` (10 × 10 requests) | 522 µs, 1470 allocs | 6.6 µs, 111 allocs |

The cached numbers include the clone-on-read cost — that is the price of race safety, and it is still two to three orders of magnitude under the disk path.

## 2. History pipeline

### Problem

Performance profiling of the history path surfaced correctness bugs as much as cost:

- `DeleteEntry` loaded history through `Load` (which reverses into display order) and then **persisted the reversed slice**, flipping the on-disk order on every delete.
- `Append` was an unlocked read-modify-write; two concurrent request completions (bubbletea `Cmd` goroutines) could lose each other's entries.
- `SaveHistory` rejected an empty slice, so the last history entry could never be deleted.
- `HistoryEntry.Response.Body` is a `*bytes.Buffer`, which marshals as `{}` — bodies were silently dropped on save and came back empty, while still costing serialization of the wrapper.
- The machine-only `history.json` was written with `MarshalIndent` on every single request.

### Change

`internal/core/domain/history.go`:

- A `sync.Mutex` serializes `Append` and `DeleteEntry` end-to-end (verified by a 50-goroutine concurrent-append test under `-race`).
- Canonical on-disk order is **ascending by timestamp**. A `canonicalize` step sorts only when needed, which **idempotently heals files already flipped** by the old bug — no migration required.
- `Load` builds the newest-first view as a fresh slice instead of reversing in place; `DeleteEntry` filters into a new slice from storage order, never display order.

`internal/adapters/secondary/json.go`: history saves use compact `json.Marshal` (collections/environments/config keep indentation — those files are user-editable; history is not), and an empty history persists as `[]`.

`internal/core/entity/response.go`: `Body` is tagged `json:"-"`. This matches effective prior behavior (bodies already serialized as `{}` and the UI never reads a stored response body) and prevents a future 15 MB × 100-entry history file if buffer serialization were ever "fixed" naively.

**Why synchronous-write-under-lock rather than write-behind:** appends happen at human scale — one per manual send — and already run inside a `Cmd` goroutine, so a sub-millisecond write never blocks the render loop. Write-behind would add flush-on-quit plumbing, lose history on a crash or `kill -9`, and defer error surfacing, all for no observable gain. The durability property was verified directly: send, hard-kill the process, restart — the entry is present.

### Measured

| Benchmark | Before | After | Delta |
| --- | --- | --- | --- |
| `SaveHistory` 10 entries | 93.9 µs / 12.8 KiB | 50.8 µs / 4.8 KiB | −46% / −63% |
| `SaveHistory` 100 entries | 315 µs / 130 KiB | 119 µs / 45.5 KiB | −62% / −65% |
| `SaveHistory` 1000 entries | 2.74 ms / 1.63 MiB | 833 µs / 446 KiB | −70% / −73% |
| `LoadHistory` 100 entries | 433 µs | 325 µs | −25% |

The load win is a knock-on effect: compact files are smaller to read and parse. (Load throughput in MB/s drops in the report only because the files themselves shrank — absolute time is what matters here.)

## 3. TUI render path

### Problem

Three per-frame/per-event costs, one of them severe:

- **Resize:** `ResponseBodyPane.SetSize` called `refresh()` on every width change, re-running `ColorizeJSON` plus word-wrapping over the whole body. A terminal drag emits a stream of `WindowSizeMsg`, so a 1 MB response cost ~3.4 s of CPU per 30-event drag.
- **Search:** `recomputeMatches` executed `strings.ToLower(<whole body>)` — a full-body allocation and copy — on every search keystroke.
- **KV editor:** `View()` accumulated rows with `rows += row + "\n"`, O(n²) in row count, every frame while a header/param tab was focused.

### Change

- **Sequence-gated resize debounce.** Viewport geometry still updates synchronously (the panel resizes immediately); only the expensive re-colorize/re-wrap is deferred behind a `refreshPending` flag. The root model stamps each `WindowSizeMsg` with an incrementing sequence and schedules a 100 ms `tea.Tick`; only the tick whose sequence matches the latest resize calls `FlushResize()`, so a drag of any length costs exactly one refresh after settling. Accepted cosmetic cost: up to 100 ms of stale wrapping mid-drag. Tests assert N resizes + flush colorize exactly once, stale ticks are dropped, and height-only changes never refresh.
- **Cached search haystack.** Both response panes precompute a lowercased copy when content is set (`rawLower`), so per-keystroke matching allocates nothing for the haystack. Golden tests pin match counts, case-insensitivity, and the tab-switch path.
- **`strings.Builder`** replaces the string concatenation in the KV editor's row loop.

**Why full per-panel view caching was rejected.** The obvious "bigger" optimization — memoize each panel's rendered string behind dirty flags — was measured and declined: `BenchmarkTUIView` renders a full 200×60 frame with a populated sidebar and a 100 KB response in **~1.6 ms**, well under a ~5 ms responsiveness gate. Dirty-flag bookkeeping across ~15 components where one missed invalidation means visibly stale UI is a poor trade for headroom that already exists. The benchmark stays in the tree as the gate for revisiting this.

### Measured

| Benchmark | Before | After | Delta |
| --- | --- | --- | --- |
| `DragResize` 1 MB body (30 events + settle) | 3.40 s / 2.13 GiB | 112 ms / 72.9 MiB | −96.7% |
| `DragResize` 10 KB body | 34.5 ms | 1.14 ms | −96.7% |
| `KVEditorView` 100 rows | 2.50 ms / 1122 KiB | 2.27 ms / 322 KiB | −9% / −71% |
| `TUIView` full frame | 1.66 ms | 1.60 ms | −3.7% |

## 4. HTTP adapter correctness (perf-neutral by design)

Profiling the adapter surfaced bugs that were fixed in their own phase, deliberately last (they change request semantics):

- **The configured timeout was dead, twice over.** `NewHttpAdapter` pre-multiplied its default into nanoseconds but used a caller-supplied `cfg.Timeout` raw — so a user timeout of `30` meant 30 *nanoseconds*. And `main.go` never passed `Config.TimeoutSeconds` or `Config.FollowRedirects` into the adapter at all. The timeout is now a real `time.Duration` (documented as seconds), and both settings are wired from `config.json` at startup — which also warms the storage cache, since the config is loaded through it.
- **Retries replayed an empty body.** `bodyBytes, err := io.ReadAll(...)` inside an `if` shadowed the outer `bodyBytes`, so every retry sent nil. Fixed to assign; a test proves the replay (it fails against the old code).
- **Retry loop hardening:** a transport-level error left `resp == nil` and the loop dereferenced it (latent panic — retries are off by default); the 2xx check was `StatusCode/10 != 20`, which classified 226–299 as failures; and backoff was a bare `time.Sleep` inside `RoundTrip`. Now: nil-guarded drain, `StatusCode/100 != 2`, and a `select` on the request context so cancellation cuts the backoff short.

benchstat confirms neutrality: HTTP send benchmarks moved ±3%, within run-to-run noise (part of the after-capture overlapped verification jobs on the same machine), and comfortably inside the ≤5% regression budget.

**Behavior changes to be aware of:** redirects are now *followed* by default — the config default was always `true` but had never reached the adapter, which defaulted to blocking them; and the effective default timeout moves from a hardcoded 10 s to the config default of 30 s. Both are user-controllable in `config.json` and apply at startup.

## 5. Variable substitution

### Problem

`Substitute` used `ReplaceAllStringFunc` and then re-ran `FindStringSubmatch` on each already-matched substring — every placeholder was regex-scanned twice. Strings with no placeholders at all still paid a full regex scan whenever any variables were defined.

### Change

Single pass with `FindAllStringSubmatchIndex` writing through one `strings.Builder`, plus a `strings.Contains(str, "{{")` bail-out so placeholder-free text (most request bodies) never reaches the regex. Semantics are pinned by a 15-case golden test written against the *old* implementation before the swap: unknown variables stay untouched, `{{ x }}` whitespace works, adjacent placeholders, values containing `{{...}}` are not re-scanned, invalid names ignored.

### Measured

| Benchmark | Before | After | Delta |
| --- | --- | --- | --- |
| `Substitute` URL, 5 vars | 1.84 µs, 11 allocs | 1.01 µs, 7 allocs | −45% |
| `Substitute` 10 KB body with vars | 167 µs | 116 µs | −30% |
| `Substitute` 10 KB body, no placeholders | 3.31 µs, 3 allocs | 1.28 µs, 0 allocs | −61% |
| `SubstituteRequest` full request | 18.2 µs, 126 allocs | 11.3 µs, 99 allocs | −38% |

One honest trade: B/op rises slightly on tiny inputs (the match-index slices plus an upfront `Grow`), while time and alloc *count* drop across the board.

## Observability and guardrails added

- `GRL_CPUPROFILE=<path>` / `GRL_MEMPROFILE=<path>` write pprof profiles for a session (no effect when unset).
- `GRL_LOG=debug` re-enables per-operation storage logging in `~/.config/grl/out.log`.
- CI now runs `go test -race ./...` (making the new concurrency guarantees permanent) and bench-smokes **all** packages (`-benchtime=1x ./...`), so every benchmark stays compiling and running.

## Honest caveats and future work

- **Search keystroke cost barely moved (−0.6%).** `BenchmarkSearchKeystroke` types a 1-character pattern into a 1 MB body; with tens of thousands of matches, the dominant cost is the viewport's highlight processing, not the full-body lowercase copy that was removed. The eliminated per-keystroke allocation is real but small in that scenario. If short-pattern search on large bodies matters, the next lever is capping or deferring highlight computation below a minimum pattern length — a viewport-level change, out of scope here.
- **OAuth mode drops the client-level timeout.** `oauth2`'s client wrapper (`clientcredentials.Config.Client`) returns a new `http.Client` without copying `Timeout`; dialer/TLS timeouts still apply per connection. Pre-existing behavior, unchanged by this work.
- **Timeout/redirect config applies at startup.** The HTTP adapter is constructed once; changing `timeout_seconds` in the config modal takes effect on restart.
- **Per-panel view caching is deferred, not forgotten.** `BenchmarkTUIView` is the gate: revisit only if a future feature pushes the full-frame render past ~5 ms.
- **Mid-session external edits to `~/.config/grl` aren't seen** until restart, by design of the cache. An mtime check before serving cached config would be a cheap enhancement if this ever matters.

## Reproducing the measurements

```sh
# capture a run (10 repetitions; GOOS=darwin if your go env targets another OS)
go test -run='^$' -bench=. -benchmem -count=10 ./internal/... > bench-new.txt

# compare against the committed baseline/after artifacts
go install golang.org/x/perf/cmd/benchstat@latest
benchstat docs/benchmarks/bench-baseline.txt docs/benchmarks/bench-after.txt
benchstat docs/benchmarks/bench-after.txt bench-new.txt

# profile a live session
GRL_CPUPROFILE=cpu.out GRL_MEMPROFILE=mem.out grl
go tool pprof cpu.out
```

`bench-baseline.txt` was captured at commit `60424f2` (before any production change); `bench-after.txt` reflects the optimized tree with all phases applied. Benchmarks compare fairly only on the same hardware — capture your own baseline before optimizing further.
