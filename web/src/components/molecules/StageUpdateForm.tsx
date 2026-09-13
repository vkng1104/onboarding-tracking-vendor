import { type FormEvent, useEffect, useState } from 'react'

import { stages, type Stage } from '../../features/vendors/api'
import { formatStage } from '../../features/vendors/format'
import { useUpdateVendorStage } from '../../features/vendors/queries'
import { ApiError } from '../../lib/api'
import { Button } from '../atoms/Button'
import { Select } from '../atoms/Select'
import { ErrorAlert } from './ErrorAlert'

type StageUpdateFormProps = {
  currentStage: Stage
  vendorId: string
}

export function StageUpdateForm({ currentStage, vendorId }: StageUpdateFormProps) {
  const [selectedStage, setSelectedStage] = useState(currentStage)
  const update = useUpdateVendorStage(vendorId)

  useEffect(() => {
    setSelectedStage(currentStage)
  }, [currentStage])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      await update.mutateAsync({ expectedCurrentStage: currentStage, newStage: selectedStage })
    } catch {
      // The mutation state renders a safe, actionable error below.
    }
  }

  const errorMessage =
    update.error instanceof ApiError && update.error.code === 'STAGE_CONFLICT'
      ? 'Another coordinator changed this vendor. Review the latest stage and try again.'
      : update.error instanceof ApiError && update.error.code === 'STAGE_UNCHANGED'
        ? 'Choose a stage different from the current stage.'
        : 'Unable to update the stage. Please try again.'

  return (
    <section aria-labelledby="update-stage-heading" className="border-t border-slate-200 pt-5">
      <h3 className="font-bold text-slate-950" id="update-stage-heading">
        Update stage
      </h3>
      <p className="mt-1 text-sm leading-6 text-slate-500">
        Select any different workflow stage. Previous stages are available for corrections.
      </p>

      <form className="mt-4 space-y-3" onSubmit={handleSubmit}>
        <div>
          <label className="mb-2 block text-sm font-semibold text-slate-700" htmlFor={`stage-${vendorId}`}>
            Move to stage
          </label>
          <Select
            disabled={update.isPending}
            id={`stage-${vendorId}`}
            onChange={(event) => {
              update.reset()
              setSelectedStage(event.target.value as Stage)
            }}
            value={selectedStage}
          >
            {stages.map((stage) => (
              <option key={stage} value={stage}>
                {formatStage(stage)}{stage === currentStage ? ' (current)' : ''}
              </option>
            ))}
          </Select>
        </div>

        <Button className="w-full" disabled={update.isPending || selectedStage === currentStage} type="submit">
          {update.isPending ? 'Updating…' : 'Update stage'}
        </Button>

        {update.isError ? <ErrorAlert>{errorMessage}</ErrorAlert> : null}
        {update.isSuccess ? (
          <p className="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800" role="status">
            Stage updated to {formatStage(update.data.new_stage)}.
          </p>
        ) : null}
      </form>
    </section>
  )
}
