import { Button } from '../atoms/Button'

type AppHeaderProps = {
  coordinatorName: string
  isLoggingOut: boolean
  onLogout: () => void
}

export function AppHeader({ coordinatorName, isLoggingOut, onLogout }: AppHeaderProps) {
  return (
    <header className="border-b border-slate-200 bg-white">
      <div className="mx-auto flex max-w-[90rem] items-center justify-between gap-4 px-4 py-4 sm:px-6">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.18em] text-orange-600">Operations</p>
          <p className="mt-1 text-lg font-semibold">Vendor onboarding</p>
        </div>
        <div className="flex items-center gap-4">
          <p className="hidden text-sm text-slate-600 sm:block">
            Signed in as <span className="font-semibold text-slate-900">{coordinatorName}</span>
          </p>
          <Button disabled={isLoggingOut} onClick={onLogout} type="button" variant="secondary">
            {isLoggingOut ? 'Logging out…' : 'Log out'}
          </Button>
        </div>
      </div>
    </header>
  )
}
