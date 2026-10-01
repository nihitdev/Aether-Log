import { useState } from 'react'
import { stages } from '../content'

export function Architecture() {
  const [active, setActive] = useState(0)
  return (
    <section className="section wrap" id="architecture">
      <div className="section-heading">
        <div>
          <p className="eyebrow">01 / THE ARCHITECTURE</p>
          <h2>
            One pipeline.
            <br />
            <span className="muted">Every step has a purpose.</span>
          </h2>
        </div>
        <p>
          Collect, frame, queue, persist. Aether keeps the path short and the moving parts visible.
        </p>
      </div>
      <div className="pipeline" role="group" aria-label="Explore pipeline stages">
        {stages.map((stage, i) => (
          <button
            key={stage.name}
            className={`pipeline-step ${i === active ? 'selected' : ''}`}
            onClick={() => setActive(i)}
            aria-pressed={i === active}
          >
            <span className="step-number">0{i + 1}</span>
            <strong>{stage.name}</strong>
            <span>{stage.sub}</span>
            {i < stages.length - 1 ? <i aria-hidden="true">→</i> : null}
          </button>
        ))}
      </div>
      <div className="pipeline-detail">
        <span className="detail-label">{stages[active].name.toUpperCase()}</span>
        <p aria-live="polite">{stages[active].description}</p>
        <a href="?page=docs&doc=architecture">
          Explore the architecture <span aria-hidden="true">↗</span>
        </a>
      </div>
    </section>
  )
}
