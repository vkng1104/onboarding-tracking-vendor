import type { InputHTMLAttributes } from 'react'

export function Input({ className = '', ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={`min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3.5 py-2 text-sm text-slate-950 outline-none transition placeholder:text-slate-400 focus:border-orange-500 focus:ring-3 focus:ring-orange-100 disabled:bg-slate-100 ${className}`}
      {...props}
    />
  )
}
