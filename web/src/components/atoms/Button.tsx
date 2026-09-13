import type { ButtonHTMLAttributes } from 'react'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary'
}

export function Button({ className = '', variant = 'primary', ...props }: ButtonProps) {
  const variantClasses =
    variant === 'primary'
      ? 'bg-orange-600 text-white hover:bg-orange-700 focus-visible:outline-orange-600'
      : 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 focus-visible:outline-slate-500'

  return (
    <button
      className={`inline-flex min-h-10 items-center justify-center rounded-lg px-4 py-2 text-sm font-semibold transition focus-visible:outline-2 focus-visible:outline-offset-2 disabled:cursor-not-allowed disabled:opacity-60 ${variantClasses} ${className}`}
      {...props}
    />
  )
}
