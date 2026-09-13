import { Navigate, Outlet } from 'react-router-dom'

import { Spinner } from '../../components/atoms/Spinner'
import { ErrorAlert } from '../../components/molecules/ErrorAlert'
import { AuthLayout } from '../../components/templates/AuthLayout'
import { ApiError } from '../../lib/api'
import { useCurrentCoordinator } from './queries'

export function ProtectedRoute() {
  const coordinator = useCurrentCoordinator()

  if (coordinator.isPending) {
    return (
      <AuthLayout>
        <Spinner label="Checking your session…" />
      </AuthLayout>
    )
  }
  if (coordinator.error instanceof ApiError && coordinator.error.status === 401) {
    return <Navigate to="/login" replace />
  }
  if (coordinator.isError) {
    return (
      <AuthLayout>
        <ErrorAlert>Unable to reach the application. Check that the API is running, then refresh.</ErrorAlert>
      </AuthLayout>
    )
  }
  return <Outlet />
}
