export const repository = 'https://github.com/nihitdev/Aether-Log'
export const quickstart = `# 1. Get the source and build
git clone https://github.com/nihitdev/Aether-Log.git
cd Aether-Log
make build

# 2. Start the Hub
./bin/aether-hub -listen 127.0.0.1:8080 \\
  -metrics 127.0.0.1:8081 -log aether.log

# 3. In a second terminal, send a record
printf 'hello aether\\n' | ./bin/aether-agent`

export const stages = [
  {
    name: 'Agent',
    sub: 'stdin / regular file',
    description:
      'Read newline-delimited records from stdin or a regular file. Reconnect with capped exponential backoff when a write fails.',
  },
  {
    name: 'Protocol',
    sub: '11-byte v1 header',
    description:
      'Send an opaque record in a length-prefixed TCP frame. Magic, version, type, and payload length use a precise, documented binary layout.',
  },
  {
    name: 'Hub + queue',
    sub: 'bounded ingestion',
    description:
      'One reader per connection feeds a bounded queue. Limit payloads and connections; when the queue fills, readers block and TCP backpressure slows producers.',
  },
  {
    name: 'Batch writer',
    sub: 'size / interval',
    description:
      'One worker combines records and writes a batch when its record limit is reached or its timer ticks. Graceful shutdown drains accepted queued records.',
  },
  {
    name: 'Storage',
    sub: 'rotating local files',
    description:
      'Append records to local files and rotate before a batch exceeds the configured size. Retain a configurable number of archives; keep each batch intact.',
  },
]

export const clients = [
  {
    name: 'Go',
    filename: 'main.go',
    description:
      'A reusable client with cancellable sends and capped exponential reconnect delays.',
    command: 'go run ./sdk/go/example',
    code: `package main

import (
    "context"
    "log"
    "time"
    client "github.com/nihitdev/Aether-Log/sdk/go"
)

func main() {
    c, err := client.New(client.Config{
        Address: "127.0.0.1:8080",
    })
    if err != nil { log.Fatal(err) }
    defer c.Close()
    ctx, cancel := context.WithTimeout(
        context.Background(), 10*time.Second,
    )
    defer cancel()
    if err := c.Send(ctx, []byte("hello from Go")); err != nil {
        log.Print(err)
    }
}`,
  },
  {
    name: 'C',
    filename: 'send.c',
    description:
      'A tiny POSIX emitter with a connect/send/close API. Blocking sockets, configured send timeouts, and no automatic retries.',
    command: "make -C sdk/c\nsdk/c/example 127.0.0.1 8080 'hello from C'",
    code: `#include "aether.h"
#include <stdio.h>
#include <string.h>

int main(void) {
    const char *message = "hello from C";
    int fd = aether_connect("127.0.0.1", "8080", 5);
    if (fd < 0) {
        perror("connect");
        return 1;
    }
    int result = aether_send(fd, message, strlen(message));
    if (result < 0) perror("send");
    aether_close(fd);
    return result < 0;
}`,
  },
  {
    name: 'Rust',
    filename: 'send.rs',
    description:
      'An idiomatic synchronous client built on std::net. Reuse a connection, configure timeouts, and send without an async runtime.',
    command:
      'cargo run --manifest-path sdk/rust/Cargo.toml \\\n  --example send -- "hello from rust"',
    code: `use aether_log_client::{AetherClient, ClientOptions};
use std::time::Duration;

fn main() -> std::io::Result<()> {
    let options = ClientOptions {
        connect_timeout: Duration::from_secs(2),
        write_timeout: Duration::from_secs(3),
        max_payload: 1024 * 1024,
    };
    let mut client = AetherClient::connect_with_options(
        "127.0.0.1:8080", options,
    )?;

    client.send(b"hello from rust")?;
    client.send(b"another record")?;
    client.close()
}`,
  },
]
