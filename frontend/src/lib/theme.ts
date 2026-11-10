export type Theme = 'light' | 'dark'

export function getInitialTheme(): Theme {
  const saved = localStorage.getItem('keyspace-theme')
  if (saved === 'light' || saved === 'dark') return saved
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle('dark', theme === 'dark')
  localStorage.setItem('keyspace-theme', theme)
}
