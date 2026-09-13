import type { PropsWithChildren } from 'react'

export function ErrorAlert({ children }: PropsWithChildren) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800" role="alert">
      {children}
    </div>
  )
}
