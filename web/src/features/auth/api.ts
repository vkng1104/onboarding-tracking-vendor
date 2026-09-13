import { request } from '../../lib/api'

export type Coordinator = {
  id: string
  name: string
  email: string
}

type AuthResponse = {
  coordinator: Coordinator
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
