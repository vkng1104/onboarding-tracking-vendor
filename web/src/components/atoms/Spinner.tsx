type SpinnerProps = {
  label?: string
}

export function Spinner({ label = 'Loading' }: SpinnerProps) {
  return (
    <div className="flex items-center gap-3 text-sm text-slate-600" role="status">
      <span className="size-5 animate-spin rounded-full border-2 border-slate-200 border-t-orange-600" aria-hidden />
      <span>{label}</span>
    </div>
  )
}
