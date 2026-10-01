import { useEffect, useRef, useState } from 'react'

type Theme = 'light' | 'dark'

function savedTheme(): Theme | null {
  try {
    const value = localStorage.getItem('aether-theme')
    return value === 'light' || value === 'dark' ? value : null
  } catch {
    return null
  }
}

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(() =>
    document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light',
  )
  const explicitChoice = useRef(savedTheme() !== null)

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    document
      .querySelector('meta[name="theme-color"]')
      ?.setAttribute('content', theme === 'dark' ? '#111b16' : '#f5f3ec')
  }, [theme])

  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const change = (event: MediaQueryListEvent) => {
      if (!explicitChoice.current) setTheme(event.matches ? 'dark' : 'light')
    }
    media.addEventListener('change', change)
    return () => media.removeEventListener('change', change)
  }, [])

  function toggleTheme() {
    const next = theme === 'dark' ? 'light' : 'dark'
    explicitChoice.current = true
    setTheme(next)
    try {
      localStorage.setItem('aether-theme', next)
    } catch {
      /* Storage may be disabled. */
    }
  }

  return { theme, toggleTheme }
}
