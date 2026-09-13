import type { PropsWithChildren, ReactNode } from 'react'

type AppLayoutProps = PropsWithChildren<{
  header: ReactNode
}>

export function AppLayout({ children, header }: AppLayoutProps) {
  return (
    <div className="min-h-screen bg-slate-50 text-slate-950">
      {header}
      <main className="mx-auto max-w-6xl px-6 py-12">{children}</main>
    </div>
  )
}
