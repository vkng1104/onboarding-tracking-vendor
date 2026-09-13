import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

import { ProtectedRoute } from '../features/auth/ProtectedRoute'
import { DashboardPage } from '../pages/DashboardPage'
import { LoginPage } from '../pages/LoginPage'

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route element={<ProtectedRoute />}>
          <Route path="/vendors" element={<DashboardPage />} />
          <Route path="/vendors/:vendorId" element={<DashboardPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/vendors" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
