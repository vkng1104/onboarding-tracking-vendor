import type { InputHTMLAttributes } from 'react'

import { Input } from '../atoms/Input'

type FormFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  id: string
  label: string
}

export function FormField({ id, label, ...inputProps }: FormFieldProps) {
  return (
    <div>
      <label className="mb-2 block text-sm font-semibold text-slate-700" htmlFor={id}>
        {label}
      </label>
      <Input id={id} {...inputProps} />
    </div>
  )
}
