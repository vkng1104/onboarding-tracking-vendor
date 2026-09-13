import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { LoginPage } from './LoginPage'

describe('LoginPage', () => {
  it('logs in with seeded credentials and navigates to the dashboard', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (path.endsWith('/me')) {
        return Promise.resolve(
          new Response(JSON.stringify({ code: 'UNAUTHENTICATED', message: 'Authentication is required.' }), {
            status: 401,
            headers: { 'Content-Type': 'application/json' },
          }),
        )
      }
      if (path.endsWith('/login')) {
        expect(init?.body).toBe(JSON.stringify({ email: 'linh@demo.local', password: 'demo1234' }))
        return Promise.resolve(
          new Response(
            JSON.stringify({
              coordinator: {
                id: '10000000-0000-0000-0000-000000000001',
                name: 'Linh Nguyen',
                email: 'linh@demo.local',
              },
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          ),
        )
      }
      return Promise.reject(new Error(`unexpected request: ${path}`))
    })
    vi.stubGlobal('fetch', fetchMock)
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/vendors" element={<h1>Authenticated dashboard</h1>} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )

    await screen.findByRole('heading', { name: 'Sign in' })
    fireEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    expect(await screen.findByRole('heading', { name: 'Authenticated dashboard' })).toBeInTheDocument()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2))
  })
})
