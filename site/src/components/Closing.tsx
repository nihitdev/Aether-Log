export function Closing() {
  return (
    <>
      <section className="honesty wrap">
        <div>
          <p className="eyebrow">BUILT TO BE UNDERSTOOD</p>
          <h2>
            Small scope.
            <br />
            Clear guarantees.
          </h2>
        </div>
        <div>
          <p>
            Aether-Log is a distributed logging experiment, not a managed observability platform. A
            successful TCP write is not a persistence acknowledgement. Records may be lost, and
            retries may duplicate them.
          </p>
          <p>
            There is no TLS, authentication, durable client spool, or query engine. Run on loopback
            or a trusted network, and choose bounds that fit your host.
          </p>
          <a className="text-link" href="?page=docs&doc=protocol#delivery-semantics">
            Understand delivery semantics <span aria-hidden="true">↗</span>
          </a>
        </div>
      </section>
      <section className="closing wrap">
        <span className="closing-mark" aria-hidden="true">
          ›_
        </span>
        <h2>
          Make your logs
          <br />
          <span className="serif">less scattered.</span>
        </h2>
        <a className="button primary" href="#quickstart">
          Start with Aether <span aria-hidden="true">↗</span>
        </a>
        <p>A little infrastructure. A lot less mystery.</p>
      </section>
    </>
  )
}
