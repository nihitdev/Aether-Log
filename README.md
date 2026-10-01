# Aether-Log
### A small, fast distributed log aggregation engine.

![Go 1.24+](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)

Aether-Log collects records over TCP and writes them to rotating local files.
It is a compact distributed logging experiment with a Go Hub, a Go Agent,
and small Go, C and Rust client SDKs. It uses only standard libraries and makes no
throughput or durable delivery guarantee.

## Architecture

```text
stdin / file → Agent ── v1 TCP ──→ Hub → bounded queue → batch writer → rotated logs
               Go/C/Rust SDKs ────────↗                  └── HTTP observability
```

See [architecture](docs/architecture.md) and the exact [wire protocol](docs/protocol.md).

## Features

- Bounded connections, payloads and ingestion queue; TCP backpressure.
- Configurable batch size, flush interval and size-based log rotation.
- Graceful reader cancellation and draining of accepted queued records.
- Agent reconnects with capped exponential backoff and cancellable timeouts.
- JSON metrics, liveness and readiness endpoints.
- Compatible v1 framing and dependency-free Go, POSIX C and Rust emitters.

## Quick start

Requires Go 1.24+; a POSIX C compiler is needed only for the C example.
Rust/Cargo is needed only for the Rust SDK targets and example.

```sh
git clone https://github.com/nihitdev/Aether-Log.git
cd Aether-Log
make build
./bin/aether-hub -log aether.log
```

In another terminal:

```sh
printf 'first record\nsecond record\n' | ./bin/aether-agent
curl http://127.0.0.1:8081/readyz
curl http://127.0.0.1:8081/metrics
# Allow the default one-second flush interval, then inspect the log.
tail aether.log
```

## Build and installation

`make build` writes ignored binaries under `bin/`. Alternatively:

```sh
go install gitlab.com/nihitdev/Aether-Log/cmd/hub@latest
go install gitlab.com/nihitdev/Aether-Log/cmd/agent@latest
```

Those install the repository revision available remotely; local work is built
with `make build`. Run binaries with `-h` for all flags. No release is claimed.

## Hub usage

```sh
./bin/aether-hub -listen 127.0.0.1:8080 -metrics 127.0.0.1:8081 \
  -log /tmp/aether.log -queue-capacity 1024 -batch-size 128 \
  -flush-interval 1s -rotate-bytes 67108864 -retain 5 \
  -max-connections 256 -max-payload 1048576 -read-timeout 1m
```

Defaults bind TCP `:8080` and HTTP `:8081` on all interfaces. Use loopback or a
trusted network. The output parent directory must already exist. Send SIGINT or
SIGTERM for graceful shutdown. Storage errors produce failed readiness and a
nonzero shutdown exit. Archives are `aether.log.1` through `aether.log.5`.
`-rotate-bytes 0` disables rotation; `-retain 0` discards old files on rotation.

## Agent usage and practical examples

```sh
./bin/aether-agent -hub 127.0.0.1:8080 -file application.log
printf '%s\n' '{"app":"payments","message":"started"}' | ./bin/aether-agent
./bin/aether-agent -file - -retry 200ms -max-retry 10s -connect-timeout 5s
```

Input is newline-delimited, with the final unterminated line accepted. The scanner
removes line endings (including CRLF). Empty lines are records. `-max-line`
defaults to 1 MiB and supports up to 16 MiB; match the Hub payload limit.
File mode reads to EOF; it does not follow a growing file. A disconnected Agent
retries its current record until sent or cancelled. SIGINT/SIGTERM stops it.

## Observability

| Endpoint | Meaning |
|---|---|
| `/healthz` | HTTP process liveness |
| `/readyz` | 200 when ready; 503 on shutdown or persistence failure |
| `/metrics`, `/stats` | JSON snapshot |

Snapshots include accepted/rejected/active connections, received frames and
payload bytes, queued/batched/persisted/dropped records, queue depth/capacity,
successful batches, write failures, readiness and uptime. JSON is not Prometheus
text format. See [counter semantics](docs/architecture.md).

