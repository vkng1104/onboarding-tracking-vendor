import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { Badge } from '../components/atoms/Badge'
import { Spinner } from '../components/atoms/Spinner'
import { ErrorAlert } from '../components/molecules/ErrorAlert'
import { AppHeader } from '../components/organisms/AppHeader'
import { AppLayout } from '../components/templates/AppLayout'
import { useCurrentCoordinator, useLogout } from '../features/auth/queries'

export function DashboardPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const coordinator = useCurrentCoordinator()
  const logout = useLogout()
  const [logoutError, setLogoutError] = useState(false)

  if (!coordinator.data) {
    return <Spinner label="Loading dashboard…" />
  }

  async function handleLogout() {
    setLogoutError(false)
    try {
      await logout.mutateAsync()
      await queryClient.cancelQueries()
      queryClient.clear()
      navigate('/login', { replace: true })
    } catch {
      setLogoutError(true)
    }
  }

  return (
    <AppLayout
      header={
        <AppHeader
          coordinatorName={coordinator.data.name}
          isLoggingOut={logout.isPending}
          onLogout={handleLogout}
        />
      }
    >
      {logoutError ? (
        <div className="mb-6">
          <ErrorAlert>Logout failed. Please try again.</ErrorAlert>
        </div>
      ) : null}
      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-100 px-8 py-7">
          <Badge>Coordinator workspace</Badge>
          <h1 className="mt-4 text-3xl font-bold tracking-tight">Welcome back, {coordinator.data.name}</h1>
          <p className="mt-3 max-w-2xl leading-7 text-slate-600">
            Authentication is ready. Seeded vendor tracking, stuck-state highlighting, and history arrive in the
            next phase.
          </p>
        </div>
        <div className="grid gap-4 p-8 sm:grid-cols-3">
          {['Vendor dashboard', 'Stuck-state tracking', 'Transition history'].map((item) => (
            <div key={item} className="rounded-xl border border-dashed border-slate-300 bg-slate-50 p-5">
              <p className="font-semibold text-slate-700">{item}</p>
              <p className="mt-1 text-sm text-slate-500">Planned for Phase 3</p>
            </div>
          ))}
        </div>
      </section>
    </AppLayout>
  )
}
