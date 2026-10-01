import { useEffect, useRef, useState } from 'react'

export function CodeBlock({
  code,
  filename,
  footer,
}: {
  code: string
  filename: string
  footer?: string
}) {
  const [status, setStatus] = useState('')
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)
  useEffect(
    () => () => {
      if (timer.current) clearTimeout(timer.current)
    },
    [],
  )

  async function copy() {
    try {
      await navigator.clipboard.writeText(code)
      setStatus('Copied')
    } catch {
      setStatus('Select the code to copy')
    }
    if (timer.current) clearTimeout(timer.current)
    timer.current = setTimeout(() => setStatus(''), 2500)
  }

  return (
    <div className="code-window">
      <div className="terminal-bar">
        <span>{filename}</span>
        <button className="copy-button" onClick={copy} aria-label={`Copy ${filename}`}>
          Copy <span aria-hidden="true">⧉</span>
        </button>
      </div>
      <pre tabIndex={0}>
        <code>
          {code.split('\n').map((line, i) => (
            <span
              className={/^\s*(#|\/\/)/.test(line) ? 'code-comment code-line' : 'code-line'}
              key={i}
            >
              {line || '\u00a0'}
              {'\n'}
            </span>
          ))}
        </code>
      </pre>
      {footer ? (
        <div className="code-footer">
          <span className="status-dot" />
          {footer}
        </div>
      ) : null}
      <span className="copy-feedback" role="status" aria-live="polite">
        {status}
      </span>
    </div>
  )
}
