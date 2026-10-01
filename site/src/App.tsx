import { lazy, Suspense, useState } from 'react'
import { Hero } from './components/Hero'
import { Architecture } from './components/Architecture'
import { Features } from './components/Features'
import { Quickstart } from './components/Quickstart'
import { Clients } from './components/Clients'
import { Closing } from './components/Closing'
import { repository } from './content'
import { useTheme } from './useTheme'
const Docs = lazy(() => import('./components/Docs'))

function Brand() {
  return (
    <a className="brand" href="./" aria-label="Aether-Log home">
      <span className="brand-mark" aria-hidden="true">
        ›_
      </span>
      Aether<span className="brand-light">-Log</span>
    </a>
  )
}

export default function App() {
  const { theme, toggleTheme } = useTheme()
  const [menuOpen, setMenuOpen] = useState(false)
  const isDocs = new URLSearchParams(window.location.search).get('page') === 'docs'
  return (
    <>
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <header className="header wrap">
        <Brand />
        <div className="header-actions">
          <button
            className="theme-toggle"
            onClick={toggleTheme}
            aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
            title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              {theme === 'dark' ? (
                <>
                  <circle cx="12" cy="12" r="4" />
                  <path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5" />
                </>
              ) : (
                <path d="M20.8 13.1A9 9 0 0 1 10.9 3.2 9 9 0 1 0 20.8 13.1Z" />
              )}
            </svg>
          </button>
          <button
            className="menu-toggle"
            aria-expanded={menuOpen}
            aria-controls="navigation"
            onClick={() => setMenuOpen((open) => !open)}
          >
            Menu <span aria-hidden="true">{menuOpen ? '−' : '+'}</span>
          </button>
        </div>
        <nav id="navigation" className={menuOpen ? 'open' : ''} aria-label="Main navigation">
          <a onClick={() => setMenuOpen(false)} href={isDocs ? './#architecture' : '#architecture'}>
            Architecture
          </a>
          <a onClick={() => setMenuOpen(false)} href={isDocs ? './#clients' : '#clients'}>
            Clients
          </a>
          <a
            onClick={() => setMenuOpen(false)}
            href="?page=docs"
            aria-current={isDocs ? 'page' : undefined}
          >
            Documentation
          </a>
          <a className="repo-link" href={repository} onClick={() => setMenuOpen(false)}>
            View on GitHub <span aria-hidden="true">↗</span>
          </a>
        </nav>
      </header>
      {isDocs ? (
        <Suspense
          fallback={
            <main id="main" className="wrap docs-loading">
              Opening the docs…
            </main>
          }
        >
          <Docs />
        </Suspense>
      ) : (
        <main id="main">
          <Hero />
          <Architecture />
          <Features />
          <Quickstart />
          <Clients />
          <Closing />
        </main>
      )}
      <footer className="footer wrap">
        <Brand />
        <span>A small distributed logging experiment.</span>
        <div>
          <a href="?page=docs">Docs</a>
          <a href={repository}>GitHub ↗</a>
        </div>
      </footer>
    </>
  )
}
