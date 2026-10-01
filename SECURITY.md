# Security

This experiment has no TLS, authentication or authorization. Bind listeners to
loopback or use a trusted isolated network. Limit payload size, connections and
queue capacity to fit the host. The HTTP endpoints reveal operational metrics.
Log bytes are untrusted; avoid interpreting them as terminal commands or HTML.

No private security contact or supported release policy is established in this
repository. For non-sensitive hardening issues, use the repository issue tracker.
Do not put secrets or exploit details in a public issue; coordinate a private
reporting channel with the repository owner first.
