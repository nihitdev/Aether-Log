# Aether wire protocol (v1)

TCP is a byte stream: a frame may span reads, and multiple frames may arrive in
one read. Read exactly 11 header bytes, validate them, then read exactly Length
payload bytes. All multibyte integers are unsigned and big endian.

| Offset | Bytes | Field | Value |
|---|---:|---|---|
| 0 | 4 | Magic | `0xAE744552` (bytes `AE 74 45 52`) |
| 4 | 2 | Version | `1` (`00 01`) |
| 6 | 1 | Type | `1` Data, `2` SlowDown |
| 7 | 4 | Length | Payload byte count; excludes header |
| 11 | Length | Payload | Type-specific bytes |

Data is one opaque record, including a possible empty record. No character
encoding is enforced. Clients typically send UTF-8 without a trailing newline.
The Hub appends one newline on disk; embedded newlines and binary data are
preserved, so the persisted file does not necessarily have one line per record.
Example: `hi` is `AE 74 45 52 00 01 01 00 00 00 02 68 69`.

SlowDown has exactly four payload bytes: an unsigned delay in milliseconds.
The codec supports it for compatibility; the current Hub does not emit it and
current SDKs do not consume it. The Hub accepts only Data from clients and closes
connections receiving control frames. Do not depend on SlowDown for flow control.

The length field supports up to 2^32−1 bytes. The Hub defaults to a 1 MiB limit,
configurable with `-max-payload`. The Go and Rust SDKs and Agent also default to 1 MiB.
The C SDK validates the wire length limit; callers must respect the Hub limit.
The Hub rejects excessive lengths before allocating their payload. Invalid magic,
version, type, truncated frames and read timeouts close the connection. There is
no resynchronization after malformed input. Configure client limits consistently.

## Backpressure and compatibility

The bounded ingestion queue blocks connection readers when full. TCP flow control
then slows clients. Connection slots and frame read deadlines are bounded.
Queue depth counts waiting records, excluding the batch and one decoded record
per active connection. Approximate record memory is bounded by
`(queue capacity + batch size + max connections) × max payload`, plus framing,
buffer capacity, socket buffers and runtime overhead. Choose limits for your RAM.

The v1 wire layout and frame types are unchanged. There is no v2, batch frame,
metadata envelope or acknowledgement in this implementation. Send multiple v1
frames to send multiple records; storage batching is independent of framing.
Clients must not send new versions or types without negotiating a future change.

## Delivery semantics

A successful send means the local TCP write completed, not that the Hub accepted,
persisted or synchronized the record. A disconnect after any partial or complete
write is ambiguous. The Go client retries the current record on a detected write
failure, which can duplicate it. The C and Rust clients do not automatically retry;
the Rust client invalidates its connection after a write error to prevent reuse
after a potentially truncated frame. Undetected disconnects, process crashes, queued
records, storage errors and cancellation can lose records. There is neither an
at-least-once nor an exactly-once guarantee, no durable client spool and no replay
cursor. Ordering is preserved per TCP stream; concurrent streams interleave.

Graceful Hub shutdown closes readers, waits for them, drains accepted queued
records and closes storage. A partial frame is not accepted. Failed batches are
counted as dropped; a partial filesystem write may nevertheless have persisted a
prefix. Successful batch writes count as persisted but are not `fsync` durability
promises. Storage failures make readiness fail; restart after resolving the error.
