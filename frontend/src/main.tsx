import { StrictMode, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ToastContainer } from 'react-toastify'
import { App } from '@/App'
import { applyTheme, getInitialTheme, type Theme } from '@/lib/theme'
import 'react-toastify/dist/ReactToastify.css'
import '@/index.css'

function Root() {
  const [theme, setTheme] = useState<Theme>(() => getInitialTheme())
  function toggleTheme() {
    const next = theme === 'dark' ? 'light' : 'dark'
    applyTheme(next)
    setTheme(next)
  }
  applyTheme(theme)
  return <BrowserRouter>
    <App theme={theme} onThemeToggle={toggleTheme} />
    <ToastContainer position="top-right" autoClose={4000} newestOnTop closeOnClick pauseOnFocusLoss draggable pauseOnHover theme={theme} limit={5} />
  </BrowserRouter>
}

createRoot(document.getElementById('root')!).render(<StrictMode><Root /></StrictMode>)
