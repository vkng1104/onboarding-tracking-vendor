import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
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

    const attentionInfo = screen.getByRole('button', { name: 'What does Need attention mean?' })
    const attentionTooltip = screen.getByRole('tooltip')
    expect(attentionInfo).toHaveAttribute('aria-describedby', attentionTooltip.id)
    expect(attentionTooltip).toHaveTextContent(
      'A vendor needs attention when it stays in the same onboarding stage longer than the configured limit (7 days by default). Active vendors are excluded.',
    )

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
    expect(screen.getByRole('link', { name: 'Close' })).toHaveClass('border-rose-300', 'text-rose-700')
    expect(await screen.findByText('Contract Signed → KYC Docs Received')).toBeInTheDocument()
    expect(screen.getByText('Changed by Linh Nguyen')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/vendors/company-b/history',
      expect.objectContaining({ credentials: 'include' }),
    )
  })

  it('moves a vendor backward and refreshes its current state and history', async () => {
    let stageUpdated = false
    const transition = {
      id: 'event-2',
      occurred_at: '2026-09-13T08:30:00Z',
      actor: { id: 'coordinator-1', name: 'Linh Nguyen' },
      previous_stage: 'KYC_DOCS_RECEIVED' as const,
      new_stage: 'CONTRACT_SIGNED' as const,
    }
    const updatedVendors = vendors.map((vendor) =>
      vendor.id === 'company-b'
        ? {
            ...vendor,
            current_stage: transition.new_stage,
            stage_entered_at: transition.occurred_at,
            hours_in_current_stage: 0,
            is_stuck: false,
            next_stage: 'KYC_DOCS_RECEIVED' as const,
          }
        : vendor,
    )
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (path === '/api/v1/vendors/company-b/stage' && init?.method === 'PATCH') {
        stageUpdated = true
        return new Response(JSON.stringify({ transition }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      if (path === '/api/v1/vendors') {
        return new Response(JSON.stringify({ vendors: updatedVendors }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      if (path === '/api/v1/vendors/company-b/history') {
        return new Response(JSON.stringify({ history: stageUpdated ? [transition] : [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      throw new Error(`Unexpected request: ${path}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    renderDashboard('/vendors/company-b')

    expect(await screen.findByRole('complementary', { name: 'Company B details' })).toBeInTheDocument()
    fireEvent.change(screen.getByLabelText('Move to stage'), { target: { value: 'CONTRACT_SIGNED' } })
    fireEvent.click(screen.getByRole('button', { name: 'Update stage' }))

    expect(await screen.findByRole('status')).toHaveTextContent('Stage updated to Contract Signed.')
    expect(await screen.findByText('KYC Docs Received → Contract Signed')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/vendors/company-b/stage',
      expect.objectContaining({
        body: JSON.stringify({
          expected_current_stage: 'KYC_DOCS_RECEIVED',
          new_stage: 'CONTRACT_SIGNED',
        }),
        credentials: 'include',
        method: 'PATCH',
      }),
    )
  })

  it('refreshes the vendor instead of overwriting a concurrent stage change', async () => {
    let conflictReturned = false
    const concurrentTransition = {
      id: 'event-concurrent',
      occurred_at: '2026-09-13T08:25:00Z',
      actor: { id: 'coordinator-2', name: 'Huy Tran' },
      previous_stage: 'KYC_DOCS_RECEIVED' as const,
      new_stage: 'KYC_VERIFIED' as const,
    }
    const concurrentlyUpdatedVendors = vendors.map((vendor) =>
      vendor.id === 'company-b'
        ? {
            ...vendor,
            current_stage: concurrentTransition.new_stage,
            stage_entered_at: concurrentTransition.occurred_at,
            hours_in_current_stage: 0,
            is_stuck: false,
            next_stage: 'ACTIVE' as const,
          }
        : vendor,
    )
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (path === '/api/v1/vendors/company-b/stage' && init?.method === 'PATCH') {
        conflictReturned = true
        return new Response(
          JSON.stringify({ code: 'STAGE_CONFLICT', message: 'Vendor stage changed since it was loaded.' }),
          { status: 409, headers: { 'Content-Type': 'application/json' } },
        )
      }
      if (path === '/api/v1/vendors') {
        return new Response(JSON.stringify({ vendors: concurrentlyUpdatedVendors }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      if (path === '/api/v1/vendors/company-b/history') {
        return new Response(JSON.stringify({ history: conflictReturned ? [concurrentTransition] : [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      throw new Error(`Unexpected request: ${path}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    renderDashboard('/vendors/company-b')

    expect(await screen.findByRole('complementary', { name: 'Company B details' })).toBeInTheDocument()
    fireEvent.change(screen.getByLabelText('Move to stage'), { target: { value: 'ACTIVE' } })
    fireEvent.click(screen.getByRole('button', { name: 'Update stage' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Another coordinator changed this vendor. Review the latest stage and try again.',
    )
    await waitFor(() => expect(screen.getByLabelText('Move to stage')).toHaveValue('KYC_VERIFIED'))
    expect(await screen.findByText('KYC Docs Received → KYC Verified')).toBeInTheDocument()
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
