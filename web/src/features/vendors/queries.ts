import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { ApiError } from '../../lib/api'
import { getVendorHistory, getVendors, type Stage, updateVendorStage } from './api'

export const vendorQueryKeys = {
  all: ['vendors'] as const,
  list: () => [...vendorQueryKeys.all, 'list'] as const,
  history: (vendorId: string) => [...vendorQueryKeys.all, 'history', vendorId] as const,
}

export function useVendors() {
  return useQuery({ queryKey: vendorQueryKeys.list(), queryFn: getVendors })
}

export function useVendorHistory(vendorId: string) {
  return useQuery({
    queryKey: vendorQueryKeys.history(vendorId),
    queryFn: () => getVendorHistory(vendorId),
  })
}

export function useUpdateVendorStage(vendorId: string) {
  const queryClient = useQueryClient()

  async function refreshVendorData() {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: vendorQueryKeys.list() }),
      queryClient.invalidateQueries({ queryKey: vendorQueryKeys.history(vendorId) }),
    ])
  }

  return useMutation({
    mutationFn: ({ expectedCurrentStage, newStage }: { expectedCurrentStage: Stage; newStage: Stage }) =>
      updateVendorStage(vendorId, expectedCurrentStage, newStage),
    onSuccess: refreshVendorData,
    onError: async (error) => {
      if (error instanceof ApiError && error.code === 'STAGE_CONFLICT') {
        await refreshVendorData()
      }
    },
  })
}
