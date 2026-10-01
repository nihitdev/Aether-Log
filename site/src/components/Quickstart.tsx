import { quickstart } from '../content'
import { CodeBlock } from './CodeBlock'
export function Quickstart() {
  return (
    <section className="quickstart-section" id="quickstart">
      <div className="wrap quickstart-grid">
        <div>
          <p className="eyebrow">02 / GET STARTED</p>
          <h2>
            Your first log.
            <br />
            <span className="serif">A few commands away.</span>
          </h2>
          <p>
            Build the Go binaries, start a Hub, and send a record. No service accounts. No database
            setup.
          </p>
          <div className="requirement">
            <span className="status-dot" />
            Requires Go 1.24+ and Make
          </div>
          <a className="text-link" href="?page=docs#quick-start">
            Full setup guide <span aria-hidden="true">→</span>
          </a>
        </div>
        <CodeBlock
          filename="QUICK START"
          code={quickstart}
          footer="Wait for the batch flush, then inspect aether.log."
        />
      </div>
    </section>
  )
}
