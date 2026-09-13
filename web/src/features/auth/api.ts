export type Coordinator = {
  id: string
  name: string
  email: string
}

type AuthResponse = {
  coordinator: Coordinator
}

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

async function request<T>(path: string, init?: RequestInit): Promise<T> {
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

export async function getCurrentCoordinator(): Promise<Coordinator> {
  const response = await request<AuthResponse>('/api/v1/auth/me')
  return response.coordinator
}

export async function login(email: string, password: string): Promise<Coordinator> {
  const response = await request<AuthResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  return response.coordinator
}

export async function logout(): Promise<void> {
  await request<void>('/api/v1/auth/logout', { method: 'POST' })
}
