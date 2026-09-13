import type { SelectHTMLAttributes } from 'react'

export function Select({ className = '', ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={`min-h-10 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-950 outline-none transition focus:border-orange-500 focus:ring-3 focus:ring-orange-100 disabled:bg-slate-100 ${className}`}
      {...props}
    />
  )
}
