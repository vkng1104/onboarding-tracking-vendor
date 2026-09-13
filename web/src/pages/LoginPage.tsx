import { type FormEvent, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'

import { Button } from '../components/atoms/Button'
import { Spinner } from '../components/atoms/Spinner'
import { ErrorAlert } from '../components/molecules/ErrorAlert'
import { FormField } from '../components/molecules/FormField'
import { AuthLayout } from '../components/templates/AuthLayout'
import { ApiError } from '../features/auth/api'
import { useCurrentCoordinator, useLogin } from '../features/auth/queries'

export function LoginPage() {
  const navigate = useNavigate()
  const currentCoordinator = useCurrentCoordinator()
  const login = useLogin()
  const [email, setEmail] = useState('linh@demo.local')
  const [password, setPassword] = useState('demo1234')
  const loginErrorMessage =
    login.error instanceof ApiError && login.error.code === 'INVALID_CREDENTIALS'
      ? 'Email or password is incorrect.'
      : 'Unable to sign in right now. Please try again.'

  if (currentCoordinator.isPending) {
    return (
      <AuthLayout>
        <Spinner label="Checking your session…" />
      </AuthLayout>
    )
  }
  if (currentCoordinator.isSuccess) {
    return <Navigate to="/vendors" replace />
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      await login.mutateAsync({ email, password })
      navigate('/vendors', { replace: true })
    } catch {
      // The mutation state renders the safe API error below.
    }
  }

  return (
    <AuthLayout>
      <div className="mb-7 text-center">
        <p className="text-xs font-bold uppercase tracking-[0.18em] text-orange-600">Vendor operations</p>
        <h1 className="mt-3 text-3xl font-bold tracking-tight">Sign in</h1>
        <p className="mt-2 text-sm leading-6 text-slate-600">Use a seeded coordinator account to continue.</p>
      </div>

      <section className="rounded-2xl border border-slate-200 bg-white p-7 shadow-sm">
        <form className="space-y-5" onSubmit={handleSubmit}>
          <FormField
            autoComplete="email"
            id="email"
            label="Email"
            onChange={(event) => setEmail(event.target.value)}
            required
            type="email"
            value={email}
          />
          <FormField
            autoComplete="current-password"
            id="password"
            label="Password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />

          {login.isError ? <ErrorAlert>{loginErrorMessage}</ErrorAlert> : null}

          <Button className="w-full" disabled={login.isPending} type="submit">
            {login.isPending ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>
      </section>

      <aside className="mt-5 rounded-xl border border-slate-200 bg-white/70 px-5 py-4 text-sm text-slate-600">
        <p className="font-semibold text-slate-800">Demo accounts</p>
        <p className="mt-1">linh@demo.local · huy@demo.local · mai@demo.local</p>
        <p className="mt-1">Shared password: demo1234</p>
      </aside>
    </AuthLayout>
  )
}
