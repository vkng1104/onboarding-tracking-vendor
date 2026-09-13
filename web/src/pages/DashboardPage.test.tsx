import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { authQueryKey } from '../features/auth/queries'
import type { Vendor } from '../features/vendors/api'
import { vendorQueryKeys } from '../features/vendors/queries'
import { DashboardPage } from './DashboardPage'

const vendors: Vendor[] = [
  {
    id: 'company-a',
    name: 'Company A',
    region: 'HCMC',
    notes: '',
    current_stage: 'ACTIVE',
    stage_entered_at: '2026-08-09T12:00:00Z',
    hours_in_current_stage: 840,
    is_stuck: false,
    assigned_coordinator: { id: 'coordinator-1', name: 'Linh Nguyen' },
    next_stage: null,
  },
  {
    id: 'company-b',
    name: 'Company B',
    region: 'Can Tho',
    notes: 'Waiting on business license re-upload',
    current_stage: 'KYC_DOCS_RECEIVED',
    stage_entered_at: '2026-09-05T12:00:00Z',
    hours_in_current_stage: 192,
    is_stuck: true,
    assigned_coordinator: { id: 'coordinator-2', name: 'Huy Tran' },
    next_stage: 'KYC_VERIFIED',
  },
  {
    id: 'company-c',
    name: 'Company C',
    region: 'Hanoi',
    notes: '',
    current_stage: 'CONTRACT_SIGNED',
    stage_entered_at: '2026-09-10T12:00:00Z',
    hours_in_current_stage: 72,
    is_stuck: false,
    assigned_coordinator: { id: 'coordinator-1', name: 'Linh Nguyen' },
    next_stage: 'KYC_DOCS_RECEIVED',
  },
]

describe('DashboardPage', () => {
  it('renders vendor status and filters the list to the signed-in coordinator', () => {
    renderDashboard('/vendors')

    const table = screen.getByRole('table', { name: 'Vendors and their current onboarding status' })
    const activeRow = within(table).getByRole('row', { name: /Company A/ })
    expect(within(activeRow).getByText('Active')).toBeInTheDocument()
    expect(within(activeRow).getByText('Complete')).toBeInTheDocument()
    expect(within(activeRow).queryByText('Stuck')).not.toBeInTheDocument()

    const stuckRow = within(table).getByRole('row', { name: /Company B/ })
    expect(within(stuckRow).getByText('KYC Docs Received')).toBeInTheDocument()
    expect(within(stuckRow).getByText('8d 0h')).toBeInTheDocument()
    expect(within(stuckRow).getByText('Stuck')).toBeInTheDocument()
    expect(within(stuckRow).getByText('Huy Tran')).toBeInTheDocument()
    expect(within(stuckRow).getByText('Waiting on business license re-upload')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Assigned to me/ }))
    expect(within(table).queryByRole('row', { name: /Company B/ })).not.toBeInTheDocument()
    expect(within(table).getByRole('row', { name: /Company C/ })).toBeInTheDocument()
  })

  it('loads attributable history when a direct vendor URL is opened', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          history: [
            {
              id: 'event-1',
              occurred_at: '2026-09-05T12:00:00Z',
              actor: { id: 'coordinator-1', name: 'Linh Nguyen' },
              previous_stage: 'CONTRACT_SIGNED',
              new_stage: 'KYC_DOCS_RECEIVED',
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    vi.stubGlobal('fetch', fetchMock)

    renderDashboard('/vendors/company-b')

    expect(await screen.findByRole('complementary', { name: 'Company B details' })).toBeInTheDocument()
    expect(await screen.findByText('Contract Signed → KYC Docs Received')).toBeInTheDocument()
    expect(screen.getByText('Changed by Linh Nguyen')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/vendors/company-b/history',
      expect.objectContaining({ credentials: 'include' }),
    )
  })

  it('renders a recoverable state for an unknown vendor URL', () => {
    renderDashboard('/vendors/missing')

    expect(screen.getByRole('heading', { name: 'This vendor is not in the current list.' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Back to all vendors' })).toHaveAttribute('href', '/vendors')
  })

  it('clears authenticated state before navigating after logout', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)
    const queryClient = renderDashboard('/vendors')

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

function renderDashboard(initialPath: string): QueryClient {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Number.POSITIVE_INFINITY } },
  })
  queryClient.setQueryData(authQueryKey, {
    id: 'coordinator-1',
    name: 'Linh Nguyen',
    email: 'linh@demo.local',
  })
  queryClient.setQueryData(vendorQueryKeys.list(), vendors)

  render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/login" element={<h1>Logged out</h1>} />
          <Route path="/vendors" element={<DashboardPage />} />
          <Route path="/vendors/:vendorId" element={<DashboardPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )

  return queryClient
}
