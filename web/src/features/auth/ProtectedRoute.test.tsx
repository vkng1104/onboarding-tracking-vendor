import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { ProtectedRoute } from './ProtectedRoute'

describe('ProtectedRoute', () => {
  it('redirects an unauthenticated visitor to login', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ code: 'UNAUTHENTICATED', message: 'Authentication is required.' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/vendors']}>
          <Routes>
            <Route path="/login" element={<h1>Login page</h1>} />
            <Route element={<ProtectedRoute />}>
              <Route path="/vendors" element={<h1>Protected dashboard</h1>} />
            </Route>
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )

    expect(await screen.findByRole('heading', { name: 'Login page' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Protected dashboard' })).not.toBeInTheDocument()
  })
})
