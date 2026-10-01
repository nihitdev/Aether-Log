# Contributing

Keep changes small and explain the concrete behavior they improve. Preserve v1
wire compatibility or document and test a deliberate versioned extension.
Go is the canonical Hub and Agent implementation; SDKs should only emit frames.
Prefer standard libraries and document actual delivery semantics.

Before proposing a change, run `make fmt`, `make check`, and `make -C sdk/c`.
For networking/storage changes, run a real Hub/client smoke test with temporary
ports and files and verify SIGTERM drains accepted records. Include relevant tests
and describe validation in the merge request. Do not commit generated binaries,
logs or benchmark output. There is currently no repository license file; clarify
licensing with the repository owner before redistributing code.
