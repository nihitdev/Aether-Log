# Aether-Log Rust client

A small synchronous v1 emitter using only Rust's standard library. No Tokio,
async runtime, external dependencies, automatic retries or server implementation.
The crate is local and not published to crates.io.

## Usage

Add a path dependency in your application's `Cargo.toml` (relative to that file):

```toml
[dependencies]
aether-log-client = { path = "../Aether-Log/sdk/rust" }
```

```rust,no_run
use aether_log_client::AetherClient;

fn main() -> std::io::Result<()> {
    let mut client = AetherClient::connect("127.0.0.1:8080")?;
    client.send(b"hello from rust")?;
    client.send(b"another record")?;
    client.close()
}
```

`connect_with_options(address, ClientOptions)` configures connection timeout,
write timeout and maximum payload. Both timeouts default to five seconds and
must be nonzero. The payload limit defaults to 1 MiB, matching the Go Hub;
`max_payload: 0` permits only empty records.

```rust,no_run
use aether_log_client::{AetherClient, ClientOptions};
use std::time::Duration;

let options = ClientOptions {
    connect_timeout: Duration::from_secs(2),
    write_timeout: Duration::from_secs(3),
    max_payload: 1024 * 1024,
};
let mut client = AetherClient::connect_with_options("127.0.0.1:8080", options)?;
client.send(b"configured client")?;
client.close()?;
# Ok::<(), std::io::Error>(())
```

Methods return `std::io::Result`. Oversized payloads return `InvalidInput`
before writing; the connection remains usable. Sending after close returns
`NotConnected`. `close` shuts down both directions and drops the socket;
repeated calls succeed. Dropping a client also releases its socket. On a write
error the connection is invalidated, even if some bytes were transmitted. Create
a new client to reconnect; decide whether to retry based on your delivery needs.
Sending requires `&mut self`; the client does not spawn threads or buffer records.

## Build, tests and example

From the repository root:

```sh
cargo build --manifest-path sdk/rust/Cargo.toml --release
cargo fmt --manifest-path sdk/rust/Cargo.toml --check
cargo test --manifest-path sdk/rust/Cargo.toml
cargo clippy --manifest-path sdk/rust/Cargo.toml --all-targets -- -D warnings
```

Or use `make sdk-rust` and `make sdk-rust-test`. These are optional: ordinary Go
builds and checks do not invoke Cargo. A Rust toolchain with rustfmt and Clippy
is needed for the full checks. Tests were validated with Rust 1.98.1.

Start the Go Hub (`make run-hub`), then from `sdk/rust`:

```sh
cargo run --example send -- "hello from rust"
cargo run --example send -- "custom address" 127.0.0.1:8080
```

From the repository root, use
`cargo run --manifest-path sdk/rust/Cargo.toml --example send -- "hello from rust"`.
The example defaults to `127.0.0.1:8080`, accepts `[message] [hub-address]`, and
sends one record. Wait for the Hub's batch flush before inspecting `aether.log`.

Tests cover literal header bytes, empty/binary/normal payloads, configured and
wire length limits (without allocating 4 GiB), partial/interrupted/zero writes,
zero timeout errors, multiple frames on a real local TCP connection, validation
without corrupting that connection, write-error connection invalidation, and
close behavior.

## Protocol compatibility and limitations

The unchanged [v1 protocol](../../docs/protocol.md) has an 11-byte header:
magic `0xAE744552`, version `1`, type `1` (Data), and a u32 byte length, all
multibyte integers big endian. Payloads are opaque bytes; empty records are valid.
No newline is appended by this client; the Hub appends one on disk. The client
enforces both its configured limit and the u32 wire maximum. Match the configured
limit to the Hub's `-max-payload` value.

Type `2` (SlowDown) remains defined by v1 but is not emitted or read by this
client. The current Hub does not send control frames. No protocol changes,
acknowledgements, TLS, authentication, automatic reconnect, durable spool or
exactly-once guarantee are provided. Successful `send`/`close` does not prove
Hub acceptance or persistence. A failed write can still have transmitted a
partial or complete record; retries can duplicate it.

DNS resolution is synchronous and not covered by the connection timeout. That
timeout applies separately to each resolved address, so total connection time
can exceed it. The write timeout applies per blocking socket write, not as a
deadline for the complete frame; timeout error kinds vary by platform. There is
no cancellation token or total operation deadline.

`target/` is ignored. `Cargo.lock` is intentionally retained for reproducible
repository builds/tests/examples. This does not pin dependency resolution for
applications using the SDK as a dependency. No third-party dependencies are
currently present. The repository is licensed under Apache License 2.0; see
[LICENSE](../../LICENSE).
