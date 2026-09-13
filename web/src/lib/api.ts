type ErrorResponse = {
  code?: string
  message?: string
}

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message)
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    credentials: 'include',
    ...init,
    headers: {
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...init?.headers,
    },
  })

  if (!response.ok) {
    const error = (await response.json().catch(() => ({}))) as ErrorResponse
    throw new ApiError(response.status, error.code ?? 'UNKNOWN_ERROR', error.message ?? 'Request failed.')
  }

  if (response.status === 204) {
    return undefined as T
  }
  return (await response.json()) as T
}
