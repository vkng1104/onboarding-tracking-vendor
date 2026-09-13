import { Link } from 'react-router-dom'

export function VendorNotFoundPanel() {
  return (
    <aside className="order-first rounded-2xl border border-slate-200 bg-white p-6 shadow-sm xl:order-none">
      <p className="text-xs font-bold uppercase tracking-[0.16em] text-rose-600">Vendor not found</p>
      <h2 className="mt-2 text-xl font-bold text-slate-950">This vendor is not in the current list.</h2>
      <p className="mt-2 text-sm leading-6 text-slate-600">
        The link may be outdated. Return to the dashboard and choose an available vendor.
      </p>
      <Link
        className="mt-5 inline-flex rounded-lg bg-orange-600 px-4 py-2 text-sm font-semibold text-white hover:bg-orange-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange-600"
        to="/vendors"
      >
        Back to all vendors
      </Link>
    </aside>
  )
}
