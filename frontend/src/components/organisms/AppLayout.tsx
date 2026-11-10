import { useState } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { Activity, Boxes, ChevronRight, Database, Menu, Moon, Plus, Sun, X } from 'lucide-react'
import type { Theme } from '@/lib/theme'

interface AppLayoutProps {
  theme: Theme
  onThemeToggle: () => void
}

export function AppLayout({ theme, onThemeToggle }: AppLayoutProps) {
  const [mobileOpen, setMobileOpen] = useState(false)
  const location = useLocation()
  const inConnection = location.pathname.startsWith('/connections/') && location.pathname !== '/connections/new'
  const onConnections = location.pathname === '/connections' || inConnection
  let breadcrumbTitle = 'Connections'
  if (inConnection) breadcrumbTitle = 'Table explorer'
  else if (location.pathname.includes('new')) breadcrumbTitle = 'New connection'

  return <div className="app-frame">
    {mobileOpen && <button type="button" className="mobile-scrim" aria-label="Close navigation" onClick={() => setMobileOpen(false)} />}
    <aside className={`sidebar ${mobileOpen ? 'sidebar-open' : ''}`}>
      <div className="brand-row"><span className="brand-mark"><Database size={19} /></span><span className="brand-copy"><span className="brand-name">dynager</span><small className="brand-subtitle">DynamoDB manager</small></span><button type="button" className="icon-button sidebar-close" aria-label="Close navigation" onClick={() => setMobileOpen(false)}><X size={18} /></button></div>
      <div className="workspace-chip"><span className="workspace-avatar">L</span><span><strong>Local workspace</strong><small>Single-user mode</small></span><ChevronRight size={15} /></div>
      <nav className="side-nav" aria-label="Main navigation">
        <p className="nav-label">WORKSPACE</p>
        <NavLink to="/connections" onClick={() => setMobileOpen(false)} className={() => `nav-item ${onConnections ? 'nav-active' : ''}`}><Boxes size={17} /><span>Connections</span><span className="nav-count"><Activity size={12} /></span></NavLink>
        <NavLink to="/connections/new" onClick={() => setMobileOpen(false)} className={({ isActive }) => `nav-item ${isActive ? 'nav-active' : ''}`}><Plus size={17} /><span>New connection</span></NavLink>
      </nav>
      <div className="sidebar-foot"><span className="secure-indicator" /><span>Credentials stay in your AWS chain</span></div>
    </aside>
    <div className="main-column">
      <header className="topbar">
        <div className="topbar-left"><button type="button" className="icon-button mobile-menu" aria-label="Open navigation" onClick={() => setMobileOpen(true)}><Menu size={20} /></button><div className="breadcrumbs"><span>Workspace</span><ChevronRight size={14} /><strong>{breadcrumbTitle}</strong></div></div>
        <div className="topbar-actions"><span className="api-indicator"><span />API workspace</span><button type="button" className="icon-button theme-toggle" aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`} onClick={onThemeToggle}>{theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}</button><span className="user-avatar" title="Local workspace">LW</span></div>
      </header>
      <main className="main-content"><Outlet /></main>
      <footer className="page-footer"><span>dynager <span className="footer-dot">·</span> DynamoDB manager</span><span>Local-first, credential-safe by design</span></footer>
    </div>
  </div>
}
