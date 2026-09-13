import type { Stage } from '../../features/vendors/api'
import { formatStage } from '../../features/vendors/format'

const stageClasses: Record<Stage, string> = {
  CONTRACT_SENT: 'border-slate-200 bg-slate-100 text-slate-700',
  CONTRACT_SIGNED: 'border-blue-200 bg-blue-50 text-blue-700',
  KYC_DOCS_RECEIVED: 'border-violet-200 bg-violet-50 text-violet-700',
  KYC_VERIFIED: 'border-cyan-200 bg-cyan-50 text-cyan-700',
  ACTIVE: 'border-emerald-200 bg-emerald-50 text-emerald-700',
}

export function StageBadge({ stage }: { stage: Stage }) {
  return (
    <span
      className={`inline-flex whitespace-nowrap rounded-full border px-2.5 py-1 text-xs font-semibold ${stageClasses[stage]}`}
    >
      {formatStage(stage)}
    </span>
  )
}
