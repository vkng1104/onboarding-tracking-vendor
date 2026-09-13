import { request } from '../../lib/api'

export type Stage =
  | 'CONTRACT_SENT'
  | 'CONTRACT_SIGNED'
  | 'KYC_DOCS_RECEIVED'
  | 'KYC_VERIFIED'
  | 'ACTIVE'

export type CoordinatorSummary = {
  id: string
  name: string
}

export type Vendor = {
  id: string
  name: string
  region: string
  notes: string
  current_stage: Stage
  stage_entered_at: string
  hours_in_current_stage: number
  is_stuck: boolean
  assigned_coordinator: CoordinatorSummary
  next_stage: Stage | null
}

export type HistoryEvent = {
  id: string
  occurred_at: string
  actor: CoordinatorSummary
  previous_stage: Stage
  new_stage: Stage
}

export async function getVendors(): Promise<Vendor[]> {
  const response = await request<{ vendors: Vendor[] }>('/api/v1/vendors')
  return response.vendors
}

export async function getVendorHistory(vendorId: string): Promise<HistoryEvent[]> {
  const response = await request<{ history: HistoryEvent[] }>(
    `/api/v1/vendors/${encodeURIComponent(vendorId)}/history`,
  )
  return response.history
}
