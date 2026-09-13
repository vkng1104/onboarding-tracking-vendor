import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { authQueryKey } from '../features/auth/queries'
import { DashboardPage } from './DashboardPage'

describe('DashboardPage', () => {
  it('clears authenticated state before navigating after logout', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    queryClient.setQueryData(authQueryKey, {
      id: '10000000-0000-0000-0000-000000000001',
      name: 'Linh Nguyen',
      email: 'linh@demo.local',
    })

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/vendors']}>
          <Routes>
            <Route path="/login" element={<h1>Logged out</h1>} />
            <Route path="/vendors" element={<DashboardPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )

    expect(screen.getByText('Signed in as')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Log out' }))

    expect(await screen.findByRole('heading', { name: 'Logged out' })).toBeInTheDocument()
    expect(queryClient.getQueryData(authQueryKey)).toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/auth/logout',
      expect.objectContaining({ credentials: 'include', method: 'POST' }),
    )
  })
})
