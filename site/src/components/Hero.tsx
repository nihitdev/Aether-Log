export function Hero() {
  return (
    <>
      <section className="hero wrap">
        <div className="hero-copy">
          <p className="eyebrow">
            <span className="status-dot"></span> SMALL BY DESIGN. OPEN TO EXPERIMENT.
          </p>
          <h1>
            A clear path
            <br />
            for your <span className="serif">logs.</span>
          </h1>
          <p className="hero-description">
            From scattered output to one place.
            <br />A small distributed logging engine with a Go core, bounded ingestion, and a
            protocol you can understand.
          </p>
          <div className="hero-actions">
            <a className="button primary" href="#quickstart">
              Start collecting <span aria-hidden="true">↗</span>
            </a>
            <a className="text-link" href="?page=docs">
              Read the docs <span aria-hidden="true">→</span>
            </a>
          </div>
          <p className="hero-note">
            <span aria-hidden="true">◈</span> Standard libraries. Three clients. One simple
            pipeline.
          </p>
        </div>
        <div
          className="hero-visual"
          aria-label="Illustration of records flowing from clients through a Hub to a log file"
        >
          <div className="visual-top">
            <span>
              <span className="status-dot"></span> THE RECORD JOURNEY
            </span>
            <span>01 — 03</span>
          </div>
          <div className="source-row">
            <div className="source-chip">
              Go <span>01</span>
            </div>
            <div className="source-chip">
              C <span>02</span>
            </div>
            <div className="source-chip">
              Rust <span>03</span>
            </div>
          </div>
          <div className="flow-wires" aria-hidden="true">
            <span></span>
            <span></span>
            <span></span>
            <i className="packet p1"></i>
            <i className="packet p2"></i>
            <i className="packet p3"></i>
          </div>
          <div className="hub-node">
            <div className="hub-symbol" aria-hidden="true">
              ↳
            </div>
            <div>
              <strong>Aether Hub</strong>
              <span>RECEIVE · QUEUE · BATCH</span>
            </div>
            <span className="hub-led" aria-hidden="true"></span>
          </div>
          <div className="output-wire" aria-hidden="true">
            <i className="packet"></i>
          </div>
          <div className="mini-terminal">
            <div className="terminal-bar">
              <span className="terminal-dots" aria-hidden="true">
                ● ● ●
              </span>
              <span>aether.log</span>
              <span aria-hidden="true">↙</span>
            </div>
            <div className="terminal-lines">
              <p>
                <span>01</span> application started
              </p>
              <p>
                <span>02</span> request received
              </p>
              <p>
                <span>03</span> job completed <i className="cursor" aria-hidden="true"></i>
              </p>
            </div>
          </div>
          <div className="visual-bottom">
            <span>TCP / PROTOCOL V1</span>
            <span>ILLUSTRATED FLOW, NOT LIVE DATA</span>
          </div>
        </div>
      </section>
      <div className="principles wrap">
        <span>
          Less infrastructure.
          <br />
          <strong>More understanding.</strong>
        </span>
        <div>
          <span className="tiny-icon" aria-hidden="true">
            ⌘
          </span>{' '}
          Standard-library core
        </div>
        <div>
          <span className="tiny-icon" aria-hidden="true">
            ⇥
          </span>{' '}
          Bounded by design
        </div>
        <div>
          <span className="tiny-icon" aria-hidden="true">
            ↳
          </span>{' '}
          Inspectable end to end
        </div>
      </div>
    </>
  )
}
