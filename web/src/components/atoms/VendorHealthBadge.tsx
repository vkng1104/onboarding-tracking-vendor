type VendorHealthBadgeProps = {
  isActive: boolean
  isStuck: boolean
}

export function VendorHealthBadge({ isActive, isStuck }: VendorHealthBadgeProps) {
  if (isActive) {
    return (
      <span className="inline-flex items-center gap-1.5 whitespace-nowrap text-sm font-semibold text-emerald-700">
        <span aria-hidden="true" className="size-2 rounded-full bg-emerald-500" />
        Complete
      </span>
    )
  }
  if (isStuck) {
    return (
      <span className="inline-flex items-center gap-1.5 whitespace-nowrap text-sm font-semibold text-rose-700">
        <span aria-hidden="true" className="size-2 rounded-full bg-rose-500" />
        Stuck
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1.5 whitespace-nowrap text-sm font-medium text-slate-600">
      <span aria-hidden="true" className="size-2 rounded-full bg-slate-400" />
      On track
    </span>
  )
}
