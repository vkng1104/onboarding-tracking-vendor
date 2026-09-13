import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

import { PlaceholderDashboardPage } from '../pages/PlaceholderDashboardPage'

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/vendors" element={<PlaceholderDashboardPage />} />
        <Route path="*" element={<Navigate to="/vendors" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
