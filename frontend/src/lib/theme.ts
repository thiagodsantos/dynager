export type Theme = 'light' | 'dark'

const themeStorageKey = 'dynager-theme'
const previousThemeStorageKey = 'dynamodb-manager-theme'
const legacyThemeStorageKey = 'keyspace-theme'

export function getInitialTheme(): Theme {
  const saved = localStorage.getItem(themeStorageKey)
    ?? localStorage.getItem(previousThemeStorageKey)
    ?? localStorage.getItem(legacyThemeStorageKey)
  if (saved === 'light' || saved === 'dark') return saved
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle('dark', theme === 'dark')
  localStorage.setItem(themeStorageKey, theme)
  localStorage.removeItem(previousThemeStorageKey)
  localStorage.removeItem(legacyThemeStorageKey)
}
