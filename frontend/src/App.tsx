import { Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from '@/components/organisms/AppLayout'
import { ConnectionSetupPage } from '@/pages/ConnectionSetupPage'
import { ConnectionsPage } from '@/pages/ConnectionsPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import { TableBrowserPage } from '@/pages/TableBrowserPage'
import type { Theme } from '@/lib/theme'

export function App({ theme, onThemeToggle }: { theme: Theme; onThemeToggle: () => void }) {
  return <Routes>
    <Route element={<AppLayout theme={theme} onThemeToggle={onThemeToggle} />}>
      <Route index element={<Navigate to="/connections" replace />} />
      <Route path="connections" element={<ConnectionsPage />} />
      <Route path="connections/new" element={<ConnectionSetupPage />} />
      <Route path="connections/:connectionId/tables" element={<TableBrowserPage />} />
      <Route path="connections/:connectionId/tables/:tableName" element={<TableBrowserPage />} />
      <Route path="*" element={<NotFoundPage />} />
    </Route>
  </Routes>
}
