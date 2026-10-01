export function Features() {
  return (
    <>
      <section className="features wrap" aria-label="Core features">
        <article>
          <span className="feature-icon" aria-hidden="true">
            ⇥
          </span>
          <h3>Know your bounds.</h3>
          <p>
            Limit connections, payloads, and queue capacity. When the queue fills, TCP backpressure
            slows producers.
          </p>
          <span className="feature-tag">CONTROL, NOT GUESSWORK</span>
        </article>
        <article>
          <span className="feature-icon" aria-hidden="true">
            ▤
          </span>
          <h3>Write together.</h3>
          <p>
            One worker batches records by count or time, then writes to size-rotated files. Accepted
            records drain on shutdown.
          </p>
          <span className="feature-tag">BATCHED PERSISTENCE</span>
        </article>
        <article>
          <span className="feature-icon" aria-hidden="true">
            ⌁
          </span>
          <h3>See what's happening.</h3>
          <p>
            Health, readiness, and JSON metrics give you a view into queue depth, connections,
            writes, and dropped records.
          </p>
          <span className="feature-tag">VISIBILITY BUILT IN</span>
        </article>
      </section>
    </>
  )
}
