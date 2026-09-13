import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { PlaceholderDashboardPage } from './PlaceholderDashboardPage'

describe('PlaceholderDashboardPage', () => {
  it('renders the Phase 1 application shell', () => {
    render(<PlaceholderDashboardPage />)

    expect(screen.getByRole('heading', { name: 'Vendor onboarding tracker' })).toBeInTheDocument()
    expect(screen.getByText('Foundation ready')).toBeInTheDocument()
  })
})
