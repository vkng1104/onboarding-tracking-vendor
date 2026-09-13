import { useId, type ReactNode } from 'react'

type InfoTooltipProps = {
  children: ReactNode
  label: string
}

export function InfoTooltip({ children, label }: InfoTooltipProps) {
  const tooltipId = useId()

  return (
    <span className="group relative inline-flex">
      <button
        aria-describedby={tooltipId}
        aria-label={label}
        className="inline-flex size-6 items-center justify-center rounded-full bg-slate-200 text-slate-700 transition hover:bg-slate-300 hover:text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-500"
        type="button"
      >
        <svg aria-hidden="true" className="size-4" fill="none" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2" />
          <path d="M12 11v6" stroke="currentColor" strokeLinecap="round" strokeWidth="2" />
          <circle cx="12" cy="7.5" fill="currentColor" r="1" />
        </svg>
      </button>
      <span
        className="pointer-events-none invisible absolute left-1/2 top-full z-30 mt-2 w-64 -translate-x-1/2 rounded-lg bg-slate-900 px-3 py-2.5 text-xs font-normal leading-5 text-white opacity-0 shadow-lg transition group-hover:visible group-hover:opacity-100 group-focus-within:visible group-focus-within:opacity-100"
        id={tooltipId}
        role="tooltip"
      >
        {children}
      </span>
    </span>
  )
}
