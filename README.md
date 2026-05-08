# Aether-Log

A high-performance, distributed log aggregation system written in Go.

## Architecture

Aether-Log consists of two main components:

1.  **Log-Hub (Server):** The central aggregator that receives log frames over TCP, manages concurrent connections with a semaphore, and persists logs using a fan-in pattern.
2.  **Log-Collector (Agent):** A lightweight agent that collects logs locally and transmits them to the Hub using a custom binary protocol.

### Key Features

*   **Custom Binary Protocol:** Low-overhead framing with magic numbers, versioning, and length-prefixed payloads.
*   **Zero-Allocation Pipeline:** Uses `sync.Pool` for byte buffer reuse to minimize Garbage Collector impact.
*   **Concurrency Control:** Goroutine-per-connection model with a global semaphore to prevent resource exhaustion.
*   **Backpressure Mechanism:** Hub can signal agents to slow down using `TypeSlowDown` control frames.
*   **Observability:** Built-in HTTP metrics API (`:8081`) for real-time monitoring.

## Protocol Specification

Each frame consists of an 11-byte header followed by an optional payload:

| Offset | Size | Field | Description |
| :--- | :--- | :--- | :--- |
| 0 | 4 | Magic Number | `0xAE744552` |
| 4 | 2 | Version | `1` |
| 6 | 1 | Type | `1` (Data), `2` (SlowDown) |
| 7 | 4 | Length | Payload length (BigEndian) |
| 11 | N | Payload | Binary log data |

## Getting Started

### Prerequisites

*   Go 1.24 or later

### Installation

1. Clone the repository:
   ```bash
   git clone github.com/gemini-cli/aether-log
   cd aether-log
   ```

2. Build the binaries:
   ```bash
   # Build the Hub
   go build -o hub ./cmd/hub

   # Build the Agent
   go build -o agent ./cmd/agent
   ```

### Running

1. Start the Log-Hub:
   ```bash
   ./hub
   ```
   The Hub listens for logs on `:8080` and metrics on `:8081`.

2. Start the Log-Collector:
   ```bash
   ./agent
   ```

3. Check Metrics:
   ```bash
   curl http://localhost:8081/
   ```

4. View Logs:
   ```bash
   tail -f aether.log
   ```

## Project Structure

*   `/cmd`: Application entry points for Hub and Agent.
*   `/internal`: Private library code (Protocol, Dispatcher, Agent logic).
*   `/pkg`: Public library code (Storage, Metrics).
