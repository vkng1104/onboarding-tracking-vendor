import { Link } from 'react-router-dom'

import type { Vendor } from '../../features/vendors/api'
import { formatDuration, formatStage, formatTimestamp } from '../../features/vendors/format'
import { useVendorHistory } from '../../features/vendors/queries'
import { StageBadge } from '../atoms/StageBadge'
import { VendorHealthBadge } from '../atoms/VendorHealthBadge'
import { ErrorAlert } from '../molecules/ErrorAlert'
import { StageUpdateForm } from '../molecules/StageUpdateForm'

export function VendorDetailPanel({ vendor }: { vendor: Vendor }) {
  const history = useVendorHistory(vendor.id)

  return (
    <aside
      aria-label={`${vendor.name} details`}
      className="order-first overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm xl:order-none"
    >
      <div className="border-b border-slate-200 px-6 py-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-orange-600">Vendor review</p>
            <h2 className="mt-2 text-2xl font-bold tracking-tight text-slate-950">{vendor.name}</h2>
            <p className="mt-1 text-sm text-slate-500">{vendor.region}</p>
          </div>
          <Link
            className="inline-flex min-h-10 items-center gap-1.5 rounded-lg border border-rose-300 bg-white px-3 py-2 text-sm font-semibold text-rose-700 shadow-sm transition hover:border-rose-400 hover:bg-rose-50 hover:text-rose-800 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-rose-500"
            to="/vendors"
          >
            <svg aria-hidden="true" className="size-4" fill="none" viewBox="0 0 24 24">
              <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" strokeLinecap="round" strokeWidth="2" />
            </svg>
            Close
          </Link>
        </div>

        <div className="mt-5 flex flex-wrap items-center gap-3">
          <StageBadge stage={vendor.current_stage} />
          <VendorHealthBadge isActive={vendor.current_stage === 'ACTIVE'} isStuck={vendor.is_stuck} />
        </div>
      </div>

      <div className="space-y-6 px-6 py-5">
        <dl className="grid grid-cols-2 gap-x-5 gap-y-4 text-sm">
          <div>
            <dt className="text-slate-500">Time in stage</dt>
            <dd className="mt-1 font-semibold text-slate-900">{formatDuration(vendor.hours_in_current_stage)}</dd>
          </div>
          <div>
            <dt className="text-slate-500">Coordinator</dt>
            <dd className="mt-1 font-semibold text-slate-900">{vendor.assigned_coordinator.name}</dd>
          </div>
          <div className="col-span-2">
            <dt className="text-slate-500">Expected next stage</dt>
            <dd className="mt-1 font-semibold text-slate-900">
              {vendor.next_stage ? formatStage(vendor.next_stage) : 'Workflow complete'}
            </dd>
          </div>
          <div className="col-span-2">
            <dt className="text-slate-500">Notes</dt>
            <dd className="mt-1 leading-6 text-slate-700">{vendor.notes || 'No notes added.'}</dd>
          </div>
        </dl>

        <StageUpdateForm currentStage={vendor.current_stage} key={vendor.id} vendorId={vendor.id} />

        <section aria-labelledby="history-heading" className="border-t border-slate-200 pt-5">
          <div className="flex items-baseline justify-between gap-3">
            <h3 className="font-bold text-slate-950" id="history-heading">
              Stage history
            </h3>
            <span className="text-xs text-slate-400">Times shown in UTC</span>
          </div>

          {history.isPending ? <p className="mt-4 text-sm text-slate-500">Loading history…</p> : null}
          {history.isError ? (
            <div className="mt-4">
              <ErrorAlert>Unable to load this vendor’s history. Please try again.</ErrorAlert>
            </div>
          ) : null}
          {history.data?.length === 0 ? (
            <p className="mt-4 rounded-lg bg-slate-50 px-4 py-3 text-sm text-slate-600">No stage changes yet.</p>
          ) : null}
          {history.data && history.data.length > 0 ? (
            <ol className="mt-5 space-y-5">
              {history.data.map((event) => (
                <li className="relative border-l-2 border-slate-200 pl-5" key={event.id}>
                  <span
                    aria-hidden="true"
                    className="absolute -left-[7px] top-1 size-3 rounded-full border-2 border-white bg-orange-500"
                  />
                  <p className="text-sm font-semibold text-slate-900">
                    {formatStage(event.previous_stage)} → {formatStage(event.new_stage)}
                  </p>
                  <p className="mt-1 text-sm text-slate-600">Changed by {event.actor.name}</p>
                  <time className="mt-1 block text-xs text-slate-400" dateTime={event.occurred_at}>
                    {formatTimestamp(event.occurred_at)}
                  </time>
                </li>
              ))}
            </ol>
          ) : null}
        </section>
      </div>
    </aside>
  )
}