## Batching and backpressure

The Hub writes batches rather than individual records. Batches flush at the
record count threshold or on a periodic timer, and remaining records flush on
shutdown. Rotation preserves a complete batch, so a file can exceed its size
threshold. A full queue blocks connection readers and eventually TCP senders.
Choose queue, batch, payload and connection limits together: record memory scales
with all four. The defaults are bounds, not a recommended memory budget for every
host. There is no durable queue or filesystem synchronization per batch.

## Protocol overview and clients

V1 has an 11-byte big-endian header: magic `0xAE744552`, version `1`, type `1`
(Data) or `2` (SlowDown), and a 32-bit payload length. Data contains opaque bytes.
The Hub appends a newline to each record. SlowDown is codec-supported but unused.
There are no acknowledgements or v2 changes. Read [protocol.md](docs/protocol.md)
before implementing a client.

- [Go SDK](sdk/go/client.go): `New`, cancellable `Send`, `Close`; capped reconnect
  backoff. Example: `go run ./sdk/go/example`.
- [C SDK](sdk/c/aether.h): POSIX connect/send/close, blocking sockets, no retries.
  Build: `make -C sdk/c`; run `sdk/c/example 127.0.0.1 8080 'hello from C'`.
  Send timeouts are configured; DNS/connect timing is platform-dependent. On
  platforms without `MSG_NOSIGNAL`, callers must handle/ignore SIGPIPE.
- [Rust SDK](sdk/rust/README.md): synchronous `AetherClient::connect`, `send`,
  `close`; configurable connect/write timeouts and payload limits, no retries.
  Build/check with `make sdk-rust` and `make sdk-rust-test`.

With the Hub running, send a Rust record from the repository root:

```sh
cargo run --manifest-path sdk/rust/Cargo.toml --example send -- "hello from rust"
# Optional second argument selects a Hub address.
cargo run --manifest-path sdk/rust/Cargo.toml --example send -- "another record" 127.0.0.1:8080
```

A successful client write is not a persistence acknowledgement. Failed-send
retries may duplicate records; crashes and undetected disconnects may lose them.
No exactly-once or at-least-once guarantee is provided.

## Project structure

```text
cmd/{hub,agent}/    Go applications
internal/protocol/  v1 codec and benchmarks
internal/ingest/    bounded queue and batch worker
internal/buffer/    optional pooled buffer helper
pkg/{storage,metrics}/
sdk/{go,c,rust}/   client SDKs and examples
docs/              architecture and wire contract
site/              React + Vite project website (pnpm)
```

## Project website

The responsive React + Vite website lives in `site/`. It includes interactive
architecture and SDK examples, plus local documentation rendered from the
repository's Markdown. Run it independently from the Go applications:

```sh
cd site
pnpm install --frozen-lockfile
pnpm dev
# Production bundle: pnpm build
```

See [site/README.md](site/README.md) for requirements and commands.

## Development and tests

```sh
make fmt
make check       # tests, race tests, vet, builds
make bench       # actual benchmarks; no published performance claims
make run-hub
printf 'example\n' | make run-agent
```

Tests cover malformed/truncated frames, limits, batching, interval flushing,
queue cancellation, draining, persistence failures, rotation, HTTP metrics and
Go client sends/retry cancellation. Protocol and batching benchmarks are
included; run them on your own hardware. `make clean` removes build artifacts.

## Limitations and roadmap

There is no authentication, TLS, durable client spool, query/index engine,
acknowledgement, metadata envelope or compression. HTTP exposes operational
information without access control. One process must own each output path.
A stalled filesystem can delay draining. Logs may contain arbitrary binary bytes.

Future work should be driven by real use: durable acknowledgement/spool design,
source metadata with explicit compatibility, clean file-follow semantics,
and deployment guidance. Go, C and Rust SDKs are implemented.
The Go Hub remains canonical.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for validation and compatibility expectations,
and [SECURITY.md](SECURITY.md) for deployment scope and reporting. No license file
exists in the repository; no license or CI status badge is claimed.
