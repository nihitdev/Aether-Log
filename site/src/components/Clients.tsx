import { useRef, useState } from 'react'
import { clients } from '../content'
import { CodeBlock } from './CodeBlock'

export function Clients() {
  const [active, setActive] = useState(0)
  const tabs = useRef<(HTMLButtonElement | null)[]>([])
  const client = clients[active]
  return (
    <section className="section wrap clients-section" id="clients">
      <div className="section-heading">
        <div>
          <p className="eyebrow">03 / SPEAK YOUR LANGUAGE</p>
          <h2>
            Three clients.
            <br />
            <span className="muted">The same small protocol.</span>
          </h2>
        </div>
        <p>
          A canonical Go Hub. Small Go, C, and Rust emitters. Pick the client that fits your
          application.
        </p>
      </div>
      <div className="client-grid">
        <div className="client-intro">
          <div
            className="sdk-tabs"
            role="tablist"
            aria-label="Client language"
            aria-orientation="vertical"
          >
            {clients.map((item, i) => (
              <button
                key={item.name}
                ref={(element) => {
                  tabs.current[i] = element
                }}
                role="tab"
                id={`tab-${item.name.toLowerCase()}`}
                aria-controls="sdk-panel"
                aria-selected={i === active}
                tabIndex={i === active ? 0 : -1}
                onClick={() => setActive(i)}
                onKeyDown={(event) => {
                  let next: number
                  if (event.key === 'ArrowDown' || event.key === 'ArrowRight')
                    next = (i + 1) % clients.length
                  else if (event.key === 'ArrowUp' || event.key === 'ArrowLeft')
                    next = (i + clients.length - 1) % clients.length
                  else if (event.key === 'Home') next = 0
                  else if (event.key === 'End') next = clients.length - 1
                  else return
                  event.preventDefault()
                  setActive(next)
                  tabs.current[next]?.focus()
                }}
              >
                {item.name}
                <span aria-hidden="true">→</span>
              </button>
            ))}
          </div>
          <p>{client.description}</p>
          <a
            className="text-link"
            href={client.name === 'Rust' ? '?page=docs&doc=rust' : '?page=docs#clients'}
          >
            {client.name} client guide <span aria-hidden="true">↗</span>
          </a>
          <div className="client-run">
            <span className="eyebrow">TRY THE EXAMPLE</span>
            <code>{client.command}</code>
          </div>
        </div>
        <div
          id="sdk-panel"
          role="tabpanel"
          aria-labelledby={`tab-${client.name.toLowerCase()}`}
          tabIndex={0}
        >
          <CodeBlock key={client.name} filename={client.filename} code={client.code} />
        </div>
      </div>
    </section>
  )
}
