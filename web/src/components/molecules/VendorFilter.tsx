export type VendorFilterValue = 'all' | 'mine'

type VendorFilterProps = {
  allCount: number
  mineCount: number
  onChange: (value: VendorFilterValue) => void
  value: VendorFilterValue
}

export function VendorFilter({ allCount, mineCount, onChange, value }: VendorFilterProps) {
  return (
    <fieldset>
      <legend className="sr-only">Filter vendors</legend>
      <div className="inline-flex rounded-lg border border-slate-200 bg-slate-100 p-1">
        <button
          aria-pressed={value === 'all'}
          className={`rounded-md px-3 py-1.5 text-sm font-semibold transition ${
            value === 'all' ? 'bg-white text-slate-950 shadow-sm' : 'text-slate-600 hover:text-slate-950'
          }`}
          onClick={() => onChange('all')}
          type="button"
        >
          All <span className="text-slate-400">{allCount}</span>
        </button>
        <button
          aria-pressed={value === 'mine'}
          className={`rounded-md px-3 py-1.5 text-sm font-semibold transition ${
            value === 'mine' ? 'bg-white text-slate-950 shadow-sm' : 'text-slate-600 hover:text-slate-950'
          }`}
          onClick={() => onChange('mine')}
          type="button"
        >
          Assigned to me <span className="text-slate-400">{mineCount}</span>
        </button>
      </div>
    </fieldset>
  )
}
