import { Link } from 'react-router-dom'

import type { Vendor } from '../../features/vendors/api'
import { formatDuration } from '../../features/vendors/format'
import { StageBadge } from '../atoms/StageBadge'
import { VendorHealthBadge } from '../atoms/VendorHealthBadge'

type VendorTableProps = {
  selectedVendorId?: string
  vendors: Vendor[]
}

export function VendorTable({ selectedVendorId, vendors }: VendorTableProps) {
  if (vendors.length === 0) {
    return (
      <div className="px-6 py-14 text-center">
        <p className="font-semibold text-slate-900">No vendors match this view</p>
        <p className="mt-1 text-sm text-slate-500">Switch to All to review the complete onboarding queue.</p>
      </div>
    )
  }

  return (
    <>
      <div className="divide-y divide-slate-100 border-t border-slate-200 md:hidden">
        {vendors.map((vendor) => {
          const isSelected = vendor.id === selectedVendorId
          return (
            <article className={isSelected ? 'bg-orange-50/70 p-5' : 'bg-white p-5'} key={vendor.id}>
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h3 className="font-bold text-slate-950">{vendor.name}</h3>
                  <p className="mt-1 text-sm text-slate-500">{vendor.region}</p>
                </div>
                <StageBadge stage={vendor.current_stage} />
              </div>
              <dl className="mt-4 grid grid-cols-2 gap-4 text-sm">
                <div>
                  <dt className="text-slate-500">Time in stage</dt>
                  <dd className="mt-1 font-mono font-semibold text-slate-800">
                    {formatDuration(vendor.hours_in_current_stage)}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">Coordinator</dt>
                  <dd className="mt-1 font-semibold text-slate-800">{vendor.assigned_coordinator.name}</dd>
                </div>
              </dl>
              <div className="mt-4 flex items-center justify-between gap-4">
                <VendorHealthBadge isActive={vendor.current_stage === 'ACTIVE'} isStuck={vendor.is_stuck} />
                <Link
                  aria-current={isSelected ? 'page' : undefined}
                  className="rounded-md px-2 py-1.5 text-sm font-semibold text-orange-700 hover:bg-orange-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange-600"
                  to={`/vendors/${vendor.id}`}
                >
                  Review
                </Link>
              </div>
              {vendor.notes ? <p className="mt-4 text-sm leading-6 text-slate-600">{vendor.notes}</p> : null}
            </article>
          )
        })}
      </div>

      <div className="hidden overflow-x-auto md:block">
        <table className="w-full min-w-[780px] border-collapse text-left">
        <caption className="sr-only">Vendors and their current onboarding status</caption>
        <thead>
          <tr className="border-y border-slate-200 bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <th className="px-5 py-3 font-semibold" scope="col">
              Vendor
            </th>
            <th className="px-4 py-3 font-semibold" scope="col">
              Region
            </th>
            <th className="px-4 py-3 font-semibold" scope="col">
              Stage
            </th>
            <th className="px-4 py-3 font-semibold" scope="col">
              Time in stage
            </th>
            <th className="px-4 py-3 font-semibold" scope="col">
              Health
            </th>
            <th className="px-4 py-3 font-semibold" scope="col">
              Coordinator
            </th>
            <th className="px-4 py-3 font-semibold" scope="col">
              Notes
            </th>
            <th className="px-5 py-3 text-right font-semibold" scope="col">
              <span className="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {vendors.map((vendor) => {
            const isSelected = vendor.id === selectedVendorId
            return (
              <tr className={isSelected ? 'bg-orange-50/70' : 'bg-white hover:bg-slate-50/80'} key={vendor.id}>
                <th className="px-5 py-4 text-sm font-semibold text-slate-950" scope="row">
                  {vendor.name}
                </th>
                <td className="px-4 py-4 text-sm text-slate-600">{vendor.region}</td>
                <td className="px-4 py-4">
                  <StageBadge stage={vendor.current_stage} />
                </td>
                <td className="px-4 py-4 font-mono text-sm text-slate-700">
                  {formatDuration(vendor.hours_in_current_stage)}
                </td>
                <td className="px-4 py-4">
                  <VendorHealthBadge isActive={vendor.current_stage === 'ACTIVE'} isStuck={vendor.is_stuck} />
                </td>
                <td className="px-4 py-4 text-sm text-slate-600">{vendor.assigned_coordinator.name}</td>
                <td className="max-w-56 px-4 py-4 text-sm text-slate-600">
                  <span className="line-clamp-2">{vendor.notes || '—'}</span>
                </td>
                <td className="px-5 py-4 text-right">
                  <Link
                    aria-current={isSelected ? 'page' : undefined}
                    className="rounded-md px-2 py-1.5 text-sm font-semibold text-orange-700 hover:bg-orange-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange-600"
                    to={`/vendors/${vendor.id}`}
                  >
                    Review
                  </Link>
                </td>
              </tr>
            )
          })}
        </tbody>
        </table>
      </div>
    </>
  )
}
