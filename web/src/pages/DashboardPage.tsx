import { useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { Spinner } from '../components/atoms/Spinner'
import { InfoTooltip } from '../components/atoms/InfoTooltip'
import { ErrorAlert } from '../components/molecules/ErrorAlert'
import { type VendorFilterValue, VendorFilter } from '../components/molecules/VendorFilter'
import { AppHeader } from '../components/organisms/AppHeader'
import { VendorDetailPanel } from '../components/organisms/VendorDetailPanel'
import { VendorNotFoundPanel } from '../components/organisms/VendorNotFoundPanel'
import { VendorTable } from '../components/organisms/VendorTable'
import { AppLayout } from '../components/templates/AppLayout'
import { useCurrentCoordinator, useLogout } from '../features/auth/queries'
import { useVendors } from '../features/vendors/queries'

export function DashboardPage() {
  const navigate = useNavigate()
  const { vendorId } = useParams<{ vendorId: string }>()
  const queryClient = useQueryClient()
  const coordinator = useCurrentCoordinator()
  const vendors = useVendors()
  const logout = useLogout()
  const [filter, setFilter] = useState<VendorFilterValue>('all')
  const [logoutError, setLogoutError] = useState(false)

  const coordinatorVendors = useMemo(
    () => vendors.data?.filter((vendor) => vendor.assigned_coordinator.id === coordinator.data?.id) ?? [],
    [coordinator.data?.id, vendors.data],
  )
  const visibleVendors = filter === 'mine' ? coordinatorVendors : (vendors.data ?? [])
  const selectedVendor = vendors.data?.find((vendor) => vendor.id === vendorId)
  const stuckCount = vendors.data?.filter((vendor) => vendor.is_stuck).length ?? 0
  const activeCount = vendors.data?.filter((vendor) => vendor.current_stage === 'ACTIVE').length ?? 0

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

      <div className="mb-7 flex flex-col justify-between gap-5 lg:flex-row lg:items-end">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.18em] text-orange-600">Coordinator workspace</p>
          <h1 className="mt-2 text-3xl font-bold tracking-tight sm:text-4xl">Vendor onboarding</h1>
          <p className="mt-2 max-w-2xl leading-7 text-slate-600">
            Review progress, spot vendors waiting too long, and inspect every recorded stage change.
          </p>
        </div>
        {vendors.data ? (
          <VendorFilter
            allCount={vendors.data.length}
            mineCount={coordinatorVendors.length}
            onChange={setFilter}
            value={filter}
          />
        ) : null}
      </div>

      {vendors.data ? (
        <div className="mb-6 grid gap-3 sm:grid-cols-3">
          <SummaryCard label="Total vendors" value={vendors.data.length} />
          <SummaryCard
            helpText="A vendor needs attention when it stays in the same onboarding stage longer than the configured limit (7 days by default). Active vendors are excluded. The system calculates this automatically from when the vendor entered its current stage."
            label="Need attention"
            tone="danger"
            value={stuckCount}
          />
          <SummaryCard label="Active" tone="success" value={activeCount} />
        </div>
      ) : null}

      {vendors.isPending ? (
        <section className="rounded-2xl border border-slate-200 bg-white px-6 py-16 shadow-sm">
          <Spinner label="Loading vendors…" />
        </section>
      ) : null}
      {vendors.isError ? (
        <ErrorAlert>Unable to load vendors. Check that the API is running, then try again.</ErrorAlert>
      ) : null}
      {vendors.data ? (
        <div className={`grid items-start gap-6 ${vendorId ? 'xl:grid-cols-[minmax(0,1fr)_24rem]' : ''}`}>
          <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
            <div className="flex items-center justify-between gap-4 px-5 py-4">
              <div>
                <h2 className="font-bold text-slate-950">{filter === 'all' ? 'All vendors' : 'Assigned to me'}</h2>
                <p className="mt-0.5 text-sm text-slate-500">
                  {visibleVendors.length} {visibleVendors.length === 1 ? 'vendor' : 'vendors'} shown
                </p>
              </div>
              <p className="hidden text-xs text-slate-400 sm:block">Stuck status calculated by the server</p>
            </div>
            <VendorTable selectedVendorId={vendorId} vendors={visibleVendors} />
          </section>

          {selectedVendor ? <VendorDetailPanel vendor={selectedVendor} /> : null}
          {vendorId && !selectedVendor ? <VendorNotFoundPanel /> : null}
        </div>
      ) : null}
    </AppLayout>
  )
}

type SummaryCardProps = {
  helpText?: string
  label: string
  tone?: 'neutral' | 'danger' | 'success'
  value: number
}

function SummaryCard({ helpText, label, tone = 'neutral', value }: SummaryCardProps) {
  const valueClass = tone === 'danger' ? 'text-rose-700' : tone === 'success' ? 'text-emerald-700' : 'text-slate-950'
  return (
    <div className="rounded-xl border border-slate-200 bg-white px-5 py-4 shadow-sm">
      <div className="flex items-center gap-2">
        <p className="text-sm font-medium text-slate-500">{label}</p>
        {helpText ? <InfoTooltip label={`What does ${label} mean?`}>{helpText}</InfoTooltip> : null}
      </div>
      <p className={`mt-1 text-2xl font-bold ${valueClass}`}>{value}</p>
    </div>
  )
}
