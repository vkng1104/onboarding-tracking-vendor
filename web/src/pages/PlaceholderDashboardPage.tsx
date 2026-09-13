import { Badge } from '../components/atoms/Badge'
import { AppLayout } from '../components/templates/AppLayout'

export function PlaceholderDashboardPage() {
  return (
    <AppLayout>
      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-100 px-8 py-7">
          <Badge>Foundation ready</Badge>
          <h1 className="mt-4 text-3xl font-bold tracking-tight">Vendor onboarding tracker</h1>
          <p className="mt-3 max-w-2xl leading-7 text-slate-600">
            The application shell is running. Seeded coordinator login and the vendor dashboard arrive in the
            next implementation phases.
          </p>
        </div>
        <div className="grid gap-4 p-8 sm:grid-cols-3">
          {['Coordinator access', 'Vendor workflow', 'Audit history'].map((item) => (
            <div key={item} className="rounded-xl border border-dashed border-slate-300 bg-slate-50 p-5">
              <p className="font-semibold text-slate-700">{item}</p>
              <p className="mt-1 text-sm text-slate-500">Planned</p>
            </div>
          ))}
        </div>
      </section>
    </AppLayout>
  )
}
