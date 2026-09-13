import { useQuery } from '@tanstack/react-query'

import { getVendorHistory, getVendors } from './api'

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
