export enum Theme {
  Light = 'light',
  Dark = 'dark',
}

const themeStorageKey = 'dynager-theme'
const previousThemeStorageKey = 'dynamodb-manager-theme'
const legacyThemeStorageKey = 'keyspace-theme'

export function getInitialTheme(): Theme {
  const saved = localStorage.getItem(themeStorageKey)
    ?? localStorage.getItem(previousThemeStorageKey)
    ?? localStorage.getItem(legacyThemeStorageKey)
  if (saved === Theme.Light || saved === Theme.Dark) return saved
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? Theme.Dark : Theme.Light
}

export function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle('dark', theme === Theme.Dark)
  localStorage.setItem(themeStorageKey, theme)
  localStorage.removeItem(previousThemeStorageKey)
  localStorage.removeItem(legacyThemeStorageKey)
}
