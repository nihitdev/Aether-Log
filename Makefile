.RECIPEPREFIX := >
.PHONY: all build test test-race vet fmt check bench clean run-hub run-agent sdk-rust sdk-rust-test sdk-rust-clean
all: check
build:
>mkdir -p bin
>go build -trimpath -o bin/aether-hub ./cmd/hub
>go build -trimpath -o bin/aether-agent ./cmd/agent
test:
>go test ./...
test-race:
>go test -race ./...
vet:
>go vet ./...
fmt:
>gofmt -w cmd internal pkg sdk/go
check: test test-race vet build
bench:
>go test -run '^$$' -bench . -benchmem ./internal/...
clean:
>rm -f bin/aether-hub bin/aether-agent
>$(MAKE) -C sdk/c clean
run-hub: build
>./bin/aether-hub
run-agent: build
>./bin/aether-agent

# Optional Rust SDK targets; ordinary Go checks do not require Cargo.
sdk-rust:
>cargo build --manifest-path sdk/rust/Cargo.toml --release
sdk-rust-test:
>cargo fmt --manifest-path sdk/rust/Cargo.toml --check
>cargo test --manifest-path sdk/rust/Cargo.toml
>cargo clippy --manifest-path sdk/rust/Cargo.toml --all-targets -- -D warnings
sdk-rust-clean:
>cargo clean --manifest-path sdk/rust/Cargo.toml
