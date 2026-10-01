import { Children, isValidElement, useEffect, type ComponentPropsWithoutRef } from 'react'
import Markdown, { type ExtraProps } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import architecture from '../../../docs/architecture.md?raw'
import protocol from '../../../docs/protocol.md?raw'
import rust from '../../../sdk/rust/README.md?raw'
import { CodeBlock } from './CodeBlock'
import { quickstart } from '../content'

const overview = `# Start with Aether

A small distributed logging experiment with a canonical Go Hub and Agent, plus Go, C and Rust clients. All core implementations use standard libraries.

## Quick start

Install Go 1.24+ and Make. Run these commands in your working directory:

\`\`\`sh
${quickstart}
\`\`\`

The Hub writes batches at a record-count threshold or on its periodic timer. Wait for the default one-second flush, then inspect \`aether.log\`.

## Hub configuration

Bind to loopback or a trusted network. The default listeners bind on all interfaces; the commands above explicitly use loopback. Output directories must already exist.

| Flag | Default | Purpose |
| --- | --- | --- |
| \`-listen\` | \`:8080\` | TCP ingestion address |
| \`-metrics\` | \`:8081\` | HTTP observability address |
| \`-log\` | \`aether.log\` | Active output file |
| \`-queue-capacity\` | \`1024\` | Waiting records |
| \`-batch-size\` | \`128\` | Records per batch |
| \`-flush-interval\` | \`1s\` | Batch timer interval |
| \`-max-connections\` | \`256\` | Concurrent connections |
| \`-max-payload\` | \`1048576\` | Payload limit in bytes |
| \`-read-timeout\` | \`1m\` | Frame/idle read timeout |
| \`-rotate-bytes\` | \`67108864\` | Rotation threshold; zero disables |
| \`-retain\` | \`5\` | Retained archives |

Send SIGINT or SIGTERM to stop readers and drain accepted queued records. A whole batch can exceed the rotation threshold. Persistence is not synchronized with fsync after each batch.

## Agent input

Omit \`-file\` or use \`-file -\` for stdin. Regular files are read to EOF; file-follow is not implemented. Empty lines and final unterminated lines are accepted; CRLF endings are removed.

\`\`\`sh
./bin/aether-agent -file application.log
printf 'a record\\n' | ./bin/aether-agent
./bin/aether-agent -retry 200ms -max-retry 10s -connect-timeout 5s
\`\`\`

The Agent retries the current record with capped exponential reconnect delays until sent or cancelled. The default line limit is 1 MiB. Match the Agent and Hub payload limits.

## Clients

The supported SDK set is Go, C and Rust. Clients emit v1 frames; the Go Hub remains the canonical server.

- **Go:** \`New\`, cancellable \`Send\`, \`Close\`; capped reconnect backoff. Run \`go run ./sdk/go/example\`.
- **C:** POSIX \`aether_connect\`, \`aether_send\`, \`aether_close\`. Build with \`make -C sdk/c\`. Run \`sdk/c/example 127.0.0.1 8080 'hello from C'\`. No retries. Send timeouts are configured; DNS/connect timing is platform-dependent. Where MSG_NOSIGNAL is unavailable, handle or ignore SIGPIPE.
- **Rust:** synchronous \`AetherClient::connect\`, \`send\`, \`close\`. Optional \`ClientOptions\` sets connection/write timeouts and payload limits. No retries or async runtime. See the Rust guide in the navigation.

## Observability

\`\`\`sh
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8081/readyz
curl http://127.0.0.1:8081/metrics
\`\`\`

\`/healthz\` reports HTTP liveness. \`/readyz\` returns 200 when ready and 503 on shutdown or persistence failure. \`/metrics\` and \`/stats\` return JSON with received frames/bytes, connections, queue depth/capacity, queued/batched/persisted/dropped records, write failures, batches and uptime. These are not Prometheus text endpoints.

## Delivery and security

Successful TCP writes do not prove Hub acceptance or durable persistence. Records may be lost; retries may duplicate them. No exactly-once or at-least-once guarantee, acknowledgements, durable spool, TLS, authentication or query engine is provided. Keep listeners on loopback or a trusted isolated network. One process must own each output path.

Read the protocol guide for the precise wire contract and delivery semantics.
`
const guides = [
  { key: 'overview', title: 'Getting started', text: overview },
  { key: 'architecture', title: 'Architecture', text: architecture },
  { key: 'protocol', title: 'Wire protocol', text: protocol },
  { key: 'rust', title: 'Rust client', text: rust },
]

function Heading({ children, node: _node, ...props }: ComponentPropsWithoutRef<'h2'> & ExtraProps) {
  const id = Children.toArray(children)
    .join('')
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .trim()
    .replace(/\s+/g, '-')
  return (
    <h2 id={id} {...props}>
      {children}
      <a
        className="heading-anchor"
        href={`#${id}`}
        aria-label={`Link to ${Children.toArray(children).join('')}`}
      >
        #
      </a>
    </h2>
  )
}

function Pre({ children }: ComponentPropsWithoutRef<'pre'>) {
  const child = Children.only(children)
  if (isValidElement<{ children: string; className?: string }>(child)) {
    const language = child.props.className?.replace('language-', '') || 'example'
    return <CodeBlock code={String(child.props.children).replace(/\n$/, '')} filename={language} />
  }
  return <pre>{children}</pre>
}

function DocLink({
  href,
  children,
  node: _node,
  ...props
}: ComponentPropsWithoutRef<'a'> & ExtraProps) {
  if (href?.includes('docs/protocol.md')) href = '?page=docs&doc=protocol'
  if (href?.includes('docs/architecture.md')) href = '?page=docs&doc=architecture'
  return (
    <a href={href} {...props}>
      {children}
    </a>
  )
}

export default function Docs() {
  const key = new URLSearchParams(window.location.search).get('doc') ?? 'overview'
  const guide = guides.find((item) => item.key === key) ?? guides[0]
  useEffect(() => {
    document.title = `${guide.title} — Aether-Log`
    if (window.location.hash) {
      const id = decodeURIComponent(window.location.hash.slice(1))
      document.getElementById(id)?.scrollIntoView()
    }
  }, [guide.title])
  return (
    <main id="main" className="docs-layout wrap">
      <aside className="docs-sidebar">
        <p className="eyebrow">DOCUMENTATION</p>
        <nav aria-label="Documentation">
          {guides.map((item) => (
            <a
              href={`?page=docs&doc=${item.key}`}
              key={item.key}
              aria-current={guide.key === item.key ? 'page' : undefined}
            >
              {item.title}
              <span aria-hidden="true">↗</span>
            </a>
          ))}
        </nav>
        <div className="docs-sidebar-note">
          <span className="status-dot" />
          Protocol v1
          <br />
          <span>Go · C · Rust</span>
        </div>
        <a className="text-link" href="./">
          ← Back to project
        </a>
      </aside>
      <article className="docs-content">
        <p className="eyebrow">AETHER-LOG / {guide.title.toUpperCase()}</p>
        <Markdown remarkPlugins={[remarkGfm]} components={{ h2: Heading, pre: Pre, a: DocLink }}>
          {guide.text}
        </Markdown>
        <div className="docs-end">Keep the path simple. Know what your logs can guarantee.</div>
      </article>
    </main>
  )
}
