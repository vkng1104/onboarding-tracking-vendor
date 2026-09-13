import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { getCurrentCoordinator, login, logout } from './api'

export const authQueryKey = ['auth', 'me'] as const

export function useCurrentCoordinator() {
  return useQuery({
    queryKey: authQueryKey,
    queryFn: getCurrentCoordinator,
    retry: false,
    staleTime: 60_000,
  })
}

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) => login(email, password),
    onSuccess: (coordinator) => queryClient.setQueryData(authQueryKey, coordinator),
  })
}

export function useLogout() {
  return useMutation({ mutationFn: logout })
}
