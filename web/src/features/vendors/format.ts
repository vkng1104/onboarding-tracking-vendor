import type { Stage } from './api'

const stageLabels: Record<Stage, string> = {
  CONTRACT_SENT: 'Contract Sent',
  CONTRACT_SIGNED: 'Contract Signed',
  KYC_DOCS_RECEIVED: 'KYC Docs Received',
  KYC_VERIFIED: 'KYC Verified',
  ACTIVE: 'Active',
}

export function formatStage(stage: Stage): string {
  return stageLabels[stage]
}

export function formatDuration(hours: number): string {
  const safeHours = Math.max(0, Math.floor(hours))
  return `${Math.floor(safeHours / 24)}d ${safeHours % 24}h`
}

export function formatTimestamp(timestamp: string): string {
  return new Intl.DateTimeFormat('en-GB', {
    dateStyle: 'medium',
    timeStyle: 'short',
    timeZone: 'UTC',
  }).format(new Date(timestamp))
}
